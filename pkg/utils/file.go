package utils

import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
