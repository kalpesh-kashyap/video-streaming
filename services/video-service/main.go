package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/config"
)

func main() {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	config.ConnectDb()
	config.MiggrateDb()

	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/check", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "message from video service"})
	})

	log.Println("Video service is running on http://localhost:3001")
	if err := app.Listen(":3001"); err != nil {
		log.Fatalf("Video service server failed to start, %v", err)
	}
}
