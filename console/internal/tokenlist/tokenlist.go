package tokenlist

import "strings"

func ParseCommaSeparated(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	return Normalize(strings.Split(value, ","))
}

func Normalize(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	tokens := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		token := strings.TrimSpace(raw)
		if token == "" {
			continue
		}
		if _, exists := seen[token]; exists {
			continue
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}
	return tokens
}
