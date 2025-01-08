package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	log.Println("Video service is running on http://localhost:3001")
	if err := app.Listen(":3001"); err != nil {
		log.Fatalf("Video service server failed to start, %v", err)
	}
}
