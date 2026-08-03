package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

// FlexCount accepts a count value as either a JSON string or number.
type FlexCount string

func (fc *FlexCount) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fc = FlexCount(s)
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*fc = FlexCount(strconv.Itoa(n))
		return nil
	}
	return fmt.Errorf("count must be a string or integer")
}

type CronRequest struct {
	Expression string    `form:"expression" json:"expression"`
	Count      FlexCount `form:"count" json:"count"`
	Timezone   string    `form:"timezone" json:"timezone"`
}

type CronResult struct {
	Expression     string   `json:"expression"`
	IsValid        bool     `json:"is_valid"`
	Error          string   `json:"error,omitempty"`
	Description    string   `json:"description,omitempty"`
	NextRuns       []string `json:"next_runs,omitempty"`
	Timezone       string   `json:"timezone,omitempty"`
	Fields         string   `json:"fields,omitempty"`
	StandardFormat string   `json:"standard_format,omitempty"`
}

var cronDescriptions = map[string]string{
	"@yearly":   "Run once a year at midnight on January 1st",
	"@annually": "Run once a year at midnight on January 1st",
	"@monthly":  "Run once a month at midnight on the first day",
	"@weekly":   "Run once a week at midnight on Sunday",
	"@daily":    "Run once a day at midnight",
	"@midnight": "Run once a day at midnight",
	"@hourly":   "Run once an hour at the beginning of the hour",
}

var cronAliases = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

func ParseCronLogic(expr string, countStr string, tz string) (CronResult, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return CronResult{}, fmt.Errorf("cron expression is required")
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.Local
	}

	count := 5
	if countStr != "" {
		if c, err := fmt.Sscanf(countStr, "%d", &count); err == nil && c == 1 {
			if count < 1 {
				count = 1
			}
			if count > 50 {
				count = 50
			}
		}
	}

	result := CronResult{
		Expression: expr,
		Timezone:   loc.String(),
	}

	standardExpr := expr
	if alias, ok := cronAliases[strings.ToLower(expr)]; ok {
		standardExpr = alias
		result.StandardFormat = alias
	}

	if desc, ok := cronDescriptions[strings.ToLower(expr)]; ok {
		result.Description = desc
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(standardExpr)
	if err != nil {
		parser6 := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		schedule, err = parser6.Parse(standardExpr)
		if err != nil {
			result.IsValid = false
			result.Error = fmt.Sprintf("Invalid cron expression: %v", err)
			return result, nil
		}
		result.Fields = "Second Minute Hour DayOfMonth Month DayOfWeek"
	} else {
		result.Fields = "Minute Hour DayOfMonth Month DayOfWeek"
	}

	result.IsValid = true
	result.StandardFormat = standardExpr

	if result.Description == "" {
		result.Description = generateCronDescription(standardExpr)
	}

	now := time.Now().In(loc)
	nextRuns := make([]string, 0, count)
	for i := 0; i < count; i++ {
		next := schedule.Next(now)
		nextRuns = append(nextRuns, next.Format("2006-01-02 15:04:05 MST"))
		now = next
	}
	result.NextRuns = nextRuns

	return result, nil
}

func HTMXParseCron(c *gin.Context) {
	var req CronRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "cron-result", gin.H{"Error": "Invalid request parameters"})
		return
	}

	res, err := ParseCronLogic(req.Expression, string(req.Count), req.Timezone)
	if err != nil {
		c.HTML(http.StatusOK, "cron-result", gin.H{"Error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "cron-result", gin.H{
		"Expression":     res.Expression,
		"IsValid":        res.IsValid,
		"Error":          res.Error,
		"Description":    res.Description,
		"NextRuns":       res.NextRuns,
		"Timezone":       res.Timezone,
		"Fields":         res.Fields,
		"StandardFormat": res.StandardFormat,
		"Count":          string(req.Count),
	})
}

