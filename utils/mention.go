package utils

import (
	"regexp"
	"strings"
)

// ExtractMentions extracts all @mentions from text content
// Returns a slice of usernames without the @ symbol
func ExtractMentions(content string) []string {
	// Regular expression to match @username
	// Username can contain letters, numbers, underscores, and dots
	// Must start with a letter and be at least 3 characters long
	re := regexp.MustCompile(`@([a-zA-Z][a-zA-Z0-9_\.]{2,})`)
	
	// Find all matches
	matches := re.FindAllStringSubmatch(content, -1)
	
	// Extract usernames (without @) from matches
	usernames := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			username := strings.TrimSpace(match[1])
			if username != "" {
				usernames = append(usernames, username)
			}
		}
	}
	
	return usernames
}
