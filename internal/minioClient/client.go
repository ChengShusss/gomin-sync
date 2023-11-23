package minioClient

import (
	"bytes"
	"context"
	"fmt"
	"gomin-sync/internal/config"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/gohouse/golib/random"
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
	TimeOffset int64
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
		return nil, nil
	}

	TimeOffset, err = getTimeOffset()
	if config.Verbose {
		fmt.Printf("Time offset: %vs\n", TimeOffset)
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

	f, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, object)

	return err
}

// getTimeOffset return time offset between remote and local
// return tRemote - tLocal, value is approximately since this function only use
// once upload to measure time delay.
func getTimeOffset() (int64, error) {
	client, err := GetClient()
	if err != nil {
		return 0, err
	}
	ctx := context.TODO()
	tmpFileName := ".sync-" + random.RandString(15)
	if config.Debug {
		fmt.Printf("tmp: %v\n", tmpFileName)
	}

	buf := []byte{}
	buff := bytes.NewReader(buf)

	remote := filepath.Join(config.Config.Prefix, tmpFileName)
	_, err = client.PutObject(
		ctx, config.Config.Bucket, remote, buff, 0,
		minio.PutObjectOptions{ContentType: "text/plain"})
	if err != nil {
		return 0, err
	}

	t := time.Now().Unix()
	info, err := client.StatObject(
		ctx, config.Config.Bucket, remote, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}

	if config.Debug {
		fmt.Printf("tLocal: %v, tRemote: %v\n", t, info.LastModified.Unix())
	}

	_ = client.RemoveObject(ctx, config.Config.Bucket, remote,
		minio.RemoveObjectOptions{})

	return info.LastModified.Unix() - t, nil
}
