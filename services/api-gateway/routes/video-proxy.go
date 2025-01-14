package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

func VideoProxyRoutes(c *fiber.App) {
	c.Get("/videos/check", func(c *fiber.Ctx) error {
		targetURL := "http://localhost:3001/check"
		return proxy.Do(c, targetURL)
	})
	c.Post("/videos/upload", func(c *fiber.Ctx) error {
		targetURL := "http://localhost:3001/upload"
		return proxy.Do(c, targetURL)
	})
	c.Get("/videos/upload/:videoId", func(c *fiber.Ctx) error {
		targetURL := "http://localhost:3001/upload/" + c.Params("videoId")
		c.Set("Connection", "close")
		return proxy.Do(c, targetURL)
	})
}