func V1ParseCron(c *gin.Context) {
	var req CronRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	res, err := ParseCronLogic(req.Expression, string(req.Count), req.Timezone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// generateCronDescription creates a human-readable description for a cron expression
func generateCronDescription(expr string) string {
	parts := strings.Fields(expr)
	if len(parts) < 5 {
		return "Custom schedule"
	}

	// 6-field expressions include seconds: second minute hour dom month dow
	hasSeconds := len(parts) == 6

	second, minute, hour, dom, month, dow := "*", parts[0], parts[1], parts[2], parts[3], parts[4]
	if hasSeconds {
		second, minute, hour, dom, month, dow = parts[0], parts[1], parts[2], parts[3], parts[4], parts[5]
	}

	var desc strings.Builder

	// Second
	if hasSeconds {
		if second == "*" {
			desc.WriteString("At every second")
		} else if strings.Contains(second, "/") {
			desc.WriteString(fmt.Sprintf("Every %s seconds", strings.Split(second, "/")[1]))
		} else {
			desc.WriteString(fmt.Sprintf("At second %s", second))
		}
		desc.WriteString(" of every minute")
	}

	// Minute
	if minute == "*" {
		desc.WriteString(" every minute")
	} else if strings.Contains(minute, "/") {
		desc.WriteString(fmt.Sprintf(" every %s minutes", strings.Split(minute, "/")[1]))
	} else if strings.Contains(minute, ",") {
		desc.WriteString(fmt.Sprintf(" at minutes: %s", minute))
	} else if strings.Contains(minute, "-") {
		desc.WriteString(fmt.Sprintf(" minutes %s through %s", strings.Split(minute, "-")[0], strings.Split(minute, "-")[1]))
	} else {
		desc.WriteString(fmt.Sprintf(" at minute %s", minute))
	}

	// Hour
	if hour != "*" {
		if strings.Contains(hour, "/") {
			desc.WriteString(fmt.Sprintf(" every %s hours", strings.Split(hour, "/")[1]))
		} else if strings.Contains(hour, ",") {
			desc.WriteString(fmt.Sprintf(" at hours: %s", hour))
		} else if strings.Contains(hour, "-") {
			desc.WriteString(fmt.Sprintf(" hours %s through %s", strings.Split(hour, "-")[0], strings.Split(hour, "-")[1]))
		} else {
			desc.WriteString(fmt.Sprintf(" at hour %s", hour))
		}
	}

	// Day of month
	if dom != "*" && dom != "?" {
		if strings.Contains(dom, "/") {
			desc.WriteString(fmt.Sprintf(" every %s days", strings.Split(dom, "/")[1]))
		} else if strings.Contains(dom, ",") {
			desc.WriteString(fmt.Sprintf(" on days: %s", dom))
		} else if strings.Contains(dom, "-") {
			desc.WriteString(fmt.Sprintf(" days %s through %s", strings.Split(dom, "-")[0], strings.Split(dom, "-")[1]))
		} else {
			desc.WriteString(fmt.Sprintf(" on day %s", dom))
		}
	}

	// Month
	if month != "*" {
		monthNames := []string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
		if strings.Contains(month, ",") {
			desc.WriteString(fmt.Sprintf(" in months: %s", month))
		} else if strings.Contains(month, "-") {
			desc.WriteString(fmt.Sprintf(" months %s through %s", strings.Split(month, "-")[0], strings.Split(month, "-")[1]))
		} else if monthNum, err := strconv.Atoi(month); err == nil && monthNum >= 1 && monthNum <= 12 {
			desc.WriteString(fmt.Sprintf(" in %s", monthNames[monthNum]))
		} else {
			desc.WriteString(fmt.Sprintf(" in %s", month))
		}
	}

	// Day of week
	if dow != "*" && dow != "?" {
		dowNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		if strings.Contains(dow, ",") {
			desc.WriteString(fmt.Sprintf(" on %s", dow))
		} else if strings.Contains(dow, "-") {
			dowParts := strings.Split(dow, "-")
			if start, err1 := strconv.Atoi(dowParts[0]); err1 == nil {
				if end, err2 := strconv.Atoi(dowParts[1]); err2 == nil {
					if start >= 0 && start <= 6 && end >= 0 && end <= 6 {
						desc.WriteString(fmt.Sprintf(" %s through %s", dowNames[start], dowNames[end]))
					}
				}
			}
		} else if dowNum, err := strconv.Atoi(dow); err == nil && dowNum >= 0 && dowNum <= 6 {
			desc.WriteString(fmt.Sprintf(" on %s", dowNames[dowNum]))
		} else {
			desc.WriteString(fmt.Sprintf(" on %s", dow))
		}
	}

	desc.WriteString(".")
	return strings.TrimSpace(desc.String())
}
