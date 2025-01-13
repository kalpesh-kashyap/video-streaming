package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/config"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/models"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/utils"
)

func UploadFile(c *fiber.Ctx) error {
	var video models.Video
	if err := c.BodyParser(&video); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}
	file, err := c.FormFile("video")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to get file" + err.Error()})
	}

	fileName := fmt.Sprintf("/videos/%d-%s", time.Now().Unix(), file.Filename)
	fileSize := file.Size

	url, err := utils.UploadFileToS3(file, "", fileName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to uplaod file" + err.Error()})
	}
	video.Size = int(fileSize)
	video.URL = url
	if err = config.DB.Create(&video).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save video metadata to the database",
		})
	}
	return c.JSON(fiber.Map{
		"message": "Video metadata saved successfully",
		"video":   video,
	})
}
