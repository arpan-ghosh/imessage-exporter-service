package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"imessage-exporter-webservice/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadToS3 uploads a single file to S3, storing it under "exports/<RequesterName>/..."
func UploadToS3(filePath, fileName, requesterName string) (string, error) {
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

	// Sanitize the requester name for safe file paths
	sanitizedRequester := sanitizeFileName(requesterName)

	// Store files under a requester-specific folder
	s3Key := fmt.Sprintf("exports/%s/%s", sanitizedRequester, fileName)
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

// UploadFolderToS3 uploads all files in a folder to S3 under "exports/<RequesterName>/"
func UploadFolderToS3(folderPath, requesterName string) ([]string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var uploadedFiles []string
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(folderPath, file.Name())
			s3URL, err := UploadToS3(filePath, file.Name(), requesterName)
			if err != nil {
				log.Printf("⚠️ Failed to upload file: %s", filePath)
				continue
			}
			uploadedFiles = append(uploadedFiles, s3URL)
		}
	}

	return uploadedFiles, nil
}

// DownloadFromS3 downloads a file from S3 into "downloads/<RequesterName>/chat.db"
func DownloadFromS3(s3URL, requesterName string) (string, error) {
	// Define correct download path
	localDir := filepath.Join("/var/imessage/downloads", sanitizeFileName(requesterName))
	err := os.MkdirAll(localDir, 0777) // Ensure the directory exists
	if err != nil {
		log.Printf("❌ ERROR: Failed to create directory: %s - %v", localDir, err)
		return "", err
	}

	localPath := filepath.Join(localDir, "chat.db")

	// Extract S3 Key from URL
	s3Key := strings.TrimPrefix(s3URL, fmt.Sprintf("https://%s.s3.amazonaws.com/", config.GetEnv("AWS_S3_BUCKET", "")))
	log.Printf("⬇️ Downloading %s from S3 to %s", s3Key, localPath)

	// Download file from S3
	input := &s3.GetObjectInput{
		Bucket: aws.String(config.GetEnv("AWS_S3_BUCKET", "")),
		Key:    aws.String(s3Key),
	}

	result, err := config.S3Client.GetObject(context.TODO(), input)
	if err != nil {
		log.Printf("❌ ERROR: Failed to download %s: %v", s3Key, err)
		return "", err
	}
	defer result.Body.Close()

	outFile, err := os.Create(localPath)
	if err != nil {
		log.Printf("❌ ERROR: Could not create file: %s - %v", localPath, err)
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, result.Body)
	if err != nil {
		log.Printf("❌ ERROR: Failed to write file: %s - %v", localPath, err)
		return "", err
	} else {
		log.Printf("✅ Successfully downloaded %s to %s", s3Key, localPath)
	}

	return localPath, nil
}



// sanitizeFileName ensures safe filenames by removing spaces and special characters
func sanitizeFileName(name string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "_")
}
