package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// RequestLogger returns a middleware function that logs request details using Fiber logger
func RequestLogger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "${time} | ${ip} | ${status} | ${latency} | ${method} ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Bangkok",
	})
}
