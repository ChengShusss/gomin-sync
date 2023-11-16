package fileinfo

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ColTimeLen   = 15
	FileInfoName = ".sync/blob"

	FileNul = iota // file not exist
	FileAdd        // file is added
	FileMod        // file is modified
	FileDel        // file is deleted
	FileUnc        // file has no change
)

const (
	OpInvalid = iota
	OpPush
	OpPull
	OpDelLocal
	OpDelRemote
	OpFork
	OpNull
)

type FileInfo struct {
	ModifyAt int64
	FileName string
	Visited  bool
}

type FileMap map[string]FileInfo

var (
	fileMap     = FileMap{}
	ErrNotFound = errors.New("no record for this file")

	// map[localStatus]map[remoteStatus]Operation
	OpMap = map[int]map[int]int{
		FileNul: { // if local file is Nul, then tUpload must be 0
			FileNul: OpInvalid, // Not supposed to use this
			FileAdd: OpPull,
			FileMod: OpInvalid, // Not supposed to use this
			FileDel: OpInvalid, // Not supposed to use this
			FileUnc: OpInvalid, // Not supposed to use this
		},
		FileAdd: { // if local file is Add, then tUpload must be 0
			FileNul: OpPush,
			FileAdd: OpFork,
			FileMod: OpInvalid, // Not supposed to use this
			FileDel: OpInvalid,
			FileUnc: OpInvalid,
		},
		FileMod: { // if local file is Mod, then tUpload must not be 0
			FileNul: OpInvalid, // Not supposed to use this
			FileAdd: OpInvalid, // Not supposed to use this
			FileMod: OpFork,
			FileDel: OpFork,
			FileUnc: OpPush,
		},
		FileDel: { // if local file is Del, then tUpload must not be 0
			FileNul: OpInvalid, // Not supposed to use this
			FileAdd: OpInvalid, // Not supposed to use this
			FileMod: OpFork,
			FileDel: OpNull,
			FileUnc: OpDelRemote,
		},
		FileUnc: { // if local file is Unc, then tUpload/tLocal must not be 0
			FileNul: OpInvalid,
			FileAdd: OpInvalid, // Not supposed to appear
			FileMod: OpPull,
			FileDel: OpDelLocal,
			FileUnc: OpNull,
		},
	}
)

func transInfoString(s string) *FileInfo {
	info := FileInfo{}
	if len(s) < ColTimeLen+1 {
		return nil
	}

	t, err := strconv.ParseInt(strings.TrimSpace(s[:ColTimeLen]), 10, 64)
	if err != nil {
		fmt.Printf("Failed to Parse Int: %v, err: %v\n", s[:10], err)
		os.Exit(1)
	}

	info.ModifyAt = t
	info.FileName = strings.TrimSpace(s[ColTimeLen:])

	return &info
}

func LoadFileInfo(basePath string) {
	fileName := filepath.Join(basePath, FileInfoName)

	file, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fileInfo := transInfoString(line)
		fileMap[fileInfo.FileName] = *fileInfo
		// fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func WriteFileInfo(basePath string) {
	fileName := filepath.Join(basePath, FileInfoName)

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to open Upload Record file [%s], err: %v\n", fileName, err)
		os.Exit(1)
	}
	defer f.Close()

	for _, fi := range fileMap {
		f.WriteString(fmt.Sprintf("%-15d%s\n", fi.ModifyAt, fi.FileName))
	}
}

func SetFileModifyTime(file string, tm int64) {
	fi, ok := fileMap[file]
	if !ok {
		fi = FileInfo{
			FileName: file,
		}
	}

	fi.ModifyAt = tm
	fi.Visited = true
	fileMap[file] = fi
}

func GetFileModifyTime(file string) int64 {
	t, ok := fileMap[file]
	if !ok {
		return 0
	}
	t.Visited = true
	fileMap[file] = t
	return t.ModifyAt
}

func GetUnvisitedFiles() []string {
	res := []string{}

	for _, fi := range fileMap {
		if !fi.Visited {
			res = append(res, fi.FileName)
		}

	}
	return res
}

// CheckFileState - return local file status and remote file status
// Params
//
//	tLocal : local file last modified time;
//	tUpload: localhost upload file time;
//	tRemote: remote file last modified time;
//
// Return
//
//	localFileStatus, remoteFileStatus
// func CheckFileState(tLocal, tUpload, tRemote int64) (int, int) {
// 	if tLocal == 0 {
// 		return Ob
// 	}
// }

func CheckFile(tTarget, tUpload int64) int {
	switch {
	case tTarget == 0 && tUpload == 0:
		return FileNul
	case tTarget == 0:
		return FileDel
	case tUpload == 0:
		return FileAdd
	case tUpload >= tTarget:
		return FileUnc
	case tTarget > tUpload:
		return FileMod
	}

	// Invalid case, should not go here
	return FileUnc
}

func GetSyncStatus(lStatus, rStatus int) int {
	mm, ok := OpMap[lStatus]
	if !ok {
		return OpInvalid
	}
	r, ok := mm[rStatus]
	if !ok {
		return OpInvalid
	}
	return r
}

func OperationString(op int) string {
	m := map[int]string{
		OpPush:      "Push",
		OpPull:      "Pull",
		OpDelLocal:  "DelL",
		OpDelRemote: "DelR",
		OpNull:      "Null",
		OpInvalid:   "Err-",
	}

	s, ok := m[op]
	if !ok {
		return "Inva"
	}
	return s
}
