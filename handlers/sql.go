package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	sqlformatter "github.com/BruceDu521/sql-formatter"
	"github.com/gin-gonic/gin"
)

// SQLRequest represents the common payload for SQL tools
type SQLRequest struct {
	Text string `form:"text" json:"text" binding:"required"`
}

// HTMXPrettifySQL handles formatting SQL via HTMX
func HTMXPrettifySQL(c *gin.Context) {
	var req SQLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusOK, "sql-result", gin.H{
			"Operation": "Prettify",
			"Error":     "Input text is required",
		})
		return
	}

	formatted, err := formatSQL(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "sql-result", gin.H{
			"Operation": "Prettify",
			"Error":     fmt.Sprintf("Failed to format SQL: %v", err),
		})
		return
	}

	c.HTML(http.StatusOK, "sql-result", gin.H{
		"Operation": "Prettify",
		"Result":    formatted,
		"ResultID":  "sql-prettify-result-text",
	})
}

// HTMXMinifySQL handles minifying SQL via HTMX
func HTMXMinifySQL(c *gin.Context) {
	var req SQLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusOK, "sql-result", gin.H{
			"Operation": "Minify",
			"Error":     "Input text is required",
		})
		return
	}

	minified := minifySQL(req.Text)

	c.HTML(http.StatusOK, "sql-result", gin.H{
		"Operation": "Minify",
		"Result":    minified,
		"ResultID":  "sql-minify-result-text",
	})
}

// V1PrettifySQL provides a REST API to format SQL
func V1PrettifySQL(c *gin.Context) {
	var req SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload or missing 'text'"})
		return
	}

	formatted, err := formatSQL(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to format SQL: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": formatted})
}

// V1MinifySQL provides a REST API to minify SQL
func V1MinifySQL(c *gin.Context) {
	var req SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload or missing 'text'"})
		return
	}

	minified := minifySQL(req.Text)
	c.JSON(http.StatusOK, gin.H{"result": minified})
}

// formatSQL prettifies SQL. It extends the third-party formatter with support
// for WITH (common table expression) statements.
func formatSQL(sql string) (string, error) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return "", fmt.Errorf("SQL statement cannot be empty")
	}

	if isWithStatement(sql) {
		return formatWithStatement(sql)
	}

	f := sqlformatter.NewFormatter()
	return f.Format(sql)
}

// isWithStatement reports whether sql starts with a WITH clause.
func isWithStatement(sql string) bool {
	if len(sql) < 4 {
		return false
	}
	prefix := strings.ToUpper(sql[:4])
	if prefix != "WITH" {
		return false
	}
	if len(sql) == 4 {
		return true
	}
	next := sql[4]
	return next == ' ' || next == '\t' || next == '\n' || next == '\r'
}

// formatWithStatement formats a SQL statement that begins with WITH.
// It extracts each CTE, formats its inner query, and then formats the
// main SELECT/INSERT/UPDATE/DELETE statement.
func formatWithStatement(sql string) (string, error) {
	mainStart := findMainQueryStart(sql)
	if mainStart == -1 {
		return "", fmt.Errorf("failed to locate main query in WITH statement")
	}

	withClause := strings.TrimSpace(sql[:mainStart])
	mainQuery := strings.TrimSpace(sql[mainStart:])

	f := sqlformatter.NewFormatter()
	formattedMain, err := f.Format(mainQuery)
	if err != nil {
		return "", err
	}

	// Strip the leading "WITH" keyword.
	withBody := strings.TrimSpace(withClause)
	if len(withBody) > 4 {
		withBody = strings.TrimSpace(withBody[4:])
	}

	// Detect a possible "RECURSIVE" modifier.
	recursive := false
	if len(withBody) >= 9 && strings.EqualFold(withBody[:9], "RECURSIVE") {
		recursive = true
		withBody = strings.TrimSpace(withBody[9:])
	}

	ctes := splitTopLevel(withBody, ',')
	if len(ctes) == 0 {
		return "", fmt.Errorf("invalid WITH clause")
	}

	var result strings.Builder
	indent := strings.Repeat(" ", f.IndentSize)
	result.WriteString(strings.ToUpper("WITH"))
	if recursive {
		result.WriteString(" " + strings.ToUpper("RECURSIVE"))
	}
	for i, cte := range ctes {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString("\n")
		result.WriteString(indent)
		formattedCTE, err := formatCTE(strings.TrimSpace(cte))
		if err != nil {
			return "", err
		}
		result.WriteString(formattedCTE)
	}
	result.WriteString("\n")
	result.WriteString(formattedMain)
	return result.String(), nil
}

