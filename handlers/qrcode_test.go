package handlers

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	qrcode "github.com/skip2/go-qrcode"
)

// decodeResultPNG pulls the rendered PNG back out of the data URL so tests
// assert on real pixels rather than on the encoding.
// intPtr builds a pointer for the optional numeric request fields, so a test
// can state an explicit 0 and have it mean 0.
func intPtr(v int) *int { return &v }

func decodeResultPNG(t *testing.T, res *QRCodeResult) {
	t.Helper()

	payload := strings.TrimPrefix(string(res.ImageURL), "data:image/png;base64,")
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("result data URL is not valid base64: %v", err)
	}
	if !bytes.Equal(raw, mustDecodeBase64(t, res.RawBase64)) {
		t.Error("image_url payload does not match raw_base64")
	}

	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("result is not a decodable PNG: %v", err)
	}
	if got := img.Bounds().Dx(); got != res.Size {
		t.Errorf("image width = %d, want Size %d", got, res.Size)
	}
	if img.Bounds().Dy() != img.Bounds().Dx() {
		t.Errorf("image is not square: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func mustDecodeBase64(t *testing.T, s string) []byte {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("raw_base64 is not valid base64: %v", err)
	}
	return raw
}

// decodeResultSVG unwraps the SVG data URL so assertions can read the markup.
func decodeResultSVG(t *testing.T, res *QRCodeResult) string {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(
		strings.TrimPrefix(string(res.SVGURL), "data:image/svg+xml;base64,"))
	if err != nil {
		t.Fatalf("SVG data URL is not valid base64: %v", err)
	}
	return string(raw)
}

func TestGenerateQRCodeDefaults(t *testing.T) {
	res, err := GenerateQRCodeLogic(QRCodeRequest{Text: "https://example.com"})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}

	if res.ECC != "m" {
		t.Errorf("ECC = %q, want %q", res.ECC, "m")
	}
	if res.FG != "#000000" {
		t.Errorf("FG = %q, want #000000", res.FG)
	}
	if res.BG != "#ffffff" {
		t.Errorf("BG = %q, want #ffffff", res.BG)
	}
	if res.Margin != qrDefaultMargin {
		t.Errorf("Margin = %d, want %d", res.Margin, qrDefaultMargin)
	}
	if res.Payload != "https://example.com" {
		t.Errorf("Payload = %q, want the input text", res.Payload)
	}
	if res.Warning != "" {
		t.Errorf("Warning = %q, want empty for default options", res.Warning)
	}
	decodeResultPNG(t, res)
}

func TestGenerateQRCodeRejectsEmptyText(t *testing.T) {
	if _, err := GenerateQRCodeLogic(QRCodeRequest{}); err == nil {
		t.Fatal("expected an error for empty text")
	}
}

func TestGenerateQRCodeRejectsUnknownPreset(t *testing.T) {
	_, err := GenerateQRCodeLogic(QRCodeRequest{Text: "hello", Preset: "nonsense"})
	if err == nil {
		t.Fatal("expected an error for an unknown preset")
	}
	if !strings.Contains(err.Error(), "nonsense") {
		t.Errorf("error = %q, want it to name the unknown preset", err)
	}
}

func TestGenerateQRCodeRejectsOversizedPayload(t *testing.T) {
	huge := strings.Repeat("A", qrMaxPayloadBytes+1)
	if _, err := GenerateQRCodeLogic(QRCodeRequest{Text: huge}); err == nil {
		t.Fatal("expected an error for an oversized payload")
	}
}

