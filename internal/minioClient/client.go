package minioClient

import (
	"context"
	"errors"
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/fileinfo"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	client                *minio.Client
	ErrFileExist          = errors.New("file already exists")
	ErrFileModifiedRemote = errors.New("file already been modified at remote")
	ErrFileForked         = errors.New("file already been modified at local and remote")
	ErrFileNoChange       = errors.New("file has no change after last upload")
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
		i, err := os.Stat(filePath)
		if err != nil {
			return 0, err
		}
		tLocal := i.ModTime().Unix()
		tLastUpload, _ := fileinfo.GetFileModifyTime(filePath)

		info, err := client.StatObject(
			ctx, bucket, remotePath, minio.StatObjectOptions{})
		if err == nil {
			tRemote := info.LastModified.Unix()
			fmt.Printf("tLocal: %d, tLastUpload: %d, tRemote: %v\n", tLocal, tLastUpload, tRemote)
			if tRemote >= tLastUpload {
				// file changed remotely after last upload from local
				if tLocal >= tLastUpload {
					return 0, ErrFileForked
				} else {
					return 0, ErrFileModifiedRemote
				}
			} else {
				if tLastUpload >= tLocal {
					// There is no change for this file
					return 0, ErrFileNoChange
				}
				// Only change happend locally, go stright
			}

		} else {
			if !strings.Contains(err.Error(), "The specified key does not exist") {
				return 0, err
			}
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
