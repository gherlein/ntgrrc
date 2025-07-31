package lib

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "String shorter than max",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "String equal to max",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "String longer than max",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "Very short max length",
			input:    "hello",
			maxLen:   3,
			expected: "hel",
		},
		{
			name:     "Empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestPrintSeparatorLine(t *testing.T) {
	tests := []struct {
		name     string
		widths   []int
		expected string
	}{
		{
			name:     "Single column",
			widths:   []int{5},
			expected: "|-------|",
		},
		{
			name:     "Multiple columns",
			widths:   []int{3, 5, 4},
			expected: "|-----|-------|------|",
		},
		{
			name:     "Empty widths",
			widths:   []int{},
			expected: "|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrintSeparatorLine(tt.widths)
			if result != tt.expected {
				t.Errorf("PrintSeparatorLine(%v) = %q, want %q", tt.widths, result, tt.expected)
			}
		})
	}
}