package config

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client holds an AWS S3 client instance
var S3Client *s3.Client

// InitAWS initializes AWS SDK and S3 client with the correct region
func InitAWS() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1")) // Force `us-east-1`
	if err != nil {
		log.Fatalf("❌ Failed to load AWS config: %v", err)
	}
	S3Client = s3.NewFromConfig(cfg)
	log.Println("✅ AWS S3 client initialized in region: us-east-1")
}
