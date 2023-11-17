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

func pushDir(localPath string) {

	fileinfo.LoadFileInfo("")
	defer fileinfo.WriteFileInfo("")

	err := statLocalFiles(localPath)
	if err != nil {
		fmt.Printf("failed to scan local files, err: %v\n", err)
		os.Exit(1)
	}

	err = statRemoteFiles(localPath, config.Config.Prefix, false)
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
		pushFile(f, i)
	}

	cnt.PrintCnt()
}

func pushFile(f string, i fileStat) {

	if config.Force {
		UploadFile(f, i.RemotePath)
		return
	}

	lStat := fileinfo.CheckFile(i.tLocal, i.tUpload)
	rStat := fileinfo.CheckFile(i.tRemote, i.tUpload)

	op := fileinfo.GetSyncStatus(lStat, rStat)
	if config.Debug {
		fmt.Printf("[%v] %v\n", fileinfo.OperationString(op), f)
	}

	switch op {
	case fileinfo.OpPush:
		UploadFile(f, i.RemotePath)
	case fileinfo.OpDelRemote:
		// delete remote file
		fmt.Println("! Need to implement: Del Remote file")
	case fileinfo.OpFork:
		// Report fork status
		cnt.Fork += 1
		fmt.Println("! Need to implement: Handle Forked file")
	//
	// ======================================================
	//
	case fileinfo.OpPull:
		// TODO should print push info in pull progress?
		return
	case fileinfo.OpDelLocal:
		// TODO should print push info in pull progress?
		return
	default:
		// No need to operate
		if config.Verbose {
			fmt.Printf("File %v is no need to modify\n", f)
		}
	}
}

func PushDir() {
	var remotePrefix string
	pflag.StringVarP(
		&remotePrefix, "remotePrefix", "p", "", "remote prefix add to path")
	pflag.BoolVarP(
		&config.Force, "forceUpload", "f", false, "force to upload files")
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

	config.LoadConfig(local)
	if remotePrefix != "" {
		config.Config.Prefix = remotePrefix
	}
	pushDir(local)

	for _, file := range fileinfo.GetUnvisitedFiles() {
		fmt.Printf("Deleted: %v\n", file)
	}

}
