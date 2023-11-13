package config

import (
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
}

var (
	Config ConfigType
	Force  bool
)

func LoadConfig(path string) {
	var file string
	if path == "" {
		ex, err := os.Executable()
		if err != nil {
			log.Fatalf("cannot get executable: %v\n", err)
		}
		path = filepath.Dir(ex)
	}
	file = filepath.Join(path, ConfigName)

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
