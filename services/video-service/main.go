package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/config"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/routes"
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit:         1024 * 1024 * 1024,
		StreamRequestBody: true,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
	})

	app.Use(logger.New())
	app.Use(cors.New())

	config.ConnectDb()
	config.MiggrateDb()
	if err := config.InitS3Client(); err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	routes.FileUploadRoutes(app)

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
