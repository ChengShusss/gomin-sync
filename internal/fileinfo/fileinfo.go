package fileinfo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ColTimeLen = 15
)

var (
	FileInfoName = ".sync/fileInfo"
)

type FileInfo struct {
	ModifyAt int64
	FileName string
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
	info.FileName = strings.TrimSpace(s[ColTimeLen+1:])

	return &info
}

func LoadFileInfo(basePath string) {
	fileName := filepath.Join(basePath, FileInfoName)
	fileMap := map[string]int64{}

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fileInfo := transInfoString(line)
		fileMap[fileInfo.FileName] = fileInfo.ModifyAt

		fmt.Println(scanner.Text())
	}

	fmt.Printf("%+v\n", fileMap)

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
