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

	"github.com/spf13/pflag"
)

type fileStat struct {
	tRemote    int64
	tLocal     int64
	tUpload    int64
	RemotePath string
}

// key: local file Name; value: file info
var syncFileMap = map[string]fileStat{}

var timeOffset int64 = -32

func syncDir(localPath, remotePath string) {
	// downCount := 0
	// failedCount := 0
	// forkCount := 0
	// omitCount := 0
	// localCount := 0

	fileinfo.LoadFileInfo("")
	defer fileinfo.WriteFileInfo("")

	err := statLocalFiles(localPath)
	if err != nil {
		fmt.Printf("failed to scan local files, err: %v\n", err)
		os.Exit(1)
	}

	err = statRemoteFiles(localPath, remotePath)
	if err != nil {
		fmt.Printf("failed to scan local files, err: %v\n", err)
		os.Exit(1)
	}

	for f, i := range syncFileMap {
		if i.RemotePath == "" {
			relative, err := filepath.Rel(localPath, f)
			if err != nil {
				fmt.Printf("Assemble remote path %v failed, err: %v\n", f, err)
				continue
			}
			i.RemotePath = filepath.Join(
				config.Config.Prefix,
				filepath.ToSlash(relative))
		}
		syncFile(f, i)
	}

	fmt.Printf("Total %v files local and remote\n", len(syncFileMap))

	cnt.PrintCnt()
}

func syncFile(f string, i fileStat) {
	lStat := fileinfo.CheckFile(i.tLocal, i.tUpload)
	rStat := fileinfo.CheckFile(i.tRemote, i.tUpload)

	op := fileinfo.GetSyncStatus(lStat, rStat)
	if config.Debug {
		fmt.Printf("[%v] %v\n", fileinfo.OperationString(op), i.RemotePath)
	}

	switch op {
	case fileinfo.OpPush:
		UploadFile(f, i.RemotePath)
	case fileinfo.OpPull:
		DownloadFile(f, i.RemotePath)
	case fileinfo.OpDelLocal:
		// delete local file
		fmt.Printf("! Need to implement: Del Local file")
	case fileinfo.OpDelRemote:
		// delete remote file
		fmt.Printf("! Need to implement: Del Remote file")
	case fileinfo.OpFork:
		// Report fork status
		fmt.Printf("! Need to implement: Handle Forked file")
	default:
		// No need to operate
		if config.Verbose {
			fmt.Printf("File %v is no need to modify\n", f)
		}
	}
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

		info, ok := syncFileMap[p]
		if !ok {
			info = fileStat{}
		}

		i, err := d.Info()
		if err != nil {
			return err
		}

		info.tLocal = i.ModTime().Unix()
		info.tUpload = fileinfo.GetFileModifyTime(p)

		syncFileMap[p] = info
		return nil
	})
}

func statRemoteFiles(localBase, remotePath string) error {
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

		statInfo, ok := syncFileMap[filePath]
		if !ok {
			statInfo = fileStat{}
		}

		statInfo.RemotePath = info.Key
		statInfo.tRemote = info.LastModified.Unix() + timeOffset
		statInfo.tUpload = fileinfo.GetFileModifyTime(filePath)

		syncFileMap[filePath] = statInfo
	}

	return nil
}

func SyncDir() {
	config.LoadConfig(".")

	pflag.BoolVarP(&config.Info, "info", "i", false,
		"only print info, not execute actually")
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
	syncDir(local, remotePath)
}
