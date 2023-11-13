package command

import (
	"errors"
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/minioClient"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

var (
	ErrInvalidRemotePath = errors.New("invalid remote path, lack of prefix")
)

func downloadDir(localPath, remotePath string) {
	downCount := 0
	failedCount := 0

	infoCh, err := minioClient.ListObjectsByPrefix(config.GetBucket(), remotePath)
	if err != nil {
		fmt.Printf("Failed to list info: %v\n", infoCh)
		return
	}

	for info := range infoCh {
		if info.Err != nil {
			fmt.Printf("Failed to get object info, err: %v\n", info.Err)
			return
		}

		if strings.HasSuffix(info.Key, "/") {
			fmt.Printf("%v is dir\n", info.Key)
			continue
		}

		// TODO need to check if localPath exist
		err := downloadObjectToLocal(localPath, info.Key, remotePath)

		if err != nil {
			fmt.Printf("Failed to download %s, err: %v\n", info.Key, err.Error())
			failedCount += 1
		} else {
			downCount += 1
		}
	}

	fmt.Printf("Download %d files, Failed %d files.\n", downCount, failedCount)

}

func downloadObjectToLocal(localBase, remotePath, prefix string) error {
	remotePath = filepath.ToSlash(remotePath)
	prefix = filepath.ToSlash(prefix)
	if !strings.HasPrefix(remotePath, prefix) {
		return ErrInvalidRemotePath
	}
	suffixPath := remotePath[len(prefix):]

	newLocalPath := filepath.Join(localBase, suffixPath)
	dir := filepath.Dir(newLocalPath)
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		os.MkdirAll(filepath.Dir(newLocalPath), os.ModePerm)
	}
	return minioClient.DownloadObject(
		config.GetBucket(), newLocalPath, remotePath)
}

func PullDir() {
	config.LoadConfig("")

	pflag.BoolVarP(&config.Force, "forcePull", "f", false, "force to download files")
	pflag.CommandLine.Parse(os.Args[2:])

	left := pflag.Args()
	var local string
	if len(left) == 0 {
		local = "."
	}
	if len(left) > 1 {
		fmt.Printf("too many path is given\n")
		os.Exit(1)
	}

	remotePath := "default"
	downloadDir(local, remotePath)
}
