package common

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	ConfigName = "config.yaml"
)

type Config struct {
	EndPoint       string `yaml:"endPoint"`
	AccessUser     string `yaml:"accessUser"`
	AccessPassword string `yaml:"accessPassword"`
	Bucket         string `yaml:"bucket"`
}

var config Config

func LoadConfig() {
	// Load config
	ex, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot get executable: %v\n", err)
	}
	exPath := filepath.Dir(ex)

	file, err := os.ReadFile(filepath.Join(exPath, ConfigName))
	if err != nil {
		log.Fatalf("cannot read config file: %v\n", err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("cannot unmarshal config file: %v\n", err)
	}
}

func GetEndPoint() string {
	return config.EndPoint
}

func GetAccessUser() string {
	return config.AccessUser
}

func GetAccessPassword() string {
	return config.AccessPassword
}

func GetBucket() string {
	return config.Bucket
}
