package command

import (
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/fileinfo"
	"gomin-sync/internal/minioClient"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

type fileStat struct {
	tRemote    int64
	tLocal     int64
	tUpload    int64
	RemotePath string
}

var cnt Counter

// key: local file Name; value: file info
var fileMap = map[string]fileStat{}

func (c *Counter) PrintCnt() {

	warnIcon := " "
	if c.Fork > 0 {
		warnIcon = "!"
	}

	fmt.Println("=====================")
	fmt.Println("| Summary:          |")
	fmt.Println("|-------------------|")
	fmt.Printf("| Upload:    %6d |\n", c.Upload)
	fmt.Printf("| Download:  %6d |\n", c.Download)
	fmt.Printf("| DelLocal:  %6d |\n", c.DelLocal)
	fmt.Printf("| DelRemote: %6d |\n", c.DelRemote)
	fmt.Printf("|%sForked:    %6d |\n", warnIcon, c.Fork)
	fmt.Printf("| Failed:    %6d |\n", c.Failed)

	fmt.Println("|-------------------|")
	fmt.Printf("| Total :    %6d |\n",
		len(fileMap))
	fmt.Println("=====================")
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

	fileinfo.SetFileModifyTime(local, time.Now().Unix())
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

func normalizeLocalPath(localBase, prefix, remotePath string) string {
	remotePath = filepath.ToSlash(remotePath)
	prefix = filepath.ToSlash(prefix)
	if !strings.HasPrefix(remotePath, prefix) {
		return ""
	}
	suffixPath := remotePath[len(prefix):]

	return filepath.Join(localBase, suffixPath)
}

func hasPrefix(path string) bool {
	// Skip file with specific prefix
	for _, ig := range ignoreList {
		if strings.HasPrefix(path, ig) {
			return true
		}
	}
	return false
}

func statLocalFiles(path string) error {
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, e error) error {
		// Skip basePath
		if p == path {
			return nil
		}

		// Skip file with specific prefix
		for _, ig := range ignoreList {
			if strings.HasPrefix(p, ig) {
				return nil
			}
		}

		// Skip dir, no need to upload
		if d.IsDir() {
			return nil
		}

		info, ok := fileMap[p]
		if !ok {
			info = fileStat{}
		}

		i, err := d.Info()
		if err != nil {
			return err
		}

		info.tLocal = i.ModTime().Unix()
		info.tUpload = fileinfo.GetFileModifyTime(p)

		fileMap[p] = info
		return nil
	})
}

func statRemoteFiles(localBase, remotePath string, addRemoteItem bool) error {
	if config.Debug {
		fmt.Printf("Remote Path: %v\n", remotePath)
	}

	infoCh, err := minioClient.ListObjectsByPrefix(config.GetBucket(), remotePath)
	if err != nil {
		return err
	}

	for info := range infoCh {
		if info.Err != nil {
			return err
		}

		if strings.HasSuffix(info.Key, "/") {
			if config.Verbose {
				fmt.Printf("%v is dir\n", info.Key)
			}
			continue
		}

		// TODO need to check if localPath exist
		filePath := normalizeLocalPath(localBase, config.Config.Prefix, info.Key)

		if hasPrefix(filePath) {
			if config.Verbose {
				fmt.Printf("Omit remote file%v\n", filePath)
			}
			continue
		}

		statInfo, ok := fileMap[filePath]
		if !ok {
			if !addRemoteItem {
				continue
			}
			statInfo = fileStat{}
		}

		statInfo.RemotePath = info.Key
		statInfo.tRemote = info.LastModified.Unix() + minioClient.TimeOffset
		statInfo.tUpload = fileinfo.GetFileModifyTime(filePath)

		fileMap[filePath] = statInfo
	}

	return nil
}
