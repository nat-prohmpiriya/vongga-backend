package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type JeagerHandler struct {
}

func NewJeagerHandler(router fiber.Router) *JeagerHandler {
	handler := &JeagerHandler{}

	router.Find("/", handler.ProxyJaegerHandler)

	return handler

}

func (h *JeagerHandler) ProxyJaegerHandler(c *fiber.Ctx) error {
	// Build target URL for Jaeger
	originalUrl := c.OriginalURL()
	trimmedPath := strings.TrimPrefix(originalUrl, "/jaeger")
	// Add /api prefix for Jaeger API
	jaegerURL := fmt.Sprintf("http://jaeger:16686/api%s", trimmedPath)

	// Create HTTP request
	req, err := http.NewRequest(c.Method(), jaegerURL, bytes.NewReader(c.Body()))
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Copy headers
	for key, values := range c.FindReqHeaders() {
		if len(values) > 0 {
			req.Header.Set(key, values[0]) // ใช้ค่าแรกจาก array
		}
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Set status code and return response body
	c.Status(resp.StatusCode)
	return c.Send(body)
}
