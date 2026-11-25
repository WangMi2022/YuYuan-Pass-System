package ai

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

type redactionRule struct {
	pattern *regexp.Regexp
	mask    string
}

var redactionRules = []redactionRule{
	{regexp.MustCompile(`(?i)(authorization\s*[:=]\s*bearer\s+)[^\s,;]+`), `${1}[REDACTED]`},
	{regexp.MustCompile(`(?i)\b(api[_-]?key|token|secret|password)\s*[:=]\s*[^\s,;"']+`), `${1}=[REDACTED]`},
	{regexp.MustCompile(`\b1[3-9]\d{9}\b`), `[PHONE_REDACTED]`},
	{regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`), `[EMAIL_REDACTED]`},
	{regexp.MustCompile(`\b\d{17}[0-9Xx]\b`), `[ID_REDACTED]`},
	{regexp.MustCompile(`\b(?:\d[ -]?){13,18}\d\b`), `[CARD_REDACTED]`},
	{regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`), `[IP_REDACTED]`},
	{regexp.MustCompile(`(?i)\b[0-9A-HJ-NPQRTUWXY]{18}\b`), `[TAX_NO_REDACTED]`},
}

func Redact(input string) (string, int) {
	return RedactWithWords(input, nil)
}

func RedactWithWords(input string, sensitiveWords []string) (string, int) {
	count := 0
	redacted := input
	for _, rule := range redactionRules {
		matches := rule.pattern.FindAllStringIndex(redacted, -1)
		if len(matches) == 0 {
			continue
		}
		count += len(matches)
		redacted = rule.pattern.ReplaceAllString(redacted, rule.mask)
	}
	for _, word := range sensitiveWords {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		matches := strings.Count(redacted, word)
		if matches == 0 {
			continue
		}
		count += matches
		redacted = strings.ReplaceAll(redacted, word, "[BUSINESS_SENSITIVE_REDACTED]")
	}
	return redacted, count
}

func RedactPayload(payload map[string]any, sensitiveWords []string) (map[string]any, int) {
	redacted, count := redactValue(payload, sensitiveWords)
	result, ok := redacted.(map[string]any)
	if !ok {
		return payload, count
	}
	return result, count
}

func redactValue(value any, sensitiveWords []string) (any, int) {
	switch typed := value.(type) {
	case string:
		return RedactWithWords(typed, sensitiveWords)
	case map[string]any:
		result := make(map[string]any, len(typed))
		count := 0
		for key, child := range typed {
			if sensitiveKey(key) {
				if text, ok := child.(string); ok && text != "" {
					result[key] = "[REDACTED]"
					count++
					continue
				}
			}
			var childCount int
			result[key], childCount = redactValue(child, sensitiveWords)
			count += childCount
		}
		return result, count
	case []any:
		result := make([]any, len(typed))
		count := 0
		for index, child := range typed {
			var childCount int
			result[index], childCount = redactValue(child, sensitiveWords)
			count += childCount
		}
		return result, count
	default:
		return value, 0
	}
}

func sensitiveKey(key string) bool {
	key = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(key), "-", ""), "_", ""))
	for _, fragment := range []string{"authorization", "apikey", "accesstoken", "refreshtoken", "token", "secret", "password", "passwd"} {
		if strings.Contains(key, fragment) {
			return true
		}
	}
	return false
}

func Hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:])
}