// formatCTE formats a single CTE definition such as
// "cte_name AS (SELECT id FROM users)".
func formatCTE(cte string) (string, error) {
	upper := strings.ToUpper(cte)
	asIdx := strings.Index(upper, " AS (")
	if asIdx == -1 {
		return "", fmt.Errorf("invalid CTE definition: %q", cte)
	}

	namePart := strings.TrimSpace(cte[:asIdx])
	// " AS (" is 5 characters long, so the subquery starts at asIdx+5.
	subquery := strings.TrimSpace(cte[asIdx+5:])
	if strings.HasSuffix(subquery, ")") {
		subquery = strings.TrimSpace(subquery[:len(subquery)-1])
	}

	f := sqlformatter.NewFormatter()
	formattedSub, err := f.Format(subquery)
	if err != nil {
		return "", err
	}

	outerIndent := strings.Repeat(" ", f.IndentSize)
	innerIndent := strings.Repeat(" ", 2*f.IndentSize)
	indentedSub := indentBlock(formattedSub, innerIndent)

	var result strings.Builder
	result.WriteString(namePart)
	result.WriteString(" ")
	result.WriteString(strings.ToUpper("AS"))
	result.WriteString(" (\n")
	result.WriteString(indentedSub)
	result.WriteString("\n")
	result.WriteString(outerIndent)
	result.WriteString(")")
	return result.String(), nil
}

// findMainQueryStart returns the index of the top-level SELECT, INSERT,
// UPDATE or DELETE keyword that follows the WITH clause.
func findMainQueryStart(sql string) int {
	upper := strings.ToUpper(sql)
	keywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	depth := 0

	for i := 0; i < len(sql); i++ {
		switch sql[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if depth != 0 {
				continue
			}
			for _, kw := range keywords {
				end := i + len(kw)
				if end > len(sql) || upper[i:end] != kw {
					continue
				}
				if end < len(sql) && isIdentifierChar(sql[end]) {
					continue
				}
				return i
			}
		}
	}
	return -1
}

// isIdentifierChar reports whether c is an ASCII identifier character.
func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// splitTopLevel splits s by sep at parenthesis depth zero.
func splitTopLevel(s string, sep byte) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '(':
			depth++
			current.WriteByte(c)
		case ')':
			if depth > 0 {
				depth--
			}
			current.WriteByte(c)
		case sep:
			if depth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteByte(c)
			}
		default:
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// indentBlock prefixes each non-empty line of block with indent.
func indentBlock(block, indent string) string {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// minifySQL is a simple helper to compress SQL by stripping comments and extra whitespace
func minifySQL(sql string) string {
	sql = stripComments(sql)

	// Remove newlines and tabs
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\r", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")

	// Replace multiple spaces with a single space
	re := regexp.MustCompile(`\s+`)
	sql = re.ReplaceAllString(sql, " ")

	// Optionally trim spaces around brackets/parentheses if needed, but simple space compression is safe
	return strings.TrimSpace(sql)
}

// stripComments removes both line comments (--) and block comments (/* */) from SQL,
// while leaving string literals untouched.
func stripComments(sql string) string {
	var b strings.Builder
	b.Grow(len(sql))

	i := 0
	n := len(sql)
	for i < n {
		c := sql[i]

		// Preserve string literals so comment markers inside them are not stripped
		if c == '\'' || c == '"' || c == '`' {
			quote := c
			b.WriteByte(c)
			i++
			for i < n {
				if sql[i] == '\\' && quote != '`' && i+1 < n {
					b.WriteByte(sql[i])
					b.WriteByte(sql[i+1])
					i += 2
					continue
				}
				b.WriteByte(sql[i])
				if sql[i] == quote {
					// Handle escaped quotes by doubling (standard SQL)
					if i+1 < n && sql[i+1] == quote {
						b.WriteByte(sql[i+1])
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			continue
		}

		// Line comment: -- until end of line
		if c == '-' && i+1 < n && sql[i+1] == '-' {
			i += 2
			for i < n && sql[i] != '\n' && sql[i] != '\r' {
				i++
			}
			continue
		}

		// Block comment: /* ... */
		if c == '/' && i+1 < n && sql[i+1] == '*' {
			i += 2
			for i+1 < n && !(sql[i] == '*' && sql[i+1] == '/') {
				i++
			}
			if i+1 < n {
				i += 2
			} else {
				i = n
			}
			continue
		}

		b.WriteByte(c)
		i++
	}

	return b.String()
}
