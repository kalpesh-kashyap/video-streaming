package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/kalpesh-kashyap/video-streaming/services/api-gateway/routes"
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit:         1024 * 1024 * 1024,
		StreamRequestBody: true,
	})
	routes.AppRouter(app)
	app.Use(logger.New())
	app.Use(cors.New())

	// health check route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "API Gateway is running"})
	})

	// proxy route video service
	app.Get("/videos/*", func(c *fiber.Ctx) error {
		originalPath := c.Params("*")
		targetURL := "http://localhost:3001/" + originalPath
		return proxy.Do(c, targetURL)
	})

	log.Println("API Gateway is running on http://localhost:3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error in starting serve %v", err)
	}
}
