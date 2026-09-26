package handlers

import (
	"html"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTMLEntityRequest binds form data or JSON payloads
type HTMLEntityRequest struct {
	Text string `form:"text" json:"text"`
}

// HTMXEncodeHTMLEntity handles encoding requests from HTMX
func HTMXEncodeHTMLEntity(c *gin.Context) {
	var req HTMLEntityRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "html-entity-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	encoded := html.EscapeString(req.Text)

	c.HTML(http.StatusOK, "html-entity-result", gin.H{
		"Result":    encoded,
		"Original":  req.Text,
		"Operation": "Encode",
	})
}

// V1EncodeHTMLEntity handles JSON REST API requests to encode HTML entities
func V1EncodeHTMLEntity(c *gin.Context) {
	var req HTMLEntityRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	encoded := html.EscapeString(req.Text)

	c.JSON(http.StatusOK, gin.H{
		"result":    encoded,
		"original":  req.Text,
		"operation": "Encode",
	})
}

// HTMXDecodeHTMLEntity handles decoding requests from HTMX
func HTMXDecodeHTMLEntity(c *gin.Context) {
	var req HTMLEntityRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "html-entity-result", gin.H{
			"Error": "Invalid request parameters",
		})
		return
	}

	decoded := html.UnescapeString(req.Text)

	c.HTML(http.StatusOK, "html-entity-result", gin.H{
		"Result":    decoded,
		"Original":  req.Text,
		"Operation": "Decode",
	})
}

// V1DecodeHTMLEntity handles JSON REST API requests to decode HTML entities
func V1DecodeHTMLEntity(c *gin.Context) {
	var req HTMLEntityRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	decoded := html.UnescapeString(req.Text)

	c.JSON(http.StatusOK, gin.H{
		"result":    decoded,
		"original":  req.Text,
		"operation": "Decode",
	})
}
