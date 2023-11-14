package command

import (
	"errors"
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/fileinfo"
	"gomin-sync/internal/minioClient"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var (
	ErrInvalidRemotePath = errors.New("invalid remote path, lack of prefix")
)

func downloadDir(localPath, remotePath string) {
	downCount := 0
	failedCount := 0
	forkCount := 0
	omitCount := 0
	localCount := 0

	fileinfo.LoadFileInfo("")
	defer fileinfo.WriteFileInfo("")

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
		filePath := normalizeLocalPath(localPath, config.Config.Prefix, info.Key)
		if filePath == "" {
			fmt.Printf("Invalid Remote Key: %v\n", info.Key)
			continue
		}

		if hasPrefix(filePath) {
			continue
		}

		if !config.Force {
			localInfo, err := os.Stat(filePath)
			if err == nil {
				// local file exists
				tLocal := localInfo.ModTime().Unix()
				tUpload := fileinfo.GetFileModifyTime(filePath)
				st, err := minioClient.CheckObject(config.GetBucket(), info.Key,
					tLocal, tUpload)
				if err != nil {
					fmt.Printf("Failed to get object, err: %v\n", err)
					os.Exit(1)
				}
				switch st {
				case minioClient.ObjectForked:
					fmt.Printf("  File %v has been modified both local and remote\n",
						filePath)
					forkCount += 1
					continue
				case minioClient.ObjectLocalModified:
					if config.Verbose {
						fmt.Printf("  File %v only change locally, no need to pull\n",
							filePath)
					}
					localCount += 1
					continue
				case minioClient.ObjectNochange:
					if config.Verbose {
						fmt.Printf("  File %v has no change, no need to pull\n",
							filePath)
					}
					omitCount += 1
					continue
				case minioClient.ObjectRemoteModified:
				default:
					continue
				}
			} else {
				if !os.IsNotExist(err) {
					fmt.Printf("Failed to compare local file, err: %v\n", err)
					os.Exit(1)
				}
			}
		}

		err := downloadObjectToLocal(filePath, info.Key)
		if err != nil {
			// fmt.Printf("Failed to download %s, err: %v\n", info.Key, err.Error())
			failedCount += 1
		} else {
			fileinfo.SetFileModifyTime(localPath, time.Now().Unix())
			downCount += 1
		}
	}

	fmt.Printf("Download %d files, Failed %d files.\n", downCount, failedCount)

	fmt.Printf("Nochange: %d , Local-Changed: %d, Forked: %d\n",
		omitCount, localCount, forkCount)
}

func downloadObjectToLocal(localPath, remotePath string) error {
	fmt.Printf("Download %v\n", remotePath)
	dir := filepath.Dir(localPath)
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		os.MkdirAll(filepath.Dir(localPath), os.ModePerm)
	}

	return minioClient.DownloadObject(
		config.GetBucket(), localPath, remotePath)
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

func PullDir() {
	config.LoadConfig("")

	pflag.BoolVarP(&config.Force, "force", "f", false,
		"force to download files, regardless of exist local files")
	pflag.BoolVarP(
		&config.Verbose, "verbose", "v", false, "show detailed infos")
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

	remotePath := config.Config.Prefix
	downloadDir(local, remotePath)
}
