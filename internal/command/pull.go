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

func PullDir() {

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
	fmt.Printf("Left args: %v\n", left)
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
	pullDir(local, remotePath)
}
