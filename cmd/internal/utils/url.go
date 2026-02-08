package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetBaseURL returns the base URL of the request
func GetBaseURL(c *gin.Context) string {
	protocol := "http"
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.Request.Host)
}

// FormatImageURL prepends the base URL and /api prefix to a relative image path
func FormatImageURL(img string, baseURL string) string {
	if img == "" {
		return ""
	}
	if strings.HasPrefix(img, "http") {
		return img
	}
	// Ensure the path starts with /
	if !strings.HasPrefix(img, "/") {
		img = "/" + img
	}
	return baseURL + "/api" + img
}

// FormatImageURLs formats a slice of image paths
func FormatImageURLs(images []string, baseURL string) []string {
	if images == nil {
		return []string{}
	}
	formatted := make([]string, len(images))
	for i, img := range images {
		formatted[i] = FormatImageURL(img, baseURL)
	}
	return formatted
}
