package minioClient

import (
	"context"
	"errors"
	"gomin-sync/internal/config"
	"mime"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	client       *minio.Client
	ErrFileExist = errors.New("file already exists")
)

func getClient() (*minio.Client, error) {
	// Initialize minio client object.
	var err error
	client, err = minio.New(config.GetEndPoint(), &minio.Options{
		Creds: credentials.NewStaticV4(
			config.GetAccessUser(),
			config.GetAccessPassword(), ""),
		Secure: config.GetUseSSL(),
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
func Upload(bucket, filePath, remotePath string, forceUpload bool) (int64, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "text/plain"
	}

	client, err := GetClient()
	if err != nil {
		return 0, err
	}
	ctx := context.Background()

	if !forceUpload {
		_, err = client.StatObject(
			ctx, bucket, remotePath, minio.StatObjectOptions{})
		if err == nil {
			return 0, ErrFileExist
		}

		if !strings.Contains(err.Error(), "The specified key does not exist") {
			return 0, err
		}
	}

	info, err := client.FPutObject(
		ctx, bucket, remotePath, filePath,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return 0, err
	}
	return info.Size, err
}
