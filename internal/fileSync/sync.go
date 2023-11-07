package fileSync

import (
	"fmt"
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
		n, err := minioClient.Upload(
			"clip", path,
			filepath.Join(remotePrefix, filepath.ToSlash(relative)))

		if err != nil {
			fmt.Printf("  Failed to Upload %s, err: %v\n", path, err)
		} else {
			fmt.Printf("  Success to Upload %s, Size: %v\n", path, n)
		}

		return nil
	})
}

func SyncDir() {
	var remotePrefix string
	pflag.StringVarP(&remotePrefix, "remotePrefix", "p", "", "remote prefix add to path")
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
	syncDir(left[0], remotePrefix)
}
