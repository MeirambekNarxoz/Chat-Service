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

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*MinioClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize MinIO client: %w", err)
	}

	bucketName := "image"
	ctx := context.Background()
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("MinIO bucket %s already exists", bucketName)
		} else {
			return nil, fmt.Errorf("check/create bucket %s: %w", bucketName, err)
		}
	} else {
		log.Printf("MinIO bucket %s created", bucketName)
	}

	// Set public read policy
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": ["s3:GetObject"],
				"Effect": "Allow",
				"Principal": "*",
				"Resource": ["arn:aws:s3:::%s/*"],
				"Sid": ""
			}
		]
	}`, bucketName)
	err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		log.Printf("Error setting bucket policy: %v\n", err)
	}

	return &MinioClient{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
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
