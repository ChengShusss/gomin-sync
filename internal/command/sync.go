package command

import (
	"fmt"
	"gomin-sync/internal/config"
	"gomin-sync/internal/fileinfo"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

func syncDir(localPath, remotePath string) {

	fileinfo.LoadFileInfo(localPath)
	defer fileinfo.WriteFileInfo(localPath)

	err := statLocalFiles(localPath)
	if err != nil {
		fmt.Printf("failed to scan local files, err: %v\n", err)
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
				fmt.Printf("Assemble remote path %v failed, err: %v\n", f, err)
				continue
			}
			i.RemotePath = filepath.Join(
				config.Config.Prefix,
				filepath.ToSlash(relative))
		}
		syncFile(f, i)
	}

	fmt.Printf("Total %v files local and remote\n", len(fileMap))

	cnt.PrintCnt()
}

func syncFile(f string, i fileStat) {
	lStat := fileinfo.CheckFile(i.tLocal, i.tUpload)
	rStat := fileinfo.CheckFile(i.tRemote, i.tUpload)

	op := fileinfo.GetSyncStatus(lStat, rStat)
	if config.Debug {
		fmt.Printf("[%v] %v\n", fileinfo.OperationString(op), f)
	}
	if config.Debug {
		fmt.Printf("  tLocal: %v, tUpload: %v, tRemote:%v\n",
			i.tLocal, i.tUpload, i.tRemote)
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
		fmt.Printf("  Local: %d, Remote: %v\n", lStat, rStat)
		// Report fork status
		fmt.Printf("! Need to implement: Handle Forked file")
	default:
		// No need to operate
		if config.Debug {
			fmt.Printf("File %v is no need to modify\n", f)
		}
	}
}

func SyncDir() {

	pflag.BoolVarP(&config.Info, "info", "i", false,
		"only print info, not execute actually")
	pflag.BoolVarP(&config.Verbose, "verbose", "v", false,
		"show detailed infos")
	pflag.BoolVarP(&config.Debug, "debug", "", false,
		"show debug infos")
	pflag.CommandLine.Parse(os.Args[2:])

	left := pflag.Args()
	var local string
	switch len(left) {
	case 0:
		local = "."
	case 1:
		local = left[0]
	default:
		fmt.Printf("too many path is given\n")
		os.Exit(1)
	}

	config.LoadConfig(local)

	remotePath := config.Config.Prefix
	syncDir(local, remotePath)
}
