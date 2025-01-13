package handler

import (
	"github.com/gofiber/fiber/v2"
)

type WebHandler struct {
}

func NewWebHandler(router fiber.Router) *WebHandler {
	handler := &WebHandler{}

	router.Get("", handler.Home)
	router.Get("", handler.Home)
	router.Get("/login", handler.Home)
	router.Get("/logs", handler.Logs)
	router.Get("/socket", handler.WebSocket)

	return handler
}

func (w *WebHandler) Home(c *fiber.Ctx) error {
	return c.SendFile("web/home.html")
}

func (w *WebHandler) Logs(c *fiber.Ctx) error {
	return c.SendFile("web/viewlog.html")
}

func (w *WebHandler) WebSocket(c *fiber.Ctx) error {
	return c.SendFile("web/websocket.html")
}
