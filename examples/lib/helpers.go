// helpers.go - Common helper functions for examples
// This file is in a separate package to allow testing

package lib

import "strings"

// Truncate truncates a string to max length
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// PrintSeparatorLine creates a separator line for tables
func PrintSeparatorLine(widths []int) string {
	var result strings.Builder
	result.WriteString("|")
	for _, w := range widths {
		result.WriteString(strings.Repeat("-", w+2))
		result.WriteString("|")
	}
	return result.String()
}