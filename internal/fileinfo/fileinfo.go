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
}

type FileMap map[string]int64

var (
	fileMap     = FileMap{}
	ErrNotFound = errors.New("no record for this file")
)

func (fi *FileInfo) String() string {
	return fmt.Sprintf("%-15d%s\n", fi.ModifyAt, fi.FileName)
}

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

func LoadFileInfo(basePath string) map[string]int64 {
	fileName := filepath.Join(basePath, FileInfoName)

	file, err := os.Open(fileName)
	if err != nil {
		return fileMap
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fileInfo := transInfoString(line)
		fileMap[fileInfo.FileName] = fileInfo.ModifyAt
		// fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return fileMap
}

func WriteFileInfo(basePath string) {
	fileName := filepath.Join(basePath, FileInfoName)

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to open Upload Record file [%s], err: %v\n", fileName, err)
		os.Exit(1)
	}
	defer f.Close()

	for name, tm := range fileMap {
		f.WriteString(fmt.Sprintf("%-15d%s\n", tm, name))
	}
}

func SetFileModifyTime(file string, tm int64) {
	fileMap[file] = tm
}

func GetFileModifyTime(file string) (int64, error) {
	t, ok := fileMap[file]
	if !ok {
		return -1, ErrNotFound
	}
	return t, nil
}
