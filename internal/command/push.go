package command

import (
	"errors"
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/fileinfo"
	"gomin-sync/internal/minioClient"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

const (
	BlobName = ".sync/blob"
)

var (
	ignoreList = []string{
		".git",
		".sync",
		"gm-sync",
	}
	ErrFileForked       = errors.New("file modified both local and remote")
	ErrFileNoChange     = errors.New("file has no change after last upload")
	ErrFileRemoteModify = errors.New("file modified remote")
	ErrFileLocalModify  = errors.New("file modified local")
)

func pushDir(basePath string) {
	uploadCount := 0
	failedCount := 0
	omitCount := 0
	forkCount := 0
	remoteCount := 0

	fileinfo.LoadFileInfo("")
	defer fileinfo.WriteFileInfo("")

	filepath.WalkDir(basePath, func(p string, d fs.DirEntry, e error) error {
		// Skip basePath
		if p == basePath {
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

		err := handleEntry(basePath, p, d, e)
		if err == nil {
			uploadCount += 1
			return nil
		}

		switch {
		case errors.Is(err, ErrFileForked):
			fmt.Printf("  File %s both change local and remote\n", p)
			forkCount += 1
		case errors.Is(err, ErrFileNoChange):
			if config.Verbose {
				fmt.Printf("  File %s has no change\n", p)
			}
			omitCount += 1
		case errors.Is(err, ErrFileRemoteModify):
			fmt.Printf("  File %s changeg at remote, need update\n", p)
			remoteCount += 1
		default:
			failedCount += 1
			fmt.Printf("  Failed to Upload %s, err: %v\n", p, err)
			return err
		}

		return nil
	})

	fmt.Printf("Uploaded %d files, Failed %d files.\n",
		uploadCount, failedCount)
	fmt.Printf("Nochange: %d , Remote-Changed: %d, Forked: %d\n",
		omitCount, remoteCount, forkCount)
}

func handleEntry(basePath, path string, d fs.DirEntry, e error) error {

	relative, err := filepath.Rel(basePath, path)
	if err != nil {
		return err
	}
	remotePath := filepath.Join(config.Config.Prefix, filepath.ToSlash(relative))

	info, err := d.Info()
	if err != nil {
		return err
	}

	st, err := minioClient.CheckObject(
		config.GetBucket(), remotePath,
		info.ModTime().Unix(), fileinfo.GetFileModifyTime(path))
	if err != nil {
		return err
	}

	if !config.Force {
		switch st {
		case minioClient.ObjectForked:
			return ErrFileForked
		case minioClient.ObjectNochange:
			return ErrFileNoChange
		case minioClient.ObjectRemoteModified:
			return ErrFileRemoteModify
		case minioClient.ObjectLocalModified,
			minioClient.ObjectRemoteNotExist:
		default:
			return errors.New("unimplemented File Status")
		}
	}

	n, err := minioClient.Upload(
		config.GetBucket(), path,
		remotePath,
		config.Force)

	if err == nil {
		if config.Verbose {
			fmt.Printf("  Success to Upload %s, Size: %v\n", relative, n)
		}

		// Mark file upload time
		fileinfo.SetFileModifyTime(path, time.Now().Unix())
		return nil
	}

	return err
}

func PushDir() {
	var remotePrefix string
	pflag.StringVarP(
		&remotePrefix, "remotePrefix", "p", "", "remote prefix add to path")
	pflag.BoolVarP(
		&config.Force, "forceUpload", "f", false, "force to upload files")
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

	config.LoadConfig(local)
	config.Config.Prefix = remotePrefix
	pushDir(local)

	for _, file := range fileinfo.GetUnvisitedFiles() {
		fmt.Printf("Deleted: %v\n", file)
	}

}
