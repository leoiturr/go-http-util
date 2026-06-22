package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image"
	"image/draw"
	"image/png"
	"net/http"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
)

// QRCodeRequest binds form data for generating a QR Code
type QRCodeRequest struct {
	Text string `form:"text" json:"text"`
}

// GenerateQRCodeLogic performs core QR code generation, returning the data URL and raw base64 encoded string
func GenerateQRCodeLogic(text string) (string, string, error) {
	if text == "" {
		return "", "", fmt.Errorf("input text cannot be empty")
	}

	// Generate QR code structure
	q, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize QR code: %v", err)
	}

	// Retrieve paletted image representation (300x300 size)
	palettedImg := q.Image(300)

	// Convert paletted image to a standard 32-bit RGBA image
	bounds := palettedImg.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	draw.Draw(rgbaImg, bounds, palettedImg, bounds.Min, draw.Src)

	// Encode standard RGBA image to PNG bytes
	var buf bytes.Buffer
	err = png.Encode(&buf, rgbaImg)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode QR code PNG: %v", err)
	}

	// Encode to base64 Data URL
	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Img)

	return dataURL, base64Img, nil
}

// HTMXGenerateQRCode handles the QR code generation request from HTMX
func HTMXGenerateQRCode(c *gin.Context) {
	var req QRCodeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "qrcode-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	dataURL, base64Img, err := GenerateQRCodeLogic(req.Text)
	if err != nil {
		c.HTML(http.StatusOK, "qrcode-result", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "qrcode-result", gin.H{
		"ImageURL":  template.URL(dataURL),
		"Text":      req.Text,
		"RawBase64": base64Img,
	})
}

// V1GenerateQRCode handles JSON REST API requests to generate a QR code
func V1GenerateQRCode(c *gin.Context) {
	var req QRCodeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	dataURL, base64Img, err := GenerateQRCodeLogic(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_url":  dataURL,
		"text":       req.Text,
		"raw_base64": base64Img,
	})
}
