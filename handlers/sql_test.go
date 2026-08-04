package handlers

import (
	"testing"
)

func TestMinifySQLStripsComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "line comment",
			input: `SELECT 1 -- this is a comment
FROM users`,
			expected: "SELECT 1 FROM users",
		},
		{
			name: "block comment",
			input: `SELECT /* inline */ 1
FROM users /* trailing */`,
			expected: "SELECT 1 FROM users",
		},
		{
			name: "multi-line block comment",
			input: `SELECT 1
/*
 * multi
 * line
 */
FROM users`,
			expected: "SELECT 1 FROM users",
		},
		{
			name:     "comment inside string literal preserved",
			input:    `SELECT '-- not a comment', '/* not a comment */'`,
			expected: "SELECT '-- not a comment', '/* not a comment */'",
		},
		{
			name:     "escaped quote inside string",
			input:    `SELECT 'it''s -- still a string'`,
			expected: "SELECT 'it''s -- still a string'",
		},
		{
			name: "only comments",
			input: `-- leading comment
/* block */`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minifySQL(tt.input)
			if got != tt.expected {
				t.Errorf("minifySQL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
