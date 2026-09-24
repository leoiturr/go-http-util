package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegexTestLogicMatchesAndCaptures(t *testing.T) {
	result, err := RegexTestLogic(
		`(?P<local>[a-z]+)@(?P<domain>[a-z.]+)`,
		"Contact ada@example.com or grace@example.org.",
		[]string{"i"},
	)
	if err != nil {
		t.Fatalf("RegexTestLogic returned error: %v", err)
	}

	if result.MatchCount != 2 {
		t.Fatalf("MatchCount = %d, want 2", result.MatchCount)
	}
	if result.GroupCount != 2 {
		t.Fatalf("GroupCount = %d, want 2", result.GroupCount)
	}
	if result.CaptureCount != 4 {
		t.Fatalf("CaptureCount = %d, want 4", result.CaptureCount)
	}
	if result.Matches[0].Value != "ada@example.com" {
		t.Errorf("first match = %q, want %q", result.Matches[0].Value, "ada@example.com")
	}
	if result.Matches[0].Captures[0].Name != "local" {
		t.Errorf("first capture name = %q, want local", result.Matches[0].Captures[0].Name)
	}
	if len(result.Flags) != 1 || result.Flags[0] != "i" {
		t.Errorf("Flags = %v, want [i]", result.Flags)
	}
	if len(result.Segments) != 4 {
		t.Fatalf("segment count = %d, want 4", len(result.Segments))
	}
	if result.Segments[0].Value != "Contact " || result.Segments[0].Matched {
		t.Errorf("first segment = %#v, want unmatched Contact text", result.Segments[0])
	}
	if result.Segments[1].Value != "ada@example.com" || !result.Segments[1].Matched {
		t.Errorf("second segment = %#v, want matched email", result.Segments[1])
	}
}

func TestRegexTestLogicSupportsMultilineAndDotAllFlags(t *testing.T) {
	result, err := RegexTestLogic(`^line:.*end$`, "first\nline:value end", []string{"m", "s"})
	if err != nil {
		t.Fatalf("RegexTestLogic returned error: %v", err)
	}
	if result.MatchCount != 1 {
		t.Fatalf("MatchCount = %d, want 1", result.MatchCount)
	}
	if result.Matches[0].Value != "line:value end" {
		t.Errorf("match = %q, want %q", result.Matches[0].Value, "line:value end")
	}
}

func TestRegexTestLogicAllowsWhitespacePattern(t *testing.T) {
	result, err := RegexTestLogic(" ", "a b", nil)
	if err != nil {
		t.Fatalf("RegexTestLogic returned error: %v", err)
	}
	if result.MatchCount != 1 || result.Matches[0].Value != " " {
		t.Fatalf("whitespace pattern result = %#v, want one space match", result.Matches)
	}
}

func TestRegexTestLogicAllowsEmptyTestText(t *testing.T) {
	result, err := RegexTestLogic("^$", "", nil)
	if err != nil {
		t.Fatalf("RegexTestLogic returned error: %v", err)
	}
	if result.MatchCount != 1 {
		t.Fatalf("MatchCount = %d, want 1", result.MatchCount)
	}
}

func TestRegexTestLogicRejectsInvalidPattern(t *testing.T) {
	_, err := RegexTestLogic("(", "text", nil)
	if err == nil || !strings.Contains(err.Error(), "invalid regular expression") {
		t.Fatalf("expected invalid regular expression error, got %v", err)
	}
}

func TestHTMXTestRegexRendersResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	})
	router.LoadHTMLGlob("../templates/*")
	router.POST("/api/htmx/regex/test", HTMXTestRegex)

	form := url.Values{}
	form.Set("pattern", "[a-z]+")
	form.Set("test_text", "ada")
	form.Add("flags", "i")
	request := httptest.NewRequest(http.MethodPost, "/api/htmx/regex/test", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Regex Match Report") || !strings.Contains(body, "regex-summary") {
		t.Fatalf("rendered result is incomplete: %s", body)
	}
}

func TestRegexTestLogicRejectsUnsupportedFlag(t *testing.T) {
	_, err := RegexTestLogic("abc", "abc", []string{"x"})
	if err == nil || !strings.Contains(err.Error(), "unsupported regex flag") {
		t.Fatalf("expected unsupported flag error, got %v", err)
	}
}

func TestRegexTestLogicBoundsResults(t *testing.T) {
	result, err := RegexTestLogic("a", strings.Repeat("a", 101), nil)
	if err != nil {
		t.Fatalf("RegexTestLogic returned error: %v", err)
	}
	if result.MatchCount != maxRegexMatches+1 {
		t.Fatalf("MatchCount = %d, want %d", result.MatchCount, maxRegexMatches+1)
	}
	if !result.Truncated {
		t.Error("Truncated = false, want true")
	}
	if result.MatchLimit != maxRegexMatches {
		t.Errorf("MatchLimit = %d, want %d", result.MatchLimit, maxRegexMatches)
	}
}

func TestRegexTestLogicRejectsOversizedInput(t *testing.T) {
	_, err := RegexTestLogic("abc", strings.Repeat("x", maxRegexTextLength+1), nil)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected test text limit error, got %v", err)
	}
}
