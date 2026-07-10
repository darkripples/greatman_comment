package character

import "strings"

// SanitizeReply removes visible boundary/citation blocks that the model may still append.
func SanitizeReply(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return content
	}
	markers := []string{
		"\n\n---\n\n**附：边界说明**",
		"\n\n---\n\n附：边界说明",
		"\n---\n\n**附：边界说明**",
		"\n---\n\n附：边界说明",
		"\n\n**附：边界说明**",
		"\n\n附：边界说明",
		"\n**附：边界说明**",
		"\n附：边界说明",
	}
	for _, m := range markers {
		if i := strings.Index(content, m); i >= 0 {
			content = strings.TrimSpace(content[:i])
		}
	}
	// Trailing horizontal rule before accidental appendix
	if i := strings.LastIndex(content, "\n---"); i >= 0 && i > len(content)/2 {
		tail := strings.TrimSpace(content[i:])
		if strings.Contains(tail, "边界说明") || strings.Contains(tail, "【依据】") {
			content = strings.TrimSpace(content[:i])
		}
	}
	return strings.TrimSpace(content)
}
