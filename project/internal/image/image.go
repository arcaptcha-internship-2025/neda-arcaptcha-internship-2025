package image

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Image interface {
	SaveImage(ctx context.Context, image []byte, filename string) (string, error)
}

type imageImpl struct {
	minioClient *minio.Client
	bucket      string
}

func NewImage(minioEndpoint, accessKey, secretKey, bucket string) Image {
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
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
	}
}

func (i *imageImpl) SaveImage(ctx context.Context, image []byte, filename string) (string, error) {
	//unique filename with timestamp
	uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), filename)

	//uploading the file
	_, err := i.minioClient.PutObject(
		ctx,
		i.bucket,
		uniqueFilename,
		bytes.NewReader(image),
		int64(len(image)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	//URL/path to the stored image
	return fmt.Sprintf("/%s/%s", i.bucket, uniqueFilename), nil
}
