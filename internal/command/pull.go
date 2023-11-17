package command

import (
	"errors"
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/fileinfo"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

var (
	ErrInvalidRemotePath = errors.New("invalid remote path, lack of prefix")
)

func pullDir(localPath, remotePath string) {

	fileinfo.LoadFileInfo(localPath)
	defer fileinfo.WriteFileInfo(localPath)

	err := statLocalFiles(localPath)
	if err != nil {
		fmt.Printf("failed to scan remote files, err: %v\n", err)
		os.Exit(1)
	}

	err = statRemoteFiles(localPath, remotePath, true)
	if err != nil {
		fmt.Printf("failed to scan remote files, err: %v\n", err)
		os.Exit(1)
	}

	for f, i := range fileMap {
		if i.RemotePath == "" {
			relative, err := filepath.Rel(localPath, f)
			if err != nil {
				cnt.Failed += 1
				fmt.Printf("Assemble remote path %v failed, err: %v\n", f, err)
				continue
			}
			i.RemotePath = filepath.Join(
				config.Config.Prefix,
				filepath.ToSlash(relative))
		}
		pullFile(f, i)
	}

	cnt.PrintCnt()

	// infoCh, err := minioClient.ListObjectsByPrefix(config.GetBucket(), remotePath)
	// if err != nil {
	// 	fmt.Printf("Failed to list info: %v\n", infoCh)
	// 	return
	// }

	// for info := range infoCh {
	// 	if info.Err != nil {
	// 		fmt.Printf("Failed to get object info, err: %v\n", info.Err)
	// 		return
	// 	}

	// 	if strings.HasSuffix(info.Key, "/") {
	// 		fmt.Printf("%v is dir\n", info.Key)
	// 		continue
	// 	}

	// 	// TODO need to check if localPath exist
	// 	filePath := normalizeLocalPath(localPath, config.Config.Prefix, info.Key)
	// 	if filePath == "" {
	// 		if config.Verbose {
	// 			fmt.Printf("Invalid Remote Key: %v\n", info.Key)
	// 		}
	// 		continue
	// 	}

	// 	if hasPrefix(filePath) {
	// 		continue
	// 	}

	// 	if config.Force {
	// 		DownloadFile(filePath, info.Key)
	// 	}

	// 	op := checkFileStatus(filePath, info.LastModified.Unix())
	// 	if config.Debug {
	// 		fmt.Printf("[%v] %v\n", fileinfo.OperationString(op), filePath)
	// 	}

	// 	switch op {
	// 	// below three is effective for pull
	// 	case fileinfo.OpPull:
	// 		DownloadFile(filePath, info.Key)
	// 	case fileinfo.OpDelLocal:
	// 		// delete local file
	// 		fmt.Println("! Need to implement: Del Local file")
	// 		continue
	// 	case fileinfo.OpFork:
	// 		// Report fork status
	// 		cnt.Fork += 1
	// 		fmt.Println("! Need to implement: Handle Forked file")
	// 		continue

	// 	case fileinfo.OpPush:
	// 		// TODO should print push info in pull progress?
	// 		continue
	// 	case fileinfo.OpDelRemote:
	// 		// delete remote file
	// 		// TODO should print push info in pull progress?
	// 		fmt.Println("! Need to implement: Del Remote file")
	// 		continue
	// 	default:
	// 		// No need to operate
	// 		if config.Verbose {
	// 			fmt.Printf("File %v is no need to modify\n", filePath)
	// 		}
	// 		continue
	// 	}
	// }

	// cnt.PrintCnt()
}

func pullFile(f string, i fileStat) {

	if config.Force {
		DownloadFile(f, i.RemotePath)
		return
	}

	lStat := fileinfo.CheckFile(i.tLocal, i.tUpload)
	rStat := fileinfo.CheckFile(i.tRemote, i.tUpload)

	op := fileinfo.GetSyncStatus(lStat, rStat)
	if config.Debug {
		fmt.Printf("[%v] %v\n", fileinfo.OperationString(op), f)
	}

	switch op {
	case fileinfo.OpPull:
		DownloadFile(f, i.RemotePath)
	case fileinfo.OpDelLocal:
		// delete local file
		fmt.Println("! Need to implement: Del Local file")
		return

	case fileinfo.OpFork:
		// Report fork status
		cnt.Fork += 1
		fmt.Println("! Need to implement: Handle Forked file")
	//
	// ======================================================
	//
	case fileinfo.OpPush:
		// TODO should print push info in pull progress?
		return
	case fileinfo.OpDelRemote:
		// TODO should print push info in pull progress?
		return
	default:
		// No need to operate
		if config.Verbose {
			fmt.Printf("File %v is no need to modify\n", f)
		}
	}
}

// func checkFileStatus(local string, tRemote int64) int {

// 	var tLocal, tUpload int64

// 	localInfo, err := os.Stat(local)
// 	if err == nil {
// 		tLocal = localInfo.ModTime().Unix()
// 	} else {
// 		if config.Debug {
// 			fmt.Printf("Failed to get local file info: %v, err: %v\n",
// 				local, err)
// 		}
// 	}

// 	tUpload = fileinfo.GetFileModifyTime(local)

// 	lStat := fileinfo.CheckFile(tLocal, tUpload)
// 	rStat := fileinfo.CheckFile(tRemote, tUpload)

// 	if config.Debug {
// 		fmt.Printf("  tLocal: %d, tRemote: %d, tUpload: %d\n",
// 			tLocal, tRemote, tUpload)
// 	}

// 	return fileinfo.GetSyncStatus(lStat, rStat)
// }

// func downloadObjectToLocal(localPath, remotePath string) error {
// 	fmt.Printf("Download %v\n", remotePath)
// 	dir := filepath.Dir(localPath)
// 	if _, err := os.Stat(dir); err != nil {
// 		if !os.IsNotExist(err) {
// 			return err
// 		}
// 		os.MkdirAll(filepath.Dir(localPath), os.ModePerm)
// 	}

// 	return minioClient.DownloadObject(
// 		config.GetBucket(), localPath, remotePath)
// }

func PullDir() {
	config.LoadConfig(".")

	pflag.BoolVarP(&config.Force, "force", "f", false,
		"force to download files, regardless of exist local files")
	pflag.BoolVarP(&config.Info, "info", "i", false,
		"only print info, not execute actually")
	pflag.BoolVarP(&config.Verbose, "verbose", "v", false,
		"show detailed infos")
	pflag.BoolVarP(&config.Debug, "debug", "", false,
		"show debug infos")
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
	pullDir(local, remotePath)
}
