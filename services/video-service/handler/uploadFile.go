package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/config"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/models"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/utils"
)

func UploadFile(c *fiber.Ctx) error {
	fmt.Println("inside the upload handler")
	var video models.Video
	video.Title = c.FormValue("title")
	video.Description = c.FormValue("description")
	file, err := c.FormFile("video")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed to get file" + err.Error()})
	}

	fileName := fmt.Sprintf("videos/%d-%s", time.Now().Unix(), file.Filename)
	fileSize := file.Size

	url, err := utils.UploadFileToS3(file, "video-stream-go", fileName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to uplaod file" + err.Error()})
	}
	video.Size = int(fileSize)
	video.URL = url
	video.Filename = fileName
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

func StreamVideo(c *fiber.Ctx) error {
	id := c.Params("videoId")

	// Check if video exists in the database
	var video models.Video
	if err := config.DB.Where("id = ?", id).First(&video).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Video not found"})
	}

	// Extract Range header from the client request
	rangeHeader := c.Get(fiber.HeaderRange)

	// Fetch the file from S3 with the Range header
	file, contentType, err := utils.GetFileFromS3("video-stream-go", video.Filename, rangeHeader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch video file: %s", err.Error()),
		})
	}
	defer file.Close()

	c.Status(fiber.StatusPartialContent)
	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderAcceptRanges, "bytes")

	// Stream S3 response directly to the client
	if _, err := io.Copy(c, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to stream video",
		})
	}
	return nil
}

// func StreamVideo(c *fiber.Ctx) error {
// 	id := c.Params("videoId")

// 	// Check if video exists in the database
// 	var video models.Video
// 	if err := config.DB.Where("id = ?", id).First(&video).Error; err != nil {
// 		log.Printf("Video not found for ID: %s", id)
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Video not found"})
// 	}

// 	rangeHeader := c.Get(fiber.HeaderRange)

// 	// Fetch the file from S3
// 	file, contentType, err := utils.GetFileFromS3("video-stream-go", video.Filename, rangeHeader)
// 	if err != nil {
// 		log.Printf("Failed to fetch file from S3: %s", err)
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": fmt.Sprintf("Failed to fetch video file: %s", err.Error()),
// 		})
// 	}
// 	defer file.Close()

// 	c.Status(fiber.StatusPartialContent)
// 	c.Set(fiber.HeaderContentType, contentType)
// 	c.Set(fiber.HeaderAcceptRanges, "bytes")

// 	if err := c.SendStream(file); err != nil {
// 		log.Printf("Failed to stream video: %s", err)
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to stream video file: " + err.Error(),
// 		})
// 	}

// 	// // Set response headers
// 	// c.Set(fiber.HeaderContentType, contentType)
// 	// if video.Size > 0 {
// 	// 	c.Set(fiber.HeaderContentLength, fmt.Sprintf("%d", video.Size))
// 	// }

// 	// err = c.SendStream(file)
// 	// if err != nil {
// 	// 	log.Printf("Failed to stream video: %s", err)
// 	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 	// 		"error": "Failed to stream video file: " + err.Error(),
// 	// 	})
// 	// }

// 	// buffer := make([]byte, 1024*1024) // 1MB buffer
// 	// for {
// 	// 	n, err := file.Read(buffer)
// 	// 	if err != nil {
// 	// 		if err.Error() == "EOF" {
// 	// 			break
// 	// 		}
// 	// 		log.Printf("Error reading from stream: %s", err)
// 	// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 	// 			"error": "Failed to read video stream: " + err.Error(),
// 	// 		})
// 	// 	}

// 	// 	log.Printf("Streaming %d bytes of data...", n)
// 	// 	if _, writeErr := c.Write(buffer[:n]); writeErr != nil {
// 	// 		log.Printf("Error writing to response: %s", writeErr)
// 	// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 	// 			"error": "Failed to write video stream: " + writeErr.Error(),
// 	// 		})
// 	// 	}
// 	// }

// 	// log.Printf("Successfully streamed video with ID: %s", id)
// 	return nil
// }