// Higher recovery levels carry more error-correction codewords, so a fixed
// payload must produce a denser symbol at Q than at L.
// Higher recovery levels carry more error-correction codewords, so a fixed
// payload needs a denser symbol as the level rises. Versions are discrete, so
// adjacent levels often share one; what must hold is that the version never
// decreases and that the extremes are strictly ordered.
func TestQRCodeECCAffectsSymbolSize(t *testing.T) {
	levels := []string{"l", "m", "q", "h"}
	versions := make(map[string]int, len(levels))

	for _, ecc := range levels {
		res, err := GenerateQRCodeLogic(QRCodeRequest{Text: strings.Repeat("A", 100), ECC: ecc})
		if err != nil {
			t.Fatalf("ECC %q: %v", ecc, err)
		}
		versions[ecc] = res.Version
	}

	for i := 1; i < len(levels); i++ {
		if versions[levels[i]] < versions[levels[i-1]] {
			t.Errorf("version %s = %d, want >= %s = %d",
				levels[i], versions[levels[i]], levels[i-1], versions[levels[i-1]])
		}
	}
	if versions["h"] <= versions["l"] {
		t.Errorf("version H = %d, want strictly larger than L = %d", versions["h"], versions["l"])
	}
}

func TestQRCodeECCDefaultsToMediumOnUnknownValue(t *testing.T) {
	res, err := GenerateQRCodeLogic(QRCodeRequest{Text: "hi", ECC: "banana"})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}
	if res.ECC != "m" {
		t.Errorf("ECC = %q, want fallback to m", res.ECC)
	}
}

func TestQRCodeMarginAddsModules(t *testing.T) {
	// The symbol itself is fixed for a given payload; only the quiet zone
	// changes, by two modules per extra margin step.
	base, err := GenerateQRCodeLogic(QRCodeRequest{Text: "margin test", Margin: intPtr(4)})
	if err != nil {
		t.Fatalf("margin 4: %v", err)
	}
	wide, err := GenerateQRCodeLogic(QRCodeRequest{Text: "margin test", Margin: intPtr(8)})
	if err != nil {
		t.Fatalf("margin 8: %v", err)
	}
	none, err := GenerateQRCodeLogic(QRCodeRequest{Text: "margin test", Margin: intPtr(0)})
	if err != nil {
		t.Fatalf("margin 0: %v", err)
	}

	if got, want := wide.Modules, base.Modules+8; got != want {
		t.Errorf("margin 8 modules = %d, want %d", got, want)
	}
	if got, want := none.Modules, base.Modules-8; got != want {
		t.Errorf("margin 0 modules = %d, want %d", got, want)
	}
	if none.Warning == "" {
		t.Error("expected a warning for a zero quiet zone")
	}
}

// A requested pixel size rarely divides evenly by the module count, so the
// renderer snaps to a whole module boundary. It rounds up so the code is never
// smaller than asked for, except at the upper bound where it rounds down to
// stay within qrMaxSize.
func TestQRCodeSizeSnapsToWholeModules(t *testing.T) {
	for _, requested := range []int{128, 300, 512, 1000} {
		res, err := GenerateQRCodeLogic(QRCodeRequest{
			Text: "https://example.com/some/path", Size: intPtr(requested)})
		if err != nil {
			t.Fatalf("size %d: %v", requested, err)
		}

		if res.Size%res.Modules != 0 {
			t.Errorf("size %d: Size = %d is not a multiple of Modules = %d",
				requested, res.Size, res.Modules)
		}
		if res.Size > qrMaxSize {
			t.Errorf("size %d: Size = %d, want at most %d", requested, res.Size, qrMaxSize)
		}
		// Snapping lands within one module of the request in either
		// direction: up normally, down when rounding up would exceed the cap.
		if diff := res.Size - requested; diff >= res.Modules || diff <= -res.Modules {
			t.Errorf("size %d: Size = %d is more than one module (%d) away",
				requested, res.Size, res.Modules)
		}
		decodeResultPNG(t, res)
	}
}

