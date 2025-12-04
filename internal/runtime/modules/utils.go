package modules

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type UtilsModule struct{}

func NewUtilsModule() *UtilsModule {
	return &UtilsModule{}
}

func (u *UtilsModule) Sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func (u *UtilsModule) Random() float64 {
	return rand.Float64()
}

func (u *UtilsModule) RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	return rand.Intn(max-min) + min
}

func (u *UtilsModule) UUID() string {
	return uuid.New().String()
}

func (u *UtilsModule) Slugify(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Replace spaces with hyphens
	text = strings.ReplaceAll(text, " ", "-")

	// Remove non-alphanumeric characters (except hyphens)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	text = re.ReplaceAllString(text, "")

	// Remove multiple consecutive hyphens
	re = regexp.MustCompile(`-+`)
	text = re.ReplaceAllString(text, "-")

	// Trim hyphens from beginning and end
	text = strings.Trim(text, "-")

	return text
}

func (u *UtilsModule) Truncate(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}

func (u *UtilsModule) Capitalize(text string) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (u *UtilsModule) RegexMatch(text, pattern string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return []string{}
	}
	return re.FindAllString(text, -1)
}

func (u *UtilsModule) RegexReplace(text, pattern, replacement string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return text
	}
	return re.ReplaceAllString(text, replacement)
}

func (u *UtilsModule) FormatDate(timestamp int64, format string) string {
	t := time.Unix(timestamp/1000, 0)

	// Convert common format tokens to Go format
	format = strings.ReplaceAll(format, "YYYY", "2006")
	format = strings.ReplaceAll(format, "MM", "01")
	format = strings.ReplaceAll(format, "DD", "02")
	format = strings.ReplaceAll(format, "HH", "15")
	format = strings.ReplaceAll(format, "mm", "04")
	format = strings.ReplaceAll(format, "ss", "05")

	return t.Format(format)
}

func (u *UtilsModule) ParseDate(text, format string) int64 {
	// Convert common format tokens to Go format
	format = strings.ReplaceAll(format, "YYYY", "2006")
	format = strings.ReplaceAll(format, "MM", "01")
	format = strings.ReplaceAll(format, "DD", "02")
	format = strings.ReplaceAll(format, "HH", "15")
	format = strings.ReplaceAll(format, "mm", "04")
	format = strings.ReplaceAll(format, "ss", "05")

	t, err := time.Parse(format, text)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

func (u *UtilsModule) Timestamp() int64 {
	return time.Now().UnixMilli()
}
