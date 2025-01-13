package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/handler"
)

func FileUploadRoutes(c *fiber.App) {
	video := c.Group("upload")
	video.Post("/", handler.UploadFile)
}
