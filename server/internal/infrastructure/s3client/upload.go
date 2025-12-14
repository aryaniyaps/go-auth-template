package s3client

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// GenerateUniqueFilename generates a unique filename for S3 upload
func GenerateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	uniqueFilename := fmt.Sprintf("%d%s", timestamp, ext)
	return uniqueFilename
}

// uploadToS3 uploads file content to S3 and returns the URL
func UploadToS3(ctx context.Context, s3Client *s3.Client, bucketName string, filename string, fileBytes []byte, contentType string) (string, error) {
	// Create PutObject input
	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead, // Make the file publicly accessible
	}

	// Upload file
	_, err := s3Client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct URL
	avatarURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, filename)

	return avatarURL, nil
}