func TestQRCodeSizeClampedToBounds(t *testing.T) {
	small, err := GenerateQRCodeLogic(QRCodeRequest{Text: "clamp", Size: intPtr(10)})
	if err != nil {
		t.Fatalf("tiny size: %v", err)
	}
	if small.Size < qrMinSize {
		t.Errorf("Size = %d, want clamped up to at least %d", small.Size, qrMinSize)
	}

	huge, err := GenerateQRCodeLogic(QRCodeRequest{Text: "clamp", Size: intPtr(99999)})
	if err != nil {
		t.Fatalf("huge size: %v", err)
	}
	if huge.Size > qrMaxSize {
		t.Errorf("Size = %d, want clamped down to at most %d", huge.Size, qrMaxSize)
	}
}

func TestQRCodeMarginClampedToBounds(t *testing.T) {
	over, err := GenerateQRCodeLogic(QRCodeRequest{Text: "clamp", Margin: intPtr(50)})
	if err != nil {
		t.Fatalf("margin 50: %v", err)
	}
	if over.Margin != qrMaxMargin {
		t.Errorf("Margin = %d, want clamped to %d", over.Margin, qrMaxMargin)
	}

	under, err := GenerateQRCodeLogic(QRCodeRequest{Text: "clamp", Margin: intPtr(-5)})
	if err != nil {
		t.Fatalf("margin -5: %v", err)
	}
	if under.Margin != qrMinMargin {
		t.Errorf("Margin = %d, want clamped to %d", under.Margin, qrMinMargin)
	}
}

func TestQRCodeColorsApplied(t *testing.T) {
	res, err := GenerateQRCodeLogic(QRCodeRequest{
		Text: "color test",
		FG:   "#1a2b3c",
		BG:   "#ffffff",
	})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}
	if res.FG != "#1a2b3c" {
		t.Errorf("FG = %q, want #1a2b3c", res.FG)
	}
	svg := decodeResultSVG(t, res)
	if !strings.Contains(svg, "1a2b3c") {
		t.Error("SVG does not carry the requested foreground color")
	}
	decodeResultPNG(t, res)
}

func TestQRCodeColorShorthandExpanded(t *testing.T) {
	res, err := GenerateQRCodeLogic(QRCodeRequest{Text: "shorthand", FG: "#abc"})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}
	if res.FG != "#aabbcc" {
		t.Errorf("FG = %q, want #aabbcc", res.FG)
	}
}

func TestQRCodeInvalidColorFallsBack(t *testing.T) {
	// A mistyped hex must not block the render.
	res, err := GenerateQRCodeLogic(QRCodeRequest{Text: "bad color", FG: "not-a-color", BG: "#12345"})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}
	if res.FG != "#000000" {
		t.Errorf("FG = %q, want fallback #000000", res.FG)
	}
	if res.BG != "#ffffff" {
		t.Errorf("BG = %q, want fallback #ffffff", res.BG)
	}
}

func TestQRCodeContrastRatio(t *testing.T) {
	high, err := GenerateQRCodeLogic(QRCodeRequest{Text: "contrast", FG: "#000000", BG: "#ffffff"})
	if err != nil {
		t.Fatalf("high contrast: %v", err)
	}
	if high.Contrast < 20.9 || high.Contrast > 21.1 {
		t.Errorf("black on white contrast = %.2f, want ~21", high.Contrast)
	}
	if high.Warning != "" {
		t.Errorf("Warning = %q, want none for maximum contrast", high.Warning)
	}

	low, err := GenerateQRCodeLogic(QRCodeRequest{Text: "contrast", FG: "#bbbbbb", BG: "#ffffff"})
	if err != nil {
		t.Fatalf("low contrast: %v", err)
	}
	if low.Contrast >= qrMinContrast {
		t.Errorf("grey on white contrast = %.2f, want below the %.1f threshold", low.Contrast, qrMinContrast)
	}
	if !strings.Contains(low.Warning, "Contrast ratio") {
		t.Errorf("Warning = %q, want a contrast warning", low.Warning)
	}
}

