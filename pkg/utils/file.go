package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxPhotoSize = 5 * 1024 * 1024 // 5 MB

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

var allowedImageMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func ValidatePhotoFile(src io.Reader, filename string, size int64) (io.Reader, error) {
	if size > maxPhotoSize {
		return nil, errors.New("ukuran foto maksimal 5 MB")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedImageExts[ext] {
		return nil, errors.New("format foto tidak didukung, gunakan JPG, PNG, atau WEBP")
	}
	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return nil, errors.New("gagal membaca file")
	}
	mime := http.DetectContentType(buf[:n])
	if !allowedImageMIME[mime] {
		return nil, errors.New("file bukan gambar yang valid")
	}
	return io.MultiReader(bytes.NewReader(buf[:n]), src), nil
}

func SaveUploadedFile(file io.Reader, dstPath string) error {
	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}

func ExtractBase64Data(dataStr string) string {
	prefixes := []string{
		"data:image/jpeg;base64,",
		"data:image/png;base64,",
		"data:image/jpg;base64,",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(dataStr, prefix) {
			return strings.TrimPrefix(dataStr, prefix)
		}
	}
	return dataStr
}

func SaveBase64Image(base64Data, dstPath string) error {
	dataStr := ExtractBase64Data(base64Data)

	decoded, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return err
	}

	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(dstPath, decoded, 0644)
}

func GetPhotoExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	return ext
}

func BuildPhotoFilename(nis string, ext string) string {
	if ext == "" {
		ext = ".jpg"
	}
	return nis + "_" + time.Now().Format("20060102150405") + ext
}
