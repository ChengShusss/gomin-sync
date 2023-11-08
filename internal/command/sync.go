package command

import (
	"errors"
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/minioClient"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

var (
	ignoreList = []string{
		".git",
	}
)

func syncDir(basePath, remotePrefix string) {
	uploadCount := 0
	failedCount := 0
	filepath.WalkDir(basePath, func(path string, d fs.DirEntry, e error) error {
		// Skip basePath
		if path == basePath {
			return nil
		}

		// Skip file with specific prefix
		for _, ig := range ignoreList {
			if strings.HasPrefix(path, ig) {
				return nil
			}
		}

		// Skip dir, no need to upload
		if d.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}
		_, err = minioClient.Upload(
			config.GetBucket(), path,
			filepath.Join(remotePrefix, filepath.ToSlash(relative)),
			config.Force)

		if errors.Is(err, minioClient.ErrFileExist) {
			fmt.Printf("%s is already exist\n", filepath.Base(relative))
			return nil
		}

		if err != nil {
			fmt.Printf("  Failed to Upload %s, err: %v\n", relative, err)
			failedCount += 1
		} else {
			// fmt.Printf("  Success to Upload %s, Size: %v\n", relative, n)
			uploadCount += 1
		}
		return nil
	})

	fmt.Printf("Uploaded %d files, Failed %d files.\n", uploadCount, failedCount)
}

func SyncDir() {
	var remotePrefix string
	pflag.StringVarP(&remotePrefix, "remotePrefix", "p", "", "remote prefix add to path")
	pflag.BoolVarP(&config.Force, "forceUpload", "f", false, "force to upload files")
	pflag.CommandLine.Parse(os.Args[2:])

	left := pflag.Args()
	if len(left) == 0 {
		fmt.Printf("Please specific local dir to sync\n")
		os.Exit(1)
	}
	if len(left) > 1 {
		fmt.Printf("too many path is given\n")
		os.Exit(1)
	}

	config.LoadConfig("")
	syncDir(left[0], remotePrefix)
}
