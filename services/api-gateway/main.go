package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kalpesh-kashyap/video-streaming/services/api-gateway/routes"
)

func main() {
	app := fiber.New()
	routes.AppRouter(app)
	app.Use(logger.New())
	app.Use(cors.New())
	log.Println("API Gateway is running on http://localhost:3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error in starting serve %v", err)
	}
}
