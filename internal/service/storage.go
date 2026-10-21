package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageService struct {
	db              *sql.DB
	accountID       string
	accessKeyID     string
	secretAccessKey string
	bucketName      string
	publicURL       string
	s3Client        *s3.Client
}

// NewStorageService initializes the storage service, pulling configuration from env or attendance_settings
func NewStorageService(db *sql.DB) *StorageService {
	s := &StorageService{db: db}
	s.loadConfig()
	return s
}

func (s *StorageService) loadConfig() {
	s.accountID = os.Getenv("R2_ACCOUNT_ID")
	s.accessKeyID = os.Getenv("R2_ACCESS_KEY_ID")
	s.secretAccessKey = os.Getenv("R2_SECRET_ACCESS_KEY")
	s.bucketName = os.Getenv("R2_BUCKET_NAME")
	s.publicURL = strings.TrimSuffix(os.Getenv("R2_PUBLIC_URL"), "/")

	// Fallback to database attendance_settings or school_settings if env variables are empty
	if s.db != nil {
		if s.accountID == "" {
			_ = s.db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = 'r2_account_id'").Scan(&s.accountID)
			if s.accountID == "" {
				_ = s.db.QueryRow("SELECT setting_value FROM school_settings WHERE setting_key = 'r2_account_id'").Scan(&s.accountID)
			}
		}
		if s.accessKeyID == "" {
			_ = s.db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = 'r2_access_key_id'").Scan(&s.accessKeyID)
			if s.accessKeyID == "" {
				_ = s.db.QueryRow("SELECT setting_value FROM school_settings WHERE setting_key = 'r2_access_key_id'").Scan(&s.accessKeyID)
			}
		}
		if s.secretAccessKey == "" {
			_ = s.db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = 'r2_secret_access_key'").Scan(&s.secretAccessKey)
			if s.secretAccessKey == "" {
				_ = s.db.QueryRow("SELECT setting_value FROM school_settings WHERE setting_key = 'r2_secret_access_key'").Scan(&s.secretAccessKey)
			}
		}
		if s.bucketName == "" {
			_ = s.db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = 'r2_bucket_name'").Scan(&s.bucketName)
			if s.bucketName == "" {
				_ = s.db.QueryRow("SELECT setting_value FROM school_settings WHERE setting_key = 'r2_bucket_name'").Scan(&s.bucketName)
			}
		}
		if s.publicURL == "" {
			var pUrl string
			if err := s.db.QueryRow("SELECT setting_value FROM attendance_settings WHERE setting_key = 'r2_public_url'").Scan(&pUrl); err == nil && pUrl != "" {
				s.publicURL = strings.TrimSuffix(pUrl, "/")
			} else if err := s.db.QueryRow("SELECT setting_value FROM school_settings WHERE setting_key = 'r2_public_url'").Scan(&pUrl); err == nil && pUrl != "" {
				s.publicURL = strings.TrimSuffix(pUrl, "/")
			}
		}
	}

	// Initialize S3 Client if credentials are provided
	if s.IsR2Configured() {
		endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", s.accountID)
		cfg := aws.Config{
			Region: "auto",
			Credentials: credentials.NewStaticCredentialsProvider(
				s.accessKeyID,
				s.secretAccessKey,
				"",
			),
		}
		s.s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}
}

// IsR2Configured checks whether R2 credentials are validly supplied
func (s *StorageService) IsR2Configured() bool {
	return s.accountID != "" && s.accessKeyID != "" && s.secretAccessKey != "" && s.bucketName != ""
}

// UploadFile uploads file to Cloudflare R2 if configured, or gracefully saves to local disk
func (s *StorageService) UploadFile(ctx context.Context, fileKey string, reader io.Reader, contentType string) (string, error) {
	// Re-check config in case it was updated dynamically via settings
	s.loadConfig()

	if s.s3Client != nil && s.IsR2Configured() {
		// Upload to Cloudflare R2
		_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.bucketName),
			Key:         aws.String(fileKey),
			Body:        reader,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return "", fmt.Errorf("gagal upload ke Cloudflare R2: %w", err)
		}

		if s.publicURL != "" {
			return fmt.Sprintf("%s/%s", s.publicURL, fileKey), nil
		}
		return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", s.accountID, s.bucketName, fileKey), nil
	}

	// Fallback to local storage: ./public/uploads/lms/<fileKey>
	localBasePath := filepath.Join(".", "public", "uploads", "lms")
	destPath := filepath.Join(localBasePath, fileKey)

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori lokal: %w", err)
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat berkas lokal: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, reader); err != nil {
		return "", fmt.Errorf("gagal menulis data berkas lokal: %w", err)
	}

	return fmt.Sprintf("/assets/uploads/lms/%s", fileKey), nil
}

// DeleteFile deletes a file from Cloudflare R2 or local disk
func (s *StorageService) DeleteFile(ctx context.Context, fileKey string) error {
	if s.s3Client != nil && s.IsR2Configured() {
		_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(fileKey),
		})
		return err
	}

	localPath := filepath.Join(".", "public", "uploads", "lms", fileKey)
	return os.Remove(localPath)
}

