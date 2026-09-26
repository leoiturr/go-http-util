package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DiffRequest binds two text blocks for comparison
type DiffRequest struct {
	TextA string `form:"text_a" json:"text_a"`
	TextB string `form:"text_b" json:"text_b"`
}

// DiffLine represents a single line in the diff output
type DiffLine struct {
	Type    string `json:"type"` // "same", "add", "del"
	Content string `json:"content"`
	LineA   int    `json:"line_a,omitempty"`
	LineB   int    `json:"line_b,omitempty"`
}

// DiffResult contains the computed diff between two text blocks
type DiffResult struct {
	Lines        []DiffLine `json:"lines"`
	LinesAdded   int        `json:"lines_added"`
	LinesRemoved int        `json:"lines_removed"`
}

// ComputeDiff calculates a line-by-line diff using LCS
func ComputeDiff(textA, textB string) DiffResult {
	linesA := strings.Split(textA, "\n")
	linesB := strings.Split(textB, "\n")

	// Trim trailing empty line from split if text doesn't end with newline
	if textA != "" && strings.HasSuffix(textA, "\n") {
		linesA = linesA[:len(linesA)-1]
	}
	if textB != "" && strings.HasSuffix(textB, "\n") {
		linesB = linesB[:len(linesB)-1]
	}

	// LCS dynamic programming table
	m, n := len(linesA), len(linesB)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if linesA[i-1] == linesB[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to build diff
	var diffLines []DiffLine
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && linesA[i-1] == linesB[j-1] {
			diffLines = append(diffLines, DiffLine{
				Type:    "same",
				Content: linesA[i-1],
				LineA:   i,
				LineB:   j,
			})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			diffLines = append(diffLines, DiffLine{
				Type:    "add",
				Content: linesB[j-1],
				LineB:   j,
			})
			j--
		} else if i > 0 {
			diffLines = append(diffLines, DiffLine{
				Type:    "del",
				Content: linesA[i-1],
				LineA:   i,
			})
			i--
		}
	}

	// Reverse to correct order
	for left, right := 0, len(diffLines)-1; left < right; left, right = left+1, right-1 {
		diffLines[left], diffLines[right] = diffLines[right], diffLines[left]
	}

	added, removed := 0, 0
	for _, l := range diffLines {
		if l.Type == "add" {
			added++
		} else if l.Type == "del" {
			removed++
		}
	}

	return DiffResult{
		Lines:        diffLines,
		LinesAdded:   added,
		LinesRemoved: removed,
	}
}

// HTMXTextDiff handles diff requests from HTMX
func HTMXTextDiff(c *gin.Context) {
	var req DiffRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "diff-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	if strings.TrimSpace(req.TextA) == "" && strings.TrimSpace(req.TextB) == "" {
		c.HTML(http.StatusOK, "diff-result", gin.H{
			"Error": "Both text fields are empty.",
		})
		return
	}

	result := ComputeDiff(req.TextA, req.TextB)

	c.HTML(http.StatusOK, "diff-result", gin.H{
		"Diff":         result,
		"LinesAdded":   result.LinesAdded,
		"LinesRemoved": result.LinesRemoved,
		"OriginalA":    req.TextA,
		"OriginalB":    req.TextB,
	})
}

// V1TextDiff handles JSON REST API requests for text diff
func V1TextDiff(c *gin.Context) {
	var req DiffRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	if strings.TrimSpace(req.TextA) == "" && strings.TrimSpace(req.TextB) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both text fields are empty"})
		return
	}

	result := ComputeDiff(req.TextA, req.TextB)

	c.JSON(http.StatusOK, gin.H{
		"lines":         result.Lines,
		"lines_added":   result.LinesAdded,
		"lines_removed": result.LinesRemoved,
		"original_a":    req.TextA,
		"original_b":    req.TextB,
	})
}

// FormatDiffAsText renders a DiffResult as unified diff text
func FormatDiffAsText(result DiffResult) string {
	var sb strings.Builder
	for _, line := range result.Lines {
		switch line.Type {
		case "add":
			sb.WriteString("+ " + line.Content + "\n")
		case "del":
			sb.WriteString("- " + line.Content + "\n")
		default:
			sb.WriteString("  " + line.Content + "\n")
		}
	}
	return sb.String()
}
