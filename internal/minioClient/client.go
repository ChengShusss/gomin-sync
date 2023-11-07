package minioClient

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	client *minio.Client
)

func getClient() (*minio.Client, error) {
	endpoint := os.Getenv("GOMIN_ENDPOINT")
	accessKeyID := os.Getenv("GOMIN_ACCESS_USER")
	secretAccessKey := os.Getenv("GOMIN_ACCESS_PASSWORD")
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		fmt.Printf("Please Set Environment For GOMIN_ENDPOINT, GOMIN_ACCESS_USER and GOMIN_ACCESS_PASSWORD")
		os.Exit(1)
	}
	useSSL := true

	// Initialize minio client object.
	var err error
	client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		client = nil
	}

	return client, err
}

func GetClient() (*minio.Client, error) {
	if client != nil {
		return client, nil
	}

	return getClient()
}

// Upload upload file to remote minio bucket
func Upload(bucketName, filePath, remotePath string) (int64, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "text/plain"
	}

	client, err := GetClient()
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	fmt.Printf("RemotePath: %v\n", remotePath)
	info, err := client.FPutObject(
		ctx, bucketName, remotePath, filePath,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return 0, err
	}
	return info.Size, err
}
