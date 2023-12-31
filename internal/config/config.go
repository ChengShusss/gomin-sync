package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

const (
	ConfigName = ".sync/config.yaml"
)

type ConfigType struct {
	EndPoint       string `yaml:"endPoint"`
	AccessUser     string `yaml:"accessUser"`
	AccessPassword string `yaml:"accessPassword"`
	Bucket         string `yaml:"bucket"`
	UseSSL         bool   `yaml:"useSSL"`
	Prefix         string `yaml:"prefix"`
}

var (
	Config  ConfigType
	Force   bool
	Verbose bool
	Info    bool
	Debug   bool = true

	BuildTime string
)

func LoadConfig(path string) {
	if Debug {
		fmt.Printf("Build Time: %v\n", BuildTime)
	}
	file := filepath.Join(path, ConfigName)

	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("cannot read config file: %v\n", err)
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Fatalf("cannot unmarshal config file: %v\n", err)
	}
}

func GetEndPoint() string {
	return Config.EndPoint
}

func GetAccessUser() string {
	return Config.AccessUser
}

func GetAccessPassword() string {
	return Config.AccessPassword
}

func GetBucket() string {
	return Config.Bucket
}

func GetUseSSL() bool {
	return Config.UseSSL
}

func GetString(key string) string {
	v := reflect.ValueOf(Config).FieldByName(key)
	return v.String()
}

func GetBool(key string) bool {
	v := reflect.ValueOf(Config).FieldByName(key)

	if isZero(v) {
		return false
	}
	return v.Bool()
}

func isZero(v interface{}) bool {
	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}
