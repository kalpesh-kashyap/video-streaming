package handlers

import "github.com/gofiber/fiber/v2"

func UploadVideo(c *fiber.Ctx) error {
	file, err := c.FormFile("video")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "faile to get uploaded file"})
	}
	data, err := file.Open()

}
