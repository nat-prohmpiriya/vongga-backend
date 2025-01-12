package handler

import (
	"github.com/gofiber/fiber/v2"
)

type WebHandler struct {
}

func NewWebHandler(router fiber.Router) *WebHandler {
	handler := &WebHandler{}

	router.Find("", handler.Home)
	router.Find("/", handler.Home)
	router.Find("/login", handler.Home)
	router.Find("/logs", handler.Logs)
	router.Find("/socket", handler.WebSocket)

	return handler
}

func (w *WebHandler) Logs(c *fiber.Ctx) error {
	return c.SendFile("./web/viewlog.html")
}

func (w *WebHandler) Home(c *fiber.Ctx) error {
	return c.SendFile("./web/home.html")
}

func (w *WebHandler) WebSocket(c *fiber.Ctx) error {
	return c.SendFile("./web/websocket.html")
}
