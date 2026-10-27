package handler

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// maxLMSFileSize caps LMS material and assignment uploads at 50 MB.
const maxLMSFileSize = 50 * 1024 * 1024

// allowedLMSDocumentExts is the extension allowlist for uploaded documents.
// HTML and SVG are deliberately excluded: they execute script when served from
// the application origin.
var allowedLMSDocumentExts = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".txt":  true,
	".csv":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".zip":  true,
	".mp3":  true,
	".mp4":  true,
}

// allowedLMSContentTypes maps each allowed extension to the content type the
// server will declare. Client-supplied types are never trusted.
var allowedLMSContentTypes = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".txt":  "text/plain; charset=utf-8",
	".csv":  "text/csv; charset=utf-8",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".zip":  "application/zip",
	".mp3":  "audio/mpeg",
	".mp4":  "video/mp4",
}

// isAllowedLMSContentType checks a sniffed MIME type against the allowlist.
func isAllowedLMSContentType(detected string) bool {
	for _, allowed := range allowedLMSContentTypes {
		base := allowed
		if i := strings.IndexByte(allowed, ';'); i >= 0 {
			base = allowed[:i]
		}
		if detected == base || strings.HasPrefix(detected, base+";") {
			return true
		}
	}
	// ZIP files are sometimes sniffed as application/octet-stream.
	return detected == "application/zip" || detected == "application/octet-stream"
}

// unsafeFilenameChars matches everything that is not safe in a storage key.
var unsafeFilenameChars = regexp.MustCompile(`[^A-Za-z0-9._-]`)

// sanitizeFilenameSegment reduces a user-supplied filename to a conservative
// token: no path separators, no traversal sequences, no control characters,
// and a bounded length.
func sanitizeFilenameSegment(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	if name == "." || name == ".." {
		return ""
	}
	name = unsafeFilenameChars.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	if len(name) > 80 {
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		if len(ext) > 10 {
			ext = ""
		}
		if len(base) > 80-len(ext) {
			base = base[:80-len(ext)]
		}
		name = base + ext
	}
	return name
}

// validateExternalURL rejects URLs that could execute script when rendered as
// a link. Only http and https are accepted, which blocks javascript:, data:,
// and vbscript: payloads.
func validateExternalURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("URL tidak boleh kosong")
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("Format URL tidak valid")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("Hanya tautan http dan https yang diizinkan")
	}
	if u.Hostname() == "" {
		return fmt.Errorf("URL tidak memiliki host yang valid")
	}
	return nil
}
