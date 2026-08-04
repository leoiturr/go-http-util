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

	f := sqlformatter.NewFormatter()
	formatted, err := f.Format(req.Text)
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

	f := sqlformatter.NewFormatter()
	formatted, err := f.Format(req.Text)
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
