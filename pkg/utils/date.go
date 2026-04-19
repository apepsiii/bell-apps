package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func DateToIndo(t time.Time) string {
	days := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	day := days[t.Weekday()]
	month := months[t.Month()]
	return fmt.Sprintf("%s, %d %s %d", day, t.Day(), month, t.Year())
}

func FormatPhone(phone string) string {
	reg := regexp.MustCompile(`[^0-9]`)
	clean := reg.ReplaceAllString(phone, "")

	if strings.HasPrefix(clean, "08") {
		return "62" + clean[1:]
	}
	if strings.HasPrefix(clean, "8") {
		return "62" + clean
	}
	return clean
}
