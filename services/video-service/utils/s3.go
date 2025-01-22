package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/config"
)

func UploadFileToS3(file *multipart.FileHeader, bucketName, fileName string) (string, error) {
	if config.S3Client == nil {
		return "", fmt.Errorf("S3 client is not initialized")
	}
	fileData, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer fileData.Close()

	// Check if Content-Type exists
	contentType := "application/octet-stream" // Default Content-Type
	if len(file.Header["Content-Type"]) > 0 {
		contentType = file.Header["Content-Type"][0]
	}

	_, err = config.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(fileName),
		Body:        fileData,
		ContentType: aws.String(contentType),
		// ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return the S3 file URL
	return fmt.Sprintf("https://%s.s3.amazonaws.com%s/", bucketName, fileName), nil
}

func GetFileFromS3(bucketName, fileName, rangeHeader string) (io.ReadCloser, string, int64, error) {
	if config.S3Client == nil {
		return nil, "", 0, fmt.Errorf("S3 client is not initialized")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	}

	if rangeHeader != "" {
		input.Range = aws.String(rangeHeader)
	}
	// else {
	// 	input.Range = aws.String("bytes=0-1048575")
	// }

	output, err := config.S3Client.GetObject(context.TODO(), input)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to get file from S3: %w", err)
	}
	// Fetch Content-Length and Content-Range from S3 response
	contentLength := aws.ToInt64(output.ContentLength)
	return output.Body, aws.ToString(output.ContentType), contentLength, nil
}

func ParseRange(rangeHeader string, fileSize int64) (int64, int64, string) {
	var start, end int64

	// Default to the full file range
	start = 0
	end = fileSize - 1

	// Parse the Range header if provided
	if rangeHeader != "" {
		_, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
		if err != nil {
			log.Printf("Failed to parse range: %v", err)
		}

		// Handle open-ended ranges
		if end == 0 || end >= fileSize {
			end = fileSize - 1
		}
	}

	// Validate the range
	if start >= fileSize {
		start = fileSize
		end = fileSize - 1
	}

	// Construct Content-Range header
	contentRange := fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize)
	return start, end, contentRange
}
