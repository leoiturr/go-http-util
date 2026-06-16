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

// GenerateQRCode handles the QR code generation request
func GenerateQRCode(c *gin.Context) {
	var req QRCodeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "qrcode-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	if req.Text == "" {
		c.HTML(http.StatusOK, "qrcode-result", gin.H{
			"Error": "Input text cannot be empty",
		})
		return
	}

	// Generate QR code structure
	q, err := qrcode.New(req.Text, qrcode.Medium)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "qrcode-result", gin.H{
			"Error": fmt.Sprintf("Failed to initialize QR code: %v", err),
		})
		return
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
		c.HTML(http.StatusInternalServerError, "qrcode-result", gin.H{
			"Error": fmt.Sprintf("Failed to encode QR code PNG: %v", err),
		})
		return
	}

	// Encode to base64 Data URL
	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Img)

	c.HTML(http.StatusOK, "qrcode-result", gin.H{
		"ImageURL":  template.URL(dataURL),
		"Text":      req.Text,
		"RawBase64": base64Img,
	})
}
