package handlers

import (
	"strings"
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

func TestFormatSQLWithStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "simple cte lowercase",
			input: "with cte as (select id from users) select * from cte",
			expected: strings.Join([]string{
				"WITH",
				"  cte AS (",
				"    SELECT",
				"      id",
				"    FROM",
				"      users",
				"  )",
				"SELECT",
				"  *",
				"FROM",
				"  cte",
			}, "\n"),
		},
		{
			name: "multiline cte",
			input: `WITH cte AS (
SELECT id
FROM users
) SELECT * FROM cte`,
			expected: strings.Join([]string{
				"WITH",
				"  cte AS (",
				"    SELECT",
				"      id",
				"    FROM",
				"      users",
				"  )",
				"SELECT",
				"  *",
				"FROM",
				"  cte",
			}, "\n"),
		},
		{
			name:  "multiple ctes",
			input: "WITH cte1 AS (SELECT id FROM t1), cte2 AS (SELECT id FROM t2) SELECT * FROM cte1 JOIN cte2 ON cte1.id = cte2.id",
			expected: strings.Join([]string{
				"WITH",
				"  cte1 AS (",
				"    SELECT",
				"      id",
				"    FROM",
				"      t1",
				"  ),",
				"  cte2 AS (",
				"    SELECT",
				"      id",
				"    FROM",
				"      t2",
				"  )",
				"SELECT",
				"  *",
				"FROM",
				"  cte1",
				"  JOIN cte2 ON cte1.id = cte2.id",
			}, "\n"),
		},
		{
			name:  "recursive cte",
			input: "WITH RECURSIVE cte AS (SELECT id FROM users) SELECT * FROM cte",
			expected: strings.Join([]string{
				"WITH RECURSIVE",
				"  cte AS (",
				"    SELECT",
				"      id",
				"    FROM",
				"      users",
				"  )",
				"SELECT",
				"  *",
				"FROM",
				"  cte",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatSQL(tt.input)
			if err != nil {
				t.Fatalf("formatSQL(%q) returned error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("formatSQL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
