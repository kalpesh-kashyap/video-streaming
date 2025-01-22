package handler

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/config"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/models"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/utils"
)

type VideoTask struct {
	metaData models.Video
	File     *multipart.FileHeader
}

func uploadWorker(id int, tasks chan VideoTask) {
	log.Printf("Worker with id is now working %d", id)
	for task := range tasks {
		saveProcess(task, id)
	}
	log.Printf("Worker with ID %d has stopped", id)
}

func UploadFile(c *fiber.Ctx) error {
	uploadQueue := make(chan VideoTask, 100)
	maxWorkers := 3

	var metadata models.Video
	metadata.Title = c.FormValue("title")
	metadata.Description = c.FormValue("description")

	var wg sync.WaitGroup

	allFiles, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "Failed to get files"})
	}

	for i := 1; i <= maxWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			uploadWorker(workerId, uploadQueue)
		}(i)
	}

	go func() {
		for {
			time.Sleep(time.Second * 10)
			if len(uploadQueue) > cap(uploadQueue)/2 && maxWorkers < 10 {
				maxWorkers++
				wg.Add(1)
				log.Printf("new worker has been assign with id %d", maxWorkers)
				go func(workerId int) {
					defer wg.Done()
					uploadWorker(workerId, uploadQueue)
				}(maxWorkers)
			}
		}
	}()

	go func() {
		defer close(uploadQueue)
		sendTaskToQueue(allFiles, uploadQueue, &metadata)
	}()

	wg.Wait()

	return c.JSON(fiber.Map{"message": "Video metadata saved successfully"})
}

func sendTaskToQueue(allVideos *multipart.Form, videos chan VideoTask, metaData *models.Video) {
	for index, file := range allVideos.File["videos"] {
		uniqueFileName := fmt.Sprintf("videos/%d-%s", time.Now().UnixNano(), file.Filename)
		fileData := VideoTask{
			metaData: models.Video{
				Title:       fmt.Sprintf("%s%d", metaData.Title, index+1),
				Description: fmt.Sprintf("%s+%d", metaData.Description, index+1),
				Size:        int(file.Size),
				Filename:    uniqueFileName,
			},
			File: file,
		}
		videos <- fileData
	}
}

func saveProcess(video VideoTask, id int) {
	uploadedFiles := make(chan string, 10)
	log.Printf("Id:%d has start process with video%s", id, video.metaData.Title)
	var wg sync.WaitGroup
	wg.Add(2)
	// Upload file to S3
	go func() {
		defer wg.Done()
		uploadFileToS3(video, uploadedFiles, id)
	}()

	// Save to database
	go func() {
		defer wg.Done()
		saveToDatabase(id, uploadedFiles, video.metaData)
	}()

	wg.Wait()
	close(uploadedFiles)

}

func uploadFileToS3(video VideoTask, urlData chan string, id int) {
	fileName := video.metaData.Filename
	file := video.File
	url, err := utils.UploadFileToS3(file, "video-stream-go", fileName)
	if err != nil {
		// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to uplaod file" + err.Error()})
		log.Printf("Error in uploading file with worker Id %d", id)
		return
	}
	log.Printf("Video has been successfully uploaded for %s with id:%d", fileName, id)
	urlData <- url
}

func saveToDatabase(id int, uploadedUrl chan string, videoData models.Video) {
	urlData, ok := <-uploadedUrl
	if !ok || urlData == "" {
		log.Printf("fail to save because URL no availabe to save")
		return
	}
	videoData.URL = urlData

	if err := config.DB.Create(&videoData).Error; err != nil {
		log.Printf("failed to save video to database with worker:%d", id)
		return
	}
	log.Printf("Video has saved successfully to database for %s with id:%d", videoData.Title, id)

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
	log.Printf("Client requested range: %s", rangeHeader)

	// Fetch the file from S3 with the Range header
	file, contentType, fileSize, err := utils.GetFileFromS3("video-stream-go", video.Filename, rangeHeader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch video file: %s", err.Error()),
		})
	}
	defer file.Close()

	// Parse the Range header
	start, end, contentRange := utils.ParseRange(rangeHeader, fileSize)
	log.Printf("Parsed Range: Start=%d, End=%d, ContentRange=%s", start, end, contentRange)

	c.Status(fiber.StatusPartialContent)
	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderAcceptRanges, "bytes")
	c.Set(fiber.HeaderContentRange, contentRange)
	c.Set(fiber.HeaderContentLength, fmt.Sprintf("%d", end-start+1))

	buf := make([]byte, end-start+1)
	if _, err := io.ReadFull(file, buf); err != nil {
		log.Printf("Error reading file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to stream video",
		})
	}
	if _, err := c.Write(buf); err != nil {
		log.Printf("Error during streaming: %v", err)
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
