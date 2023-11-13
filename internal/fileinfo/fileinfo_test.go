package fileinfo

import (
	"testing"
)

func TestTransInfoString(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		modifyAt int64
		fileName string
	}{
		{
			"Basic without file name",
			"1600000000       ",
			1600000000,
			"",
		},
		{
			"Basic with file name",
			"1699891828     gm-sync",
			1699891828,
			"gm-sync",
		},
	}
	for _, c := range cases {
		res := transInfoString(c.input)
		if res == nil {
			t.Fatalf("[%s] trans(%s), expect %v, %v, got nil\n",
				c.name, c.input, c.modifyAt, c.fileName)
		}
		if res.ModifyAt != c.modifyAt || res.FileName != c.fileName {
			t.Fatalf("[%s] trans(%s), expect %v, %v, got %v, %v\n",
				c.name, c.input, c.modifyAt, c.fileName,
				res.ModifyAt, res.FileName)
		}
	}
}

func TestLoadFileInfo(t *testing.T) {

	LoadFileInfo("/home/cheng/workSpace/go/src/gomin-sync/build")
	// fmt.Printf("%+v\n", res)
	SetFileModifyTime("123", 12345)
	r, _ := GetFileModifyTime("123")
	// fmt.Printf("Modifytime: %v\n", r)

	if r != 12345 {
		t.Fatalf("Failed to get setted value")
	}

	WriteFileInfo("/home/cheng/workSpace/go/src/gomin-sync/build")
}
