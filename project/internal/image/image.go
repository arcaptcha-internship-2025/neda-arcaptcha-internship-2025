package image

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Image interface {
	SaveImage(ctx context.Context, image []byte, filename string) (string, error)
	GetImageURL(ctx context.Context, filename string) (string, error)
	DeleteImage(ctx context.Context, filename string) error
}

type imageImpl struct {
	minioClient *minio.Client
	bucket      string
	endpoint    string
}

func NewImage(minioEndpoint, accessKey, secretKey, bucket string) Image {
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	//creating bucket if not exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil || !exists {
		err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	}

	return &imageImpl{
		minioClient: minioClient,
		bucket:      bucket,
		endpoint:    minioEndpoint,
	}
}

func (i *imageImpl) SaveImage(ctx context.Context, image []byte, filename string) (string, error) {
	if len(image) > 10*1024*1024 {
		return "", fmt.Errorf("image size exceeds 10MB limit")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
	}

	if !allowedExts[ext] {
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}

	//unique filename with timestamp
	uniqueFilename := fmt.Sprintf("bills/%d_%s", time.Now().Unix(), filename)

	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".pdf":
		contentType = "application/pdf"
	}

	//uploading the file
	_, err := i.minioClient.PutObject(
		ctx,
		i.bucket,
		uniqueFilename,
		bytes.NewReader(image),
		int64(len(image)),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	//returning the object key (not full URL)
	return uniqueFilename, nil
}

func (i *imageImpl) GetImageURL(ctx context.Context, objectKey string) (string, error) {
	if objectKey == "" {
		return "", nil
	}

	//generating presigned URL for secure access (expires in 24 hours)
	presignedURL, err := i.minioClient.PresignedGetObject(
		ctx,
		i.bucket,
		objectKey,
		24*time.Hour,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func (i *imageImpl) DeleteImage(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return nil
	}

	err := i.minioClient.RemoveObject(ctx, i.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}
