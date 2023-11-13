package command

import (
	"fmt"
	"gomin-sync/internal/config"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	ConfigDir  = ".sync"
	ConfigFile = "config.yaml"
)

func initDir(path string, cf config.ConfigType) {
	basePath := filepath.Join(path, ConfigDir)
	configPath := filepath.Join(basePath, ConfigFile)

	_, err := os.Stat(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Failed to init path, err: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("This Path is already init, use command 'config' to configure")
		os.Exit(1)
	}

	if cf.AccessPassword == "" || cf.AccessUser == "" ||
		cf.Bucket == "" || cf.EndPoint == "" {

		fmt.Println("Endpoint, User, Password, Bucket must be specific")
		pflag.PrintDefaults()
		os.Exit(1)
	}

	out, err := yaml.Marshal(cf)
	if err != nil {
		fmt.Printf("Failed to write config, err: %v\n", err)
	}

	os.MkdirAll(basePath, os.ModePerm)
	os.WriteFile(configPath, out, 0744)
}

func Init() {
	cf := config.ConfigType{}

	pflag.StringVarP(&cf.EndPoint, "endpoint", "e", "", "endpoint for minio, like HOST[:PORT]")
	pflag.StringVarP(&cf.AccessUser, "user", "u", "", "minio user for auth")
	pflag.StringVarP(&cf.AccessPassword, "password", "p", "", "minio password for auth")
	pflag.StringVarP(&cf.Bucket, "bucket", "b", "", "specific bucket for sync")
	pflag.BoolVarP(&cf.UseSSL, "sslEnable", "s", false, "whether use ssl for minio")
	pflag.CommandLine.Parse(os.Args[2:])

	left := pflag.Args()
	var workDir string
	switch len(left) {
	case 0:
		workDir = ""
	case 1:
		workDir = left[0]
	default:
		fmt.Println("too many path, accept no path specific or one path")
		os.Exit(1)
	}

	initDir(workDir, cf)
}
