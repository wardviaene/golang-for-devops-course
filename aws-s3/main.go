package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const s3BucketName = "go-aws-test-xz9"

func main() {
	var (
		s3Client *s3.Client
		err      error
	)
	if s3Client, err = initS3Client(context.Background(), "us-east-1"); err != nil {
		fmt.Printf("initConfig error: %s", err)
		os.Exit(1)
	}
	if err = createS3Bucket(context.Background(), s3Client); err != nil {
		fmt.Printf("createS3Bucket error: %s", err)
		os.Exit(1)
	}
	if err = uploadFileToS3(context.Background(), s3Client); err != nil {
		fmt.Printf("uploadFileToS3 error: %s", err)
		os.Exit(1)
	}
}

func initS3Client(ctx context.Context, region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("Config error: %s", err)
	}
	return s3.NewFromConfig(cfg), nil
}

func createS3Bucket(ctx context.Context, s3Client *s3.Client) error {

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s3BucketName),
	})
	if err != nil {
		return fmt.Errorf("CreateBucket error: %s", err)
	}
	return nil
}

func uploadFileToS3(ctx context.Context, s3Client *s3.Client) error {
	uploader := manager.NewUploader(s3Client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String("test.txt"),
		Body:   strings.NewReader("this is a test"),
	})
	if err != nil {
		return fmt.Errorf("Upload error: %s", err)
	}
	return nil
}
