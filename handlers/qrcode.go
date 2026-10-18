package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
)

// QR code rendering bounds. Margin is counted in modules, size in pixels.
const (
	qrDefaultSize = 300
	qrMinSize     = 128
	qrMaxSize     = 1024

	qrDefaultMargin = 4
	qrMinMargin     = 0
	qrMaxMargin     = 8

	// Scanners need a clear luminance gap between modules and background.
	// Below this ratio we warn but still return the image.
	qrMinContrast = 3.0

	// Symbol version at which the payload is dense enough to be worth a
	// warning about scanning reliability.
	qrLargeVersion = 30

	qrMaxPayloadBytes = 2953 // Version 40-L byte-mode capacity.
)

// QRCodeRequest binds form data for generating a QR Code. Every field except
// Text is optional and falls back to a documented default, so existing clients
// that only send Text keep the original rendering.
type QRCodeRequest struct {
	Text   string `form:"text"   json:"text"`
	ECC    string `form:"ecc"    json:"ecc"`
	FG     string `form:"fg"     json:"fg"`
	BG     string `form:"bg"     json:"bg"`
	Preset string `form:"preset" json:"preset"`

	// Margin and Size are pointers so an omitted field is distinguishable from
	// an explicit 0. A zero quiet zone is a real choice, not an unset value.
	Margin *int `form:"margin" json:"margin"`
	Size   *int `form:"size"   json:"size"`

	// Preset payload fields. Only those relevant to the active preset are
	// read, and every one of them is optional.
	SSID      string `form:"ssid"      json:"ssid"`
	Password  string `form:"password"  json:"password"`
	Security  string `form:"security"  json:"security"`
	Hidden    string `form:"hidden"    json:"hidden"`
	Name      string `form:"name"      json:"name"`
	Phone     string `form:"phone"     json:"phone"`
	Email     string `form:"email"     json:"email"`
	Org       string `form:"org"       json:"org"`
	Note      string `form:"note"      json:"note"`
	Address   string `form:"address"   json:"address"`
	City      string `form:"city"      json:"city"`
	URL       string `form:"url"       json:"url"`
	Message   string `form:"message"   json:"message"`
	Latitude  string `form:"latitude"  json:"latitude"`
	Longitude string `form:"longitude" json:"longitude"`
}

// QRCodeResult is the full outcome of one generation. The HTMX fragment and
// the REST endpoint both render from it, so the two surfaces cannot disagree.
//
// Error is a field rather than a separate map shape because the HTMX template
// branches on it: the success and failure paths must pass the same type or the
// template fails to execute halfway through rendering.
type QRCodeResult struct {
	Error     string
	Text      string
	Payload   string
	ECC       string
	FG        string
	BG        string
	Margin    int
	ImageURL  template.URL
	SVGURL    template.URL
	RawBase64 string
	Version   int
	Modules   int
	ByteLen   int
	Size      int
	Contrast  float64
	Warning   string
}

// qrECCLevels maps the spec's recovery levels onto the library's names. Note
// that the library calls level Q "High", which is unrelated to its "Highest".
var qrECCLevels = map[string]qrcode.RecoveryLevel{
	"l": qrcode.Low,     // 7%
	"m": qrcode.Medium,  // 15%
	"q": qrcode.High,    // 25%
	"h": qrcode.Highest, // 30%
}

// parseQRColor accepts #rgb and #rrggbb. Alpha forms are rejected so the code
// always renders opaque. Invalid input falls back to the default rather than
// failing the request: a mistyped hex should not block the user from seeing
// their code.
func parseQRColor(raw, fallback string) (color.RGBA, string) {
	if c, ok := tryParseQRColor(raw); ok {
		return c, hexString(c)
	}
	c, _ := tryParseQRColor(fallback)
	return c, hexString(c)
}