func TestQRCodeLargeVersionHint(t *testing.T) {
	res, err := GenerateQRCodeLogic(QRCodeRequest{
		Text: strings.Repeat("x", 1000),
		ECC:  "h",
		Size: intPtr(1024),
	})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}
	if res.Version < qrLargeVersion {
		t.Fatalf("payload produced version %d, want at least %d to reach the hint threshold", res.Version, qrLargeVersion)
	}
	if !strings.Contains(res.Warning, "symbol version") {
		t.Errorf("Warning = %q, want a large-version hint", res.Warning)
	}
}

func TestQRCodeMetadataReadout(t *testing.T) {
	payload := "https://example.com/very/long/path/for/metadata"
	res, err := GenerateQRCodeLogic(QRCodeRequest{Text: payload, Margin: intPtr(4)})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}
	if res.ByteLen != len(payload) {
		t.Errorf("ByteLen = %d, want %d", res.ByteLen, len(payload))
	}
	if res.Version < 1 || res.Version > 40 {
		t.Errorf("Version = %d, want 1-40", res.Version)
	}
	// A QR symbol is always 4*version+17 modules across, plus the quiet zone.
	wantModules := 4*res.Version + 17 + 2*res.Margin
	if res.Modules != wantModules {
		t.Errorf("Modules = %d, want %d", res.Modules, wantModules)
	}
}

