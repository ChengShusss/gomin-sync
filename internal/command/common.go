package command

import (
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/fileinfo"
	"gomin-sync/internal/minioClient"
	"os"
	"path/filepath"
	"time"
)

type Counter struct {
	Upload    int
	Download  int
	Fork      int
	DelLocal  int
	DelRemote int
	Failed    int
}

var cnt Counter

func (c *Counter) PrintCnt() {

	fmt.Println("Summary:")
	fmt.Printf("! Forked files:    %v\n", c.Fork)
	fmt.Printf("  Upload files:    %v\n", c.Upload)
	fmt.Printf("  Download files:  %v\n", c.Download)
	fmt.Printf("  DelLocal files:  %v\n", c.DelLocal)
	fmt.Printf("  DelRemote files: %v\n", c.DelRemote)
	fmt.Printf("  Failed:          %v\n", c.Failed)
}

func UploadFile(local, remote string) {
	if config.Debug {
		fmt.Printf("Upload: %v\n", local)
	}

	var err error
	var n int64
	if config.Info {
		cnt.Upload += 1
		fmt.Printf("Need to upload %v\n", local)
		return
	}
	n, err = minioClient.Upload(
		config.GetBucket(), local, remote, config.Force)

	if err != nil {
		cnt.Failed += 1
		fmt.Printf("Failed to upload %v, err: %v\n", local, err)
		return
	}

	fileinfo.SetFileModifyTime(local, time.Now().Unix())
	cnt.Upload += 1
	if config.Verbose {
		fmt.Printf("  Success to Upload %s, Size: %v\n", local, n)
	}
}

func DownloadFile(local, remote string) {

	if config.Info {
		fmt.Printf("  Need to download %v\n", local)
		cnt.Download += 1
		return
	}

	fmt.Printf("  Download [%v]\n", remote)

	dir := filepath.Dir(local)
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("  Failed to stat dir %v, err: %v\n", dir, err)
			return
		}
		os.MkdirAll(filepath.Dir(local), os.ModePerm)
	}

	err := minioClient.DownloadObject(
		config.GetBucket(), local, remote)
	if err != nil {
		fmt.Printf("  Failed to download %v, err: %v\n", remote, err)
		cnt.Failed += 1
		return
	}

	cnt.Download += 1
	if config.Verbose {
		fmt.Printf("  Success to download %s\n", local)
	}
}

func DeleteRemoteFile(remote string) error {

	return nil
}

func DeleteLocalFile(local string) error {

	return nil
}
