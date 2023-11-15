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
		return -1
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
