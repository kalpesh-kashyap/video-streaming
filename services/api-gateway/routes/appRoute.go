package routes

import "github.com/gofiber/fiber/v2"

func AppRouter(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the API Gateway!")
	})
}
