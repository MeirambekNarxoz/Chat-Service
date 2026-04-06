package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) *MinioClient {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	bucketName := "chat-media"
	ctx := context.Background()
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalf("Error checking bucket: %v", err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	return &MinioClient{
		client:     minioClient,
		bucketName: bucketName,
	}
}

func (m *MinioClient) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fileName := file.Filename
	// You might want to generate a unique UUID for filename here to prevent overrides
	info, err := m.client.PutObject(ctx, m.bucketName, fileName, src, file.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/%s/%s", m.bucketName, info.Key), nil
}
