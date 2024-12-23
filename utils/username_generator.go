package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]`)

// GenerateUsername creates a username from display name or email and adds random numbers if needed
func GenerateUsername(displayName string, email string) string {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Try to use display name first, if empty use email
	baseName := displayName
	if baseName == "" {
		// Use part before @ in email
		baseName = strings.Split(email, "@")[0]
	}

	// Clean the base name
	baseName = strings.ToLower(baseName)
	baseName = nonAlphanumeric.ReplaceAllString(baseName, "")

	// If base name is too short, use a default
	if len(baseName) < 3 {
		baseName = "user"
	}

	// Truncate if too long
	if len(baseName) > 15 {
		baseName = baseName[:15]
	}

	// Add random numbers
	randomNum := rand.Intn(9999)
	return fmt.Sprintf("%s%04d", baseName, randomNum)
}
