package minioClient

import (
	"context"
	"fmt"
	"gomin-sync/internal/config"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	ObjectInvalid = iota
	ObjectRemoteNotExist
	ObjectRemoteModified
	ObjectLocalModified
	ObjectForked
	ObjectNochange
)

var (
	client     *minio.Client
	timeOffset int64 = -32
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

	info, err := client.FPutObject(
		ctx, bucket, remotePath, filePath,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return 0, err
	}
	return info.Size, err
}

func ListObjectsByPrefix(bucket, prefix string) (<-chan minio.ObjectInfo, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	return client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}), nil
}

func DownloadObject(bucket, localPath, remotePath string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	object, err := client.GetObject(
		ctx, bucket, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer object.Close()

	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, object)

	return err
}

func CheckObject(bucket, remotePath string, tLocal, tLastUpload int64) (int, error) {
	client, err := GetClient()
	if err != nil {
		return ObjectInvalid, err
	}
	ctx := context.TODO()
	info, err := client.StatObject(
		ctx, bucket, remotePath, minio.StatObjectOptions{})
	if err == nil {
		tRemote := info.LastModified.Unix() + timeOffset
		if config.Verbose {
			fmt.Printf("%v: tLocal: %d, tLastUpload: %d, tRemote: %v\n",
				remotePath, tLocal, tLastUpload, tRemote)
		}
		if tRemote >= tLastUpload {
			// file changed remotely after last upload from local
			if tLocal >= tLastUpload {
				// file also changed locally
				return ObjectForked, nil
			}
			return ObjectRemoteModified, nil

		} else {
			if tLastUpload >= tLocal {
				// There is no change for this file
				return ObjectNochange, nil
			}
			// Only change happend locally, go stright
			return ObjectLocalModified, nil
		}

	}

	if strings.Contains(err.Error(), "The specified key does not exist") {
		// Means not Object-Not-Exists error, should return origin error
		return ObjectRemoteNotExist, nil
	}

	return ObjectInvalid, err
}
