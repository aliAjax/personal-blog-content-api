package slug

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"
)

func Make(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var builder strings.Builder
	lastDash := false

	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if r <= unicode.MaxASCII {
				builder.WriteRune(r)
			} else {
				builder.WriteString(fmt.Sprintf("u%x", r))
			}
			lastDash = false
		case r == ' ' || r == '-' || r == '_':
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		result = hash(value)
	}
	return result
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "item-" + fmt.Sprintf("%x", sum)[:12]
}