func tryParseQRColor(raw string) (color.RGBA, bool) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "#") {
		return color.RGBA{}, false
	}
	digits := raw[1:]

	// Expand the shorthand form: "#abc" -> "#aabbcc".
	if len(digits) == 3 {
		digits = string([]byte{
			digits[0], digits[0],
			digits[1], digits[1],
			digits[2], digits[2],
		})
	}
	if len(digits) != 6 {
		return color.RGBA{}, false
	}

	v, err := strconv.ParseUint(digits, 16, 32)
	if err != nil {
		return color.RGBA{}, false
	}
	return color.RGBA{
		R: uint8(v >> 16),
		G: uint8(v >> 8),
		B: uint8(v),
		A: 0xff,
	}, true
}

func hexString(c color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// clampInt clamps v into [lo, hi]. The caller supplies the default, so a
// genuine 0 survives when 0 is a legal value.
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clampOptional resolves a possibly-absent option to its default, then clamps.
func clampOptional(v *int, lo, hi, fallback int) int {
	if v == nil {
		return fallback
	}
	return clampInt(*v, lo, hi)
}

// escapeWifiField applies the escaping the WIFI: scheme requires. A literal
// backslash has to be doubled first, otherwise it would escape the next
// separator instead of itself.
func escapeWifiField(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, ":", "\\:")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// escapeSemicolonField escapes only the semicolon terminator, which is what
// the MECARD: scheme requires for its field values.
func escapeSemicolonField(s string) string {
	return strings.ReplaceAll(s, ";", "\\;")
}

func escapeVCARDField(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func hasScheme(s string) bool {
	lower := strings.ToLower(s)
	for _, p := range []string{"http://", "https://", "mailto:", "tel:", "ftp://", "sms:", "smsto:", "geo:"} {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

// buildQRPayload expands the active preset into the final encoded string.
// Presets resolve server-side so the REST endpoint and the UI can never
// disagree about what was actually encoded.
func buildQRPayload(req QRCodeRequest) (string, error) {
	preset := strings.ToLower(strings.TrimSpace(req.Preset))

	switch preset {
	case "", "text":
		if strings.TrimSpace(req.Text) == "" {
			return "", fmt.Errorf("input text cannot be empty")
		}
		return req.Text, nil

	case "url":
		target := firstNonEmpty(req.URL, req.Text)
		if target == "" {
			return "", fmt.Errorf("a URL is required for the URL preset")
		}
		if !hasScheme(target) {
			target = "https://" + target
		}
		return target, nil

	case "mailto":
		addr := firstNonEmpty(req.Email, req.Text)
		if addr == "" {
			return "", fmt.Errorf("an email address is required for the mailto preset")
		}
		payload := "mailto:" + addr
		if req.Note != "" {
			payload += "?body=" + url.QueryEscape(req.Note)
		}
		return payload, nil

	case "tel":
		number := firstNonEmpty(req.Phone, req.Text)
		if number == "" {
			return "", fmt.Errorf("a phone number is required for the tel preset")
		}
		return "tel:" + number, nil

	case "sms":
		number := firstNonEmpty(req.Phone, req.Text)
		if number == "" {
			return "", fmt.Errorf("a phone number is required for the SMS preset")
		}
		payload := "SMSTO:" + number
		if req.Message != "" {
			payload += ":" + req.Message
		}
		return payload, nil

	case "geo":
		lat, lon := strings.TrimSpace(req.Latitude), strings.TrimSpace(req.Longitude)
		if lat == "" || lon == "" {
			return "", fmt.Errorf("both latitude and longitude are required for the geo preset")
		}
		if _, err := strconv.ParseFloat(lat, 64); err != nil {
			return "", fmt.Errorf("latitude must be a number")
		}
		if _, err := strconv.ParseFloat(lon, 64); err != nil {
			return "", fmt.Errorf("longitude must be a number")
		}
		return "geo:" + lat + "," + lon, nil

	case "wifi":
		ssid := strings.TrimSpace(req.SSID)
		if ssid == "" {
			return "", fmt.Errorf("a network name (SSID) is required for the WiFi preset")
		}
		auth := strings.ToUpper(strings.TrimSpace(req.Security))
		if auth == "" {
			if req.Password == "" {
				auth = "nopass"
			} else {
				auth = "WPA"
			}
		}
		payload := "WIFI:T:" + auth + ";S:" + escapeWifiField(ssid) + ";"
		if req.Password != "" {
			payload += "P:" + escapeWifiField(req.Password) + ";"
		}
		// The hidden field is only meaningful for open networks; adding it
		// alongside a password would contradict itself.
		if strings.EqualFold(auth, "nopass") {
			payload += "T:nopass;"
		}
		if strings.EqualFold(strings.TrimSpace(req.Hidden), "true") {
			payload += "H:true;"
		}
		return payload + ";", nil

	case "mecard":
		name := firstNonEmpty(req.Name, req.Text)
		if name == "" {
			return "", fmt.Errorf("a name is required for the MECARD preset")
		}
		payload := "MECARD:N:" + escapeSemicolonField(name) + ";"
		if req.Phone != "" {
			payload += "TEL:" + escapeSemicolonField(req.Phone) + ";"
		}
		if req.Email != "" {
			payload += "EMAIL:" + escapeSemicolonField(req.Email) + ";"
		}
		if req.Note != "" {
			payload += "NOTE:" + escapeSemicolonField(req.Note) + ";"
		}
		return payload + ";", nil

	case "vcard":
		name := firstNonEmpty(req.Name, req.Text)
		if name == "" {
			return "", fmt.Errorf("a name is required for the vCard preset")
		}
		var b strings.Builder
		b.WriteString("BEGIN:VCARD\nVERSION:3.0\n")
		b.WriteString("N:" + escapeVCARDField(name) + "\n")
		b.WriteString("FN:" + escapeVCARDField(name) + "\n")
		if req.Org != "" {
			b.WriteString("ORG:" + escapeVCARDField(req.Org) + "\n")
		}
		if req.Phone != "" {
			b.WriteString("TEL;TYPE=CELL:" + req.Phone + "\n")
		}
		if req.Email != "" {
			b.WriteString("EMAIL:" + req.Email + "\n")
		}
		if req.Address != "" || req.City != "" {
			b.WriteString("ADR:;;" + escapeVCARDField(req.Address) + ";" + escapeVCARDField(req.City) + ";;\n")
		}
		if req.Note != "" {
			b.WriteString("NOTE:" + escapeVCARDField(req.Note) + "\n")
		}
		if strings.TrimSpace(req.URL) != "" {
			b.WriteString("URL:" + req.URL + "\n")
		}
		b.WriteString("END:VCARD")
		return b.String(), nil

	default:
		return "", fmt.Errorf("unknown QR preset: %s", preset)
	}
}

// relativeLuminance implements the WCAG 2.x luminance formula used by the
// contrast ratio check.
func relativeLuminance(c color.Color) float64 {
	sr, sg, sb, _ := c.RGBA()
	lin := func(v uint32) float64 {
		cs := float64(v) / 65535.0
		if cs <= 0.03928 {
			return cs / 12.92
		}
		return math.Pow((cs+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(sr) + 0.7152*lin(sg) + 0.0722*lin(sb)
}

// contrastRatio returns the WCAG contrast ratio between two colors, 1 to 21.
func contrastRatio(a, b color.Color) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// renderQRMatrix pads the bare symbol matrix with the requested quiet zone.
// The library's built-in border is disabled because it is fixed at four
// modules with no setter; rendering both outputs from one padded matrix keeps
// the PNG and the SVG in exact agreement.
func renderQRMatrix(q *qrcode.QRCode, margin int) [][]bool {
	bare := q.Bitmap()

	height := len(bare)
	width := 0
	for _, row := range bare {
		if len(row) > width {
			width = len(row)
		}
	}
	if height == 0 || width == 0 {
		return nil
	}

	total := height + 2*margin
	padded := make([][]bool, total)
	for y := 0; y < total; y++ {
		padded[y] = make([]bool, total)
		for x := 0; x < total; x++ {
			sy, sx := y-margin, x-margin
			if sy >= 0 && sy < height && sx >= 0 && sx < width {
				padded[y][x] = bare[sy][sx]
			}
		}
	}
	return padded
}

// renderQRPNG draws the padded matrix at an integer module size so no module
// is ever resampled and edges stay crisp for scanning.
func renderQRPNG(matrix [][]bool, fg, bg color.Color, size int) ([]byte, int, error) {
	n := len(matrix)
	if n == 0 {
		return nil, 0, fmt.Errorf("empty QR matrix")
	}
	if size < n {
		return nil, 0, fmt.Errorf("requested size %dpx is too small for a %d-module code", size, n)
	}

	// Round the module size up so the canvas is never smaller than requested.
	// At the upper bound, round down instead so we never exceed it.
	moduleSize := size / n
	if moduleSize*n < size {
		if moduleSize*n+n <= qrMaxSize {
			moduleSize++
		}
	}
	canvas := moduleSize * n

	img := image.NewRGBA(image.Rect(0, 0, canvas, canvas))
	draw.Draw(img, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)

	fgUniform := image.NewUniform(fg)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if !matrix[y][x] {
				continue
			}
			draw.Draw(img,
				image.Rect(x*moduleSize, y*moduleSize, (x+1)*moduleSize, (y+1)*moduleSize),
				fgUniform, image.Point{}, draw.Src)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, 0, fmt.Errorf("failed to encode QR code PNG: %v", err)
	}
	return buf.Bytes(), canvas, nil
}

// renderQRSVG emits one merged path over a viewBox, so the output stays small
// and scales losslessly for print. shape-rendering keeps module edges hard.
// Colors arrive as resolved hex strings so the SVG matches the PNG exactly.
func renderQRSVG(matrix [][]bool, fgHex, bgHex string, size int) string {
	n := len(matrix)
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="`)
	b.WriteString(strconv.Itoa(size))
	b.WriteString(`" height="`)
	b.WriteString(strconv.Itoa(size))
	b.WriteString(`" viewBox="0 0 `)
	b.WriteString(strconv.Itoa(n))
	b.WriteByte(' ')
	b.WriteString(strconv.Itoa(n))
	b.WriteString(`" shape-rendering="crispEdges" role="img">`)
	b.WriteString(`<rect width="100%" height="100%" fill="`)
	b.WriteString(bgHex)
	b.WriteString(`"/><path fill="`)
	b.WriteString(fgHex)
	b.WriteString(`" d="`)

	// Merge horizontal runs of dark modules into single rectangles.
	for y := 0; y < n; y++ {
		for x := 0; x < n; {
			if !matrix[y][x] {
				x++
				continue
			}
			start := x
			for x < n && matrix[y][x] {
				x++
			}
			b.WriteString("M")
			b.WriteString(strconv.Itoa(start))
			b.WriteByte(' ')
			b.WriteString(strconv.Itoa(y))
			b.WriteString("h")
			b.WriteString(strconv.Itoa(x - start))
			b.WriteString("v1h-")
			b.WriteString(strconv.Itoa(x - start))
			b.WriteString("z")
		}
	}

	b.WriteString(`"/></svg>`)
	return b.String()
}

// qrWarning composes the advisory messages for a render. Low contrast and a
// zero quiet zone are both real scanning risks, but neither blocks output: the
// user asked to see the code, so we show it and explain the trade-off.
func qrWarning(contrast float64, margin, version int) string {
	var warnings []string

	if contrast < qrMinContrast {
		warnings = append(warnings, fmt.Sprintf(
			"Contrast ratio is %.1f:1. Most scanners need at least %.0f:1 — darken the modules or lighten the background.",
			contrast, qrMinContrast))
	}
	if margin == 0 {
		warnings = append(warnings,
			"A zero quiet zone leaves no clear margin around the code. Many scanners will fail to read it.")
	}
	if version >= qrLargeVersion {
		warnings = append(warnings, fmt.Sprintf(
			"This payload uses symbol version %d (near the QR maximum). Keep the module count high and scan at close range.", version))
	}
	return strings.Join(warnings, " ")
}

// GenerateQRCodeLogic performs core QR code generation, returning the full
// result set used by both the HTMX fragment and the REST endpoint.
func GenerateQRCodeLogic(req QRCodeRequest) (*QRCodeResult, error) {
	payload, err := buildQRPayload(req)
	if err != nil {
		return nil, err
	}
	if len(payload) > qrMaxPayloadBytes {
		return nil, fmt.Errorf("payload is %d bytes; QR codes hold at most %d", len(payload), qrMaxPayloadBytes)
	}

	ecc := strings.ToLower(strings.TrimSpace(req.ECC))
	if ecc == "" {
		ecc = "m"
	}
	level, ok := qrECCLevels[ecc]
	if !ok {
		ecc = "m"
		level = qrcode.Medium
	}

	fg, fgHexValue := parseQRColor(req.FG, "#000000")
	bg, bgHexValue := parseQRColor(req.BG, "#ffffff")

	margin := clampOptional(req.Margin, qrMinMargin, qrMaxMargin, qrDefaultMargin)
	size := clampOptional(req.Size, qrMinSize, qrMaxSize, qrDefaultSize)

	q, err := qrcode.New(payload, level)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize QR code: %v", err)
	}
	// Disable the fixed four-module border so the margin control owns the
	// quiet zone instead of stacking on top of it.
	q.DisableBorder = true

	matrix := renderQRMatrix(q, margin)
	if matrix == nil {
		return nil, fmt.Errorf("failed to render QR matrix")
	}

	// A dense symbol can need more pixels than requested. Grow to the next
	// whole module boundary that fits rather than failing.
	modules := len(matrix)
	if size < modules {
		size = modules
	}

	pngBytes, actualSize, err := renderQRPNG(matrix, fg, bg, size)
	if err != nil {
		return nil, err
	}
	svg := renderQRSVG(matrix, fgHexValue, bgHexValue, actualSize)

	base64Img := base64.StdEncoding.EncodeToString(pngBytes)
	contrast := contrastRatio(fg, bg)

	return &QRCodeResult{
		Text:      req.Text,
		Payload:   payload,
		ECC:       ecc,
		FG:        fgHexValue,
		BG:        bgHexValue,
		Margin:    margin,
		ImageURL:  template.URL("data:image/png;base64," + base64Img),
		SVGURL:    template.URL("data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))),
		RawBase64: base64Img,
		Version:   q.VersionNumber,
		Modules:   modules,
		ByteLen:   len(payload),
		Size:      actualSize,
		Contrast:  contrast,
		Warning:   qrWarning(contrast, margin, q.VersionNumber),
	}, nil
}

// HTMXGenerateQRCode handles the QR code generation request from HTMX
func HTMXGenerateQRCode(c *gin.Context) {
	var req QRCodeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "qrcode-result", &QRCodeResult{
			Error: "Invalid request parameters",
		})
		return
	}

	result, err := GenerateQRCodeLogic(req)
	if err != nil {
		c.HTML(http.StatusOK, "qrcode-result", &QRCodeResult{
			Error: err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "qrcode-result", result)
}

// V1GenerateQRCode handles JSON REST API requests to generate a QR code
func V1GenerateQRCode(c *gin.Context) {
	var req QRCodeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	result, err := GenerateQRCodeLogic(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_url":      string(result.ImageURL),
		"svg_url":        string(result.SVGURL),
		"text":           result.Text,
		"payload":        result.Payload,
		"raw_base64":     result.RawBase64,
		"ecc":            result.ECC,
		"foreground":     result.FG,
		"background":     result.BG,
		"margin":         result.Margin,
		"size":           result.Size,
		"version":        result.Version,
		"modules":        result.Modules,
		"byte_length":    result.ByteLen,
		"contrast_ratio": math.Round(result.Contrast*10) / 10,
		"warning":        result.Warning,
	})
}
