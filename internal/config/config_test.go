package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSetBool(t *testing.T) {

	ex, err := os.Executable()
	if err != nil {
		t.Fatalf("cannot get executable: %v\n", err)
	}
	exPath := filepath.Dir(ex)

	fmt.Printf("ExePath: %v\n", exPath)
	LoadConfig("/home/cheng/workSpace/codeSpace/tinyGoProjects/gomin-sync/build/config.yaml")
	Config.Force = true
	fmt.Printf("Force: %v\n", GetBool("Fore"))
}
