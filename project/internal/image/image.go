package image

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Image interface {
	SaveImage(ctx context.Context, image []byte) error
}

type imageImpl struct {
	minioClient *minio.Client
	bucket      string
}

func Newimage(minioEndpoint, accessKey, secretKey, bucket string) Image {
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}
	return &imageImpl{
		minioClient: minioClient,
		bucket:      bucket}
}

func (n *imageImpl) SaveImage(ctx context.Context, image []byte) error {
	return nil
}
