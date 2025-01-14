package utils

import (
	"context"
	"fmt"
	"io"
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
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, fileName), nil
}

func GetFileFromS3(bucketName, fileName, rangeHeader string) (io.ReadCloser, string, error) {
	if config.S3Client == nil {
		return nil, "", fmt.Errorf("S3 client is not initialized")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	}

	if rangeHeader != "" {
		input.Range = aws.String(rangeHeader)
	} else {
		input.Range = aws.String("bytes=0-1048575")
	}

	output, err := config.S3Client.GetObject(context.TODO(), input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file from S3: %w", err)
	}
	return output.Body, aws.ToString(output.ContentType), nil
}
