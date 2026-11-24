package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBuildWhatsAppLink(t *testing.T) {
	tests := []struct {
		name          string
		phone         string
		message       string
		wantPhone     string
		wantFormatted string
		wantMessage   string
		wantLink      string
		wantDeepLink  string
		wantError     string
	}{
		{
			name:          "international number with separators",
			phone:         " +1 (415) 555-0132 ",
			wantPhone:     "14155550132",
			wantFormatted: "+14155550132",
			wantLink:      "https://wa.me/14155550132",
			wantDeepLink:  "whatsapp://send?phone=14155550132",
		},
		{
			name:          "leading international 00 prefix",
			phone:         "0044 20.7946.0958",
			wantPhone:     "442079460958",
			wantFormatted: "+442079460958",
			wantLink:      "https://wa.me/442079460958",
			wantDeepLink:  "whatsapp://send?phone=442079460958",
		},
		{
			name:          "message with spaces newline and emoji",
			phone:         "+1 415 555 0132",
			message:       "  Hello there \n👋  ",
			wantPhone:     "14155550132",
			wantFormatted: "+14155550132",
			wantMessage:   "Hello there \n👋",
			wantLink:      "https://wa.me/14155550132?text=Hello%20there%20%0A%F0%9F%91%8B",
			wantDeepLink:  "whatsapp://send?phone=14155550132&text=Hello%20there%20%0A%F0%9F%91%8B",
		},
		{
			name:          "empty message omits text parameter",
			phone:         "+1 415 555 0132",
			message:       " \n\t ",
			wantPhone:     "14155550132",
			wantFormatted: "+14155550132",
			wantLink:      "https://wa.me/14155550132",
			wantDeepLink:  "whatsapp://send?phone=14155550132",
		},
		{
			name:      "too short",
			phone:     "+123456",
			wantError: "7 to 15 digits",
		},
		{
			name:      "too long",
			phone:     "+1234567890123456",
			wantError: "7 to 15 digits",
		},
		{
			name:      "letters and invalid characters",
			phone:     "+1 415 CALL!",
			wantError: "only digits",
		},
		{
			name:      "empty phone",
			phone:     " \t ",
			wantError: "country code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildWhatsAppLink(tt.phone, tt.message)
			if tt.wantError != "" {
				if err == nil {
					t.Fatal("expected an error")
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %q, want it to contain %q", err, tt.wantError)
				}
				if result.Error != err.Error() {
					t.Errorf("result.Error = %q, want %q", result.Error, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildWhatsAppLink: %v", err)
			}

			if result.Phone != tt.wantPhone {
				t.Errorf("Phone = %q, want %q", result.Phone, tt.wantPhone)
			}
			if result.FormattedPhone != tt.wantFormatted {
				t.Errorf("FormattedPhone = %q, want %q", result.FormattedPhone, tt.wantFormatted)
			}
			if result.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", result.Message, tt.wantMessage)
			}
			if result.Link != tt.wantLink {
				t.Errorf("Link = %q, want %q", result.Link, tt.wantLink)
			}
			if result.DeepLink != tt.wantDeepLink {
				t.Errorf("DeepLink = %q, want %q", result.DeepLink, tt.wantDeepLink)
			}
			if result.HasMessage != (tt.wantMessage != "") {
				t.Errorf("HasMessage = %t, want %t", result.HasMessage, tt.wantMessage != "")
			}
			if result.Error != "" {
				t.Errorf("Error = %q, want empty", result.Error)
			}
		})
	}
}

func TestHTMXGenerateWhatsAppLink(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.SetFuncMap(TemplateFuncs)
	router.LoadHTMLGlob("../templates/*")
	router.POST("/api/htmx/whatsapp/link", HTMXGenerateWhatsAppLink)

	form := url.Values{
		"phone":   {"+1 415 555 0132"},
		"message": {"Hi there"},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/htmx/whatsapp/link", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	for _, expected := range []string{
		"https://wa.me/14155550132?text=Hi%20there",
		"&#43;14155550132",
		"Open WhatsApp",
	} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Errorf("response does not contain %q: %s", expected, rec.Body.String())
		}
	}
}

func TestHTMXGenerateWhatsAppLinkRendersValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.SetFuncMap(TemplateFuncs)
	router.LoadHTMLGlob("../templates/*")
	router.POST("/api/htmx/whatsapp/link", HTMXGenerateWhatsAppLink)

	form := url.Values{"phone": {"123"}}
	req := httptest.NewRequest(http.MethodPost, "/api/htmx/whatsapp/link", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "alert-error") || !strings.Contains(rec.Body.String(), "7 to 15 digits") {
		t.Fatalf("expected rendered validation alert, got: %s", rec.Body.String())
	}
}
