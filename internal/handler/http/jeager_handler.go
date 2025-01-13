package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type JeagerHandler struct {
}

func NewJeagerHandler(router fiber.Router) *JeagerHandler {
	handler := &JeagerHandler{}

	router.Get("", handler.Ping)
	router.Get("/traces", handler.ProxyJaegerHandler)
	router.Get("/services", handler.GetServices)
	return handler

}

func (h *JeagerHandler) Ping(c *fiber.Ctx) error {
	return c.Status(200).SendString("ok")
}

func (h *JeagerHandler) ProxyJaegerHandler(c *fiber.Ctx) error {
	// Build target URL for Jaeger
	originalUrl := c.OriginalURL()
	trimmedPath := strings.TrimPrefix(originalUrl, "/api/jaeger")
	// Add /api prefix for Jaeger API
	jaegerURL := fmt.Sprintf("http://jaeger:16686/api%s", trimmedPath)

	// Create HTTP request
	req, err := http.NewRequest(c.Method(), jaegerURL, bytes.NewReader(c.Body()))
	if err != nil {
		log.Default().Printf("Jaeger URL: %s", trimmedPath)
		return err
	}

	// Copy headers
	for key, values := range c.GetReqHeaders() {
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

func (h *JeagerHandler) GetServices(c *fiber.Ctx) error {
	jaegerURL := "http://jaeger:16686/api/services"

	// Create HTTP request
	req, err := http.NewRequest("GET", jaegerURL, nil)
	if err != nil {
		return err
	}

	// Forward the request to Jaeger
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Set status code and return response body
	c.Status(resp.StatusCode)
	return c.Send(body)
}
