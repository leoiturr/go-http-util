package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	maxRegexPatternLength = 4 * 1024
	maxRegexTextLength    = 200 * 1024
	maxRegexMatches       = 100
)

// RegexTestRequest binds a pattern, sample text, and optional RE2 flags.
type RegexTestRequest struct {
	Pattern  string   `form:"pattern" json:"pattern"`
	TestText string   `form:"test_text" json:"test_text"`
	Flags    []string `form:"flags" json:"flags"`
}

// RegexCapture describes one capture group within a match.
type RegexCapture struct {
	Index   int    `json:"index"`
	Name    string `json:"name,omitempty"`
	Value   string `json:"value,omitempty"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Matched bool   `json:"matched"`
}

// RegexMatch describes a complete match and its capture groups.
type RegexMatch struct {
	Index    int            `json:"index"`
	Value    string         `json:"value"`
	Start    int            `json:"start"`
	End      int            `json:"end"`
	Captures []RegexCapture `json:"captures"`
}

// RegexTextSegment separates highlighted matches from surrounding test text.
type RegexTextSegment struct {
	Value      string `json:"value"`
	Matched    bool   `json:"matched"`
	MatchIndex int    `json:"match_index,omitempty"`
}

// RegexTestResult contains a bounded, safe RE2 match report.
type RegexTestResult struct {
	Pattern      string             `json:"pattern"`
	Flags        []string           `json:"flags"`
	Segments     []RegexTextSegment `json:"segments"`
	Matches      []RegexMatch       `json:"matches"`
	MatchCount   int                `json:"match_count"`
	GroupCount   int                `json:"group_count"`
	CaptureCount int                `json:"capture_count"`
	MatchLimit   int                `json:"match_limit"`
	Truncated    bool               `json:"truncated"`
	Engine       string             `json:"engine"`
}

var regexFlagOrder = []string{"i", "m", "s", "U"}

func normalizeRegexFlags(flags []string) ([]string, string, error) {
	selected := make(map[string]bool)
	for _, rawFlag := range flags {
		for _, flag := range strings.Split(strings.TrimSpace(rawFlag), ",") {
			flag = strings.TrimSpace(flag)
			if flag == "" {
				continue
			}
			if flag == "u" {
				flag = "U"
			}
			switch flag {
			case "i", "m", "s", "U":
				selected[flag] = true
			default:
				return nil, "", fmt.Errorf("unsupported regex flag %q; supported flags are i, m, s, and U", flag)
			}
		}
	}

	normalized := make([]string, 0, len(regexFlagOrder))
	prefix := ""
	for _, flag := range regexFlagOrder {
		if selected[flag] {
			normalized = append(normalized, flag)
			prefix += flag
		}
	}
	return normalized, prefix, nil
}

// RegexTestLogic compiles a bounded RE2 expression and returns at most 100 matches.
func RegexTestLogic(pattern, testText string, flags []string) (RegexTestResult, error) {
	if pattern == "" {
		return RegexTestResult{}, fmt.Errorf("regex pattern is required")
	}
	if len(pattern) > maxRegexPatternLength {
		return RegexTestResult{}, fmt.Errorf("regex pattern exceeds the %d KB limit", maxRegexPatternLength/1024)
	}
	if len(testText) > maxRegexTextLength {
		return RegexTestResult{}, fmt.Errorf("test text exceeds the %d KB limit", maxRegexTextLength/1024)
	}

	normalizedFlags, flagPrefix, err := normalizeRegexFlags(flags)
	if err != nil {
		return RegexTestResult{}, err
	}

	compiledPattern := pattern
	if flagPrefix != "" {
		compiledPattern = "(?" + flagPrefix + ")" + pattern
	}
	expression, err := regexp.Compile(compiledPattern)
	if err != nil {
		return RegexTestResult{}, fmt.Errorf("invalid regular expression: %w", err)
	}

	matchIndexes := expression.FindAllStringSubmatchIndex(testText, maxRegexMatches+1)
	matchCount := len(expression.FindAllStringIndex(testText, -1))
	truncated := matchCount > maxRegexMatches
	if len(matchIndexes) > maxRegexMatches {
		matchIndexes = matchIndexes[:maxRegexMatches]
	}

	result := RegexTestResult{
		Pattern:    pattern,
		Flags:      normalizedFlags,
		Matches:    make([]RegexMatch, 0, len(matchIndexes)),
		MatchCount: matchCount,
		GroupCount: expression.NumSubexp(),
		MatchLimit: maxRegexMatches,
		Truncated:  truncated,
		Engine:     "Go RE2",
	}

	groupNames := expression.SubexpNames()
	for matchIndex, indexes := range matchIndexes {
		fullStart, fullEnd := indexes[0], indexes[1]
		match := RegexMatch{
			Index:    matchIndex + 1,
			Value:    testText[fullStart:fullEnd],
			Start:    fullStart,
			End:      fullEnd,
			Captures: make([]RegexCapture, 0, expression.NumSubexp()),
		}

		for groupIndex := 1; groupIndex < len(indexes)/2; groupIndex++ {
			start, end := indexes[groupIndex*2], indexes[groupIndex*2+1]
			capture := RegexCapture{
				Index:   groupIndex,
				Start:   start,
				End:     end,
				Matched: start >= 0 && end >= start,
			}
			if groupIndex < len(groupNames) {
				capture.Name = groupNames[groupIndex]
			}
			if capture.Matched {
				capture.Value = testText[start:end]
				result.CaptureCount++
			}
			match.Captures = append(match.Captures, capture)
		}

		result.Matches = append(result.Matches, match)
	}

	cursor := 0
	result.Segments = make([]RegexTextSegment, 0, len(result.Matches)*2+1)
	for _, match := range result.Matches {
		if match.Start > cursor {
			result.Segments = append(result.Segments, RegexTextSegment{Value: testText[cursor:match.Start]})
		}
		result.Segments = append(result.Segments, RegexTextSegment{
			Value:      match.Value,
			Matched:    true,
			MatchIndex: match.Index,
		})
		if match.End > cursor {
			cursor = match.End
		}
	}
	if cursor < len(testText) {
		result.Segments = append(result.Segments, RegexTextSegment{Value: testText[cursor:]})
	}

	return result, nil
}

// HTMXTestRegex handles Regex Playground requests from HTMX.
func HTMXTestRegex(c *gin.Context) {
	var req RegexTestRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "regex-result", gin.H{"Error": "Invalid request parameters"})
		return
	}

	result, err := RegexTestLogic(req.Pattern, req.TestText, req.Flags)
	if err != nil {
		c.HTML(http.StatusOK, "regex-result", gin.H{"Error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "regex-result", gin.H{"Result": result})
}

// V1TestRegex handles Regex Playground requests from the JSON REST API.
func V1TestRegex(c *gin.Context) {
	var req RegexTestRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	result, err := RegexTestLogic(req.Pattern, req.TestText, req.Flags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
