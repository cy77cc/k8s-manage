package ai

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func jsonMarshal(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	default:
		return fmt.Sprintf("%v", x)
	}
}

func toBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		ok, _ := strconv.ParseBool(strings.TrimSpace(x))
		return ok
	case json.Number:
		n, _ := x.Int64()
		return n != 0
	case int:
		return x != 0
	case int64:
		return x != 0
	case int32:
		return x != 0
	case float64:
		return x != 0
	case float32:
		return x != 0
	default:
		return false
	}
}

func normalizeSessionTitle(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, trimmed)
	rs := []rune(strings.TrimSpace(trimmed))
	if len(rs) > 64 {
		rs = rs[:64]
	}
	return strings.TrimSpace(string(rs))
}

func inferSessionTitle(userInput string) string {
	raw := strings.TrimSpace(userInput)
	if raw == "" {
		return defaultAISessionTitle
	}
	compact := strings.Join(strings.Fields(raw), " ")
	compact = strings.TrimSpace(compact)
	if compact == "" {
		return defaultAISessionTitle
	}
	for _, sep := range []string{"\n", "。", "！", "？", ".", "!", "?", ";", "；", "，", ","} {
		if idx := strings.Index(compact, sep); idx > 0 {
			compact = compact[:idx]
			break
		}
	}
	rs := []rune(strings.TrimSpace(compact))
	if len(rs) > 24 {
		rs = rs[:24]
	}
	title := normalizeSessionTitle(string(rs))
	if title == "" {
		return defaultAISessionTitle
	}
	return title
}
