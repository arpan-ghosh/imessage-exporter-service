package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"imessage-exporter-webservice/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadToS3 uploads a single file to S3
func UploadToS3(filePath, fileName string) (string, error) {
	// Ensure S3Client is initialized
	if config.S3Client == nil {
		log.Println("❌ ERROR: S3Client is not initialized")
		return "", fmt.Errorf("S3 client is not initialized")
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("❌ ERROR: Unable to open file %s: %v\n", filePath, err)
		return "", err
	}
	defer file.Close()

	// Ensure bucket name is set
	bucketName := config.GetEnv("AWS_S3_BUCKET", "")
	if bucketName == "" {
		log.Println("❌ ERROR: AWS_S3_BUCKET environment variable is not set")
		return "", fmt.Errorf("AWS_S3_BUCKET is not set")
	}

	s3Key := fmt.Sprintf("exports/%s", fileName)
	log.Printf("📤 Uploading file to S3: bucket=%s, key=%s\n", bucketName, s3Key)

	// Read file contents
	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(file)
	if err != nil {
		log.Printf("❌ ERROR: Failed to read file content: %v\n", err)
		return "", err
	}

	// Upload to S3
	_, err = config.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
		Body:   buffer,
	})
	if err != nil {
		log.Printf("❌ ERROR: S3 upload failed for %s: %v\n", s3Key, err)
		return "", err
	}

	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, s3Key)
	log.Printf("✅ Successfully uploaded to S3: %s\n", s3URL)

	return s3URL, nil
}

// UploadFolderToS3 uploads all files in a folder to S3
func UploadFolderToS3(folderPath string) ([]string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var uploadedFiles []string
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(folderPath, file.Name())
			s3URL, err := UploadToS3(filePath, file.Name())
			if err != nil {
				log.Printf("⚠️ Failed to upload file: %s", filePath)
				continue
			}
			uploadedFiles = append(uploadedFiles, s3URL)
		}
	}

	return uploadedFiles, nil
}

// DownloadFromS3 downloads a file from S3 and saves it locally
func DownloadFromS3(s3URL, destinationPath string) error {
	parsedURL, err := url.Parse(s3URL)
	if err != nil {
		return fmt.Errorf("invalid S3 URL: %v", err)
	}

	// Extract bucket and key from the S3 URL
	bucketName := strings.Split(parsedURL.Host, ".")[0]
	objectKey := strings.TrimPrefix(parsedURL.Path, "/")

	// Download the file
	outputFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	resp, err := config.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = outputFile.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("✅ Downloaded file from S3: %s", destinationPath)
	return nil
}