func TestRenderQRSVGShape(t *testing.T) {
	res, err := GenerateQRCodeLogic(QRCodeRequest{Text: "svg test", FG: "#000000", BG: "#ffffff"})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}

	out := decodeResultSVG(t, res)

	if !strings.HasPrefix(out, `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Error("SVG output is missing its namespace declaration")
	}
	if !strings.Contains(out, `viewBox="0 0 `) {
		t.Error("SVG output is missing a viewBox")
	}
	if !strings.Contains(out, `shape-rendering="crispEdges"`) {
		t.Error("SVG output is missing crispEdges rendering")
	}
	if !strings.HasSuffix(out, "</svg>") {
		t.Error("SVG output is not closed")
	}

	viewBox := `viewBox="0 0 ` + itoa(res.Modules) + " " + itoa(res.Modules) + `"`
	if !strings.Contains(out, viewBox) {
		t.Errorf("SVG viewBox = missing %q", viewBox)
	}
	if !strings.Contains(out, `width="`+itoa(res.Size)+`"`) {
		t.Errorf("SVG width does not match the rendered size %d", res.Size)
	}
	if !strings.Contains(out, "fill=\"#000000\"") {
		t.Error("SVG path does not use the requested foreground color")
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

func TestQRCodePayloadPresets(t *testing.T) {
	tests := []struct {
		name string
		req  QRCodeRequest
		want string
	}{
		{
			name: "url adds scheme",
			req:  QRCodeRequest{Text: "example.com", Preset: "url"},
			want: "https://example.com",
		},
		{
			name: "url uses its dedicated field without generic text",
			req:  QRCodeRequest{URL: "google.com", Preset: "url"},
			want: "https://google.com",
		},
		{
			name: "url keeps existing scheme",
			req:  QRCodeRequest{Text: "http://example.com", Preset: "url"},
			want: "http://example.com",
		},
		{
			name: "mailto with body",
			req:  QRCodeRequest{Email: "ada@example.com", Note: "ping", Preset: "mailto"},
			want: "mailto:ada@example.com?body=ping",
		},
		{
			name: "tel",
			req:  QRCodeRequest{Phone: "+15551234567", Preset: "tel"},
			want: "tel:+15551234567",
		},
		{
			name: "sms with message",
			req:  QRCodeRequest{Phone: "+15551234567", Message: "hi", Preset: "sms"},
			want: "SMSTO:+15551234567:hi",
		},
		{
			name: "geo",
			req:  QRCodeRequest{Latitude: "37.7749", Longitude: "-122.4194", Preset: "geo"},
			want: "geo:37.7749,-122.4194",
		},
		{
			name: "wifi defaults to WPA when a password is set",
			req:  QRCodeRequest{SSID: "Office", Password: "hunter2", Preset: "wifi"},
			want: `WIFI:T:WPA;S:Office;P:hunter2;;`,
		},
		{
			name: "wifi open network",
			req:  QRCodeRequest{SSID: "Guest", Preset: "wifi"},
			want: "WIFI:T:nopass;S:Guest;T:nopass;;",
		},
		{
			name: "wifi hidden",
			req:  QRCodeRequest{SSID: "Guest", Hidden: "true", Preset: "wifi"},
			want: "WIFI:T:nopass;S:Guest;T:nopass;H:true;;",
		},
		{
			name: "wifi escapes reserved characters",
			req:  QRCodeRequest{SSID: "My;Net,Work:1", Password: `p;a"ss`, Preset: "wifi"},
			want: `WIFI:T:WPA;S:My\;Net\,Work\:1;P:p\;a\"ss;;`,
		},
		{
			name: "mecard",
			req:  QRCodeRequest{Name: "Ada", Phone: "+15551234567", Email: "ada@example.com", Preset: "mecard"},
			want: "MECARD:N:Ada;TEL:+15551234567;EMAIL:ada@example.com;;",
		},
		{
			name: "mecard escapes semicolons",
			req:  QRCodeRequest{Name: "Ada;Lovelace", Preset: "mecard"},
			want: `MECARD:N:Ada\;Lovelace;;`,
		},
		{
			name: "vcard",
			req:  QRCodeRequest{Name: "Ada Lovelace", Org: "Analytical Engines", Preset: "vcard"},
			want: "BEGIN:VCARD\nVERSION:3.0\nN:Ada Lovelace\nFN:Ada Lovelace\nORG:Analytical Engines\nEND:VCARD",
		},
		{
			name: "vcard escapes separators",
			req:  QRCodeRequest{Name: "Lovelace; Ada", Preset: "vcard"},
			want: "BEGIN:VCARD\nVERSION:3.0\nN:Lovelace\\; Ada\nFN:Lovelace\\; Ada\nEND:VCARD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildQRPayload(tt.req)
			if err != nil {
				t.Fatalf("buildQRPayload: %v", err)
			}
			if got != tt.want {
				t.Errorf("payload =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestQRCodePresetMissingFields(t *testing.T) {
	tests := []struct {
		name string
		req  QRCodeRequest
	}{
		{"wifi without ssid", QRCodeRequest{Preset: "wifi"}},
		{"geo without coordinates", QRCodeRequest{Preset: "geo", Latitude: "1"}},
		{"geo with non-numeric latitude", QRCodeRequest{Preset: "geo", Latitude: "north", Longitude: "1"}},
		{"mecard without name", QRCodeRequest{Preset: "mecard", Phone: "+1"}},
		{"vcard without name", QRCodeRequest{Preset: "vcard"}},
		{"tel without number", QRCodeRequest{Preset: "tel"}},
		{"mailto without address", QRCodeRequest{Preset: "mailto"}},
		{"sms without number", QRCodeRequest{Preset: "sms"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildQRPayload(tt.req); err == nil {
				t.Error("expected an error for an incomplete preset")
			}
		})
	}
}

func TestQRCodePresetRendersThroughTheFullPath(t *testing.T) {
	res, err := GenerateQRCodeLogic(QRCodeRequest{
		Preset:   "wifi",
		SSID:     "Office; Guest",
		Password: "hunter2",
		ECC:      "h",
		Size:     intPtr(512),
	})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}
	if !strings.HasPrefix(res.Payload, "WIFI:T:") {
		t.Errorf("Payload = %q, want a WIFI: prefix", res.Payload)
	}
	if res.ECC != "h" {
		t.Errorf("ECC = %q, want h", res.ECC)
	}
	decodeResultPNG(t, res)
}

func TestContrastRatioSymmetry(t *testing.T) {
	fg, _ := tryParseQRColor("#1a2b3c")
	bg, _ := tryParseQRColor("#f0e0d0")
	if a, b := contrastRatio(fg, bg), contrastRatio(bg, fg); a != b {
		t.Errorf("contrast ratio is not symmetric: %.4f vs %.4f", a, b)
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		v, lo, hi, want int
	}{
		{0, 0, 8, 0},  // zero is a legal margin, not "unset"
		{5, 0, 8, 5},  // in range
		{50, 0, 8, 8}, // above
		{-5, 0, 8, 0}, // below
		{300, 128, 1024, 300},
		{10, 128, 1024, 128},
		{99999, 128, 1024, 1024},
	}
	for _, tt := range tests {
		if got := clampInt(tt.v, tt.lo, tt.hi); got != tt.want {
			t.Errorf("clampInt(%d, %d, %d) = %d, want %d", tt.v, tt.lo, tt.hi, got, tt.want)
		}
	}
}

func TestClampOptional(t *testing.T) {
	if got := clampOptional(nil, 0, 8, 4); got != 4 {
		t.Errorf("absent margin = %d, want the default 4", got)
	}
	if got := clampOptional(intPtr(0), 0, 8, 4); got != 0 {
		t.Errorf("explicit zero margin = %d, want 0", got)
	}
	if got := clampOptional(intPtr(99), 0, 8, 4); got != 8 {
		t.Errorf("margin 99 = %d, want clamped to 8", got)
	}
}

// The original implementation used qrcode.New(text, Medium).Image(300), which
// draws the symbol with the library's built-in four-module quiet zone. Our
// default path disables that border and re-adds the quiet zone ourselves, so
// the module matrix must come out identical for existing clients.
func TestQRCodeDefaultMatchesLibraryDefaults(t *testing.T) {
	for _, text := range []string{"https://example.com", "hello world", "a"} {
		t.Run(text, func(t *testing.T) {
			reference, err := qrcode.New(text, qrcode.Medium)
			if err != nil {
				t.Fatalf("qrcode.New: %v", err)
			}
			want := reference.Bitmap()

			got, err := GenerateQRCodeLogic(QRCodeRequest{Text: text})
			if err != nil {
				t.Fatalf("GenerateQRCodeLogic: %v", err)
			}
			if got.Version != reference.VersionNumber {
				t.Errorf("Version = %d, want %d", got.Version, reference.VersionNumber)
			}

			// Compare the padded matrix our pipeline actually produces against
			// the library's own bitmap, which carries its built-in border.
			ours, err := qrcode.New(text, qrcode.Medium)
			if err != nil {
				t.Fatalf("qrcode.New: %v", err)
			}
			ours.DisableBorder = true
			got2 := renderQRMatrix(ours, got.Margin)

			if len(got2) != len(want) {
				t.Fatalf("our matrix is %dx%d, want %d", len(got2), len(got2), len(want))
			}
			for y := range want {
				for x := range want[y] {
					if got2[y][x] != want[y][x] {
						t.Fatalf("module (%d,%d) = %v, want %v", x, y, got2[y][x], want[y][x])
					}
				}
			}
		})
	}
}

// Scanners locate a code by its three finder patterns, then read the module
// grid. If the quiet zone is added asymmetrically or the grid gets resampled,
// those patterns smear and real devices reject the code. These tests assert
// the rendered pixels reproduce the module grid exactly.
func TestRenderedPNGFinderPatternsIntact(t *testing.T) {
	res, err := GenerateQRCodeLogic(QRCodeRequest{Text: "finder test", Margin: intPtr(4), Size: intPtr(480)})
	if err != nil {
		t.Fatalf("GenerateQRCodeLogic: %v", err)
	}
	img := decodeResultImage(t, res)

	moduleSize := res.Size / res.Modules
	fg := mustParseColor(t, res.FG)
	bg := mustParseColor(t, res.BG)

	// A finder pattern is a 7x7 ring: dark border, light ring, dark 3x3 core.
	// A finder pattern is a 7x7 square: a solid dark outer ring, a light
	// gap, and a solid dark 3x3 core in the middle.
	isFinderOrigin := func(mx, my int) bool {
		for dy := 0; dy < 7; dy++ {
			for dx := 0; dx < 7; dx++ {
				outerRing := dx == 0 || dx == 6 || dy == 0 || dy == 6
				innerCore := dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4
				wantDark := outerRing || innerCore
				if got := pixelIsDark(t, img, (mx+dx)*moduleSize, (my+dy)*moduleSize, fg, bg); got != wantDark {
					return false
				}
			}
		}
		return true
	}

	// The library always places finder patterns in three corners, offset by
	// the quiet zone.
	off := res.Margin
	last := res.Modules - off - 7
	corners := [][2]int{{off, off}, {last, off}, {off, last}}

	for i, c := range corners {
		if !isFinderOrigin(c[0], c[1]) {
			t.Errorf("finder pattern %d at module (%d,%d) is not intact", i, c[0], c[1])
		}
	}
}

// pixelIsDark samples the top-left pixel of a module and classifies it against
// the exact foreground and background colors.
func pixelIsDark(t *testing.T, img image.Image, x, y int, fg, bg color.RGBA) bool {
	t.Helper()
	r, g, b, _ := img.At(x, y).RGBA()
	c := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0xff}

	dToFg := absDiff(c, fg)
	dToBg := absDiff(c, bg)
	if dToFg > 8 && dToBg > 8 {
		t.Fatalf("pixel at (%d,%d) = %v, which is neither fg %v nor bg %v (antialiased?)", x, y, c, fg, bg)
	}
	// Dark means "closer to the module color".
	return dToFg < dToBg
}

func absDiff(a, b color.RGBA) int {
	return int(absInt(int(a.R)-int(b.R)) + absInt(int(a.G)-int(b.G)) + absInt(int(a.B)-int(b.B)))
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// decodeResultImage decodes the rendered PNG into a pixel buffer.
func decodeResultImage(t *testing.T, res *QRCodeResult) image.Image {
	t.Helper()
	raw := mustDecodeBase64(t, res.RawBase64)
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("decoding PNG: %v", err)
	}
	return img
}

func mustParseColor(t *testing.T, hex string) color.RGBA {
	t.Helper()
	c, ok := tryParseQRColor(hex)
	if !ok {
		t.Fatalf("could not parse color %q", hex)
	}
	return c
}

// The quiet zone must be uniform background on all four sides, otherwise a
// scanner's finder-pattern sweep fails on the leading edge.
func TestRenderedPNGQuietZoneUniform(t *testing.T) {
	for _, margin := range []int{0, 1, 4, 8} {
		res, err := GenerateQRCodeLogic(QRCodeRequest{Text: "quiet zone", Margin: intPtr(margin), Size: intPtr(600)})
		if err != nil {
			t.Fatalf("margin %d: %v", margin, err)
		}
		if margin == 0 {
			continue // nothing to assert beyond the no-margin warning
		}

		img := decodeResultImage(t, res)
		moduleSize := res.Size / res.Modules
		bg := mustParseColor(t, res.BG)
		fg := mustParseColor(t, res.FG)

		for i := 0; i < margin*moduleSize; i++ {
			edges := [][2]int{
				{i, 0}, {i, res.Size - 1},
				{0, i}, {res.Size - 1, i},
			}
			for _, e := range edges {
				if pixelIsDark(t, img, e[0], e[1], fg, bg) {
					t.Fatalf("margin %d: edge pixel (%d,%d) is dark, quiet zone is not clear", margin, e[0], e[1])
				}
			}
		}
	}
}
