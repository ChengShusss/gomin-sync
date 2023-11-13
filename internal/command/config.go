package command

import (
	"fmt"
	"gomin-sync/internal/config"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

func set(key, value string) {
	switch key {
	case "accessUser":
		config.Config.AccessUser = value
	case "accessPassword":
		config.Config.AccessPassword = value
	case "bucket":
		config.Config.Bucket = value
	case "endPoint":
		config.Config.EndPoint = value
	case "prefix":
		config.Config.Prefix = value
	case "useSSL":
		switch value {
		case "true", "True", "T", "t":
			config.Config.UseSSL = true
		case "false", "False", "F", "f":
			config.Config.UseSSL = false
		default:
			fmt.Printf("invalid params for useSSL")
			os.Exit(1)
		}
	}

	configPath := filepath.Join(".sync", ConfigFile)
	out, err := yaml.Marshal(config.Config)
	if err != nil {
		fmt.Printf("Failed to write config, err: %v\n", err)
	}
	os.WriteFile(configPath, out, 0744)

}

func get(keys []string) {
	keyMap := map[string]interface{}{
		"endPoint":       config.Config.EndPoint,
		"accessUser":     config.Config.AccessUser,
		"accessPassword": config.Config.AccessPassword,
		"bucket":         config.Config.Bucket,
		"prefix":         config.Config.Prefix,
		"useSSL":         config.Config.UseSSL,
	}

	if len(keys) == 0 {
		for key, value := range keyMap {
			fmt.Printf("%s = %v\n", key, value)
		}
		return
	}

	for _, key := range keys {
		value, ok := keyMap[key]
		if !ok {
			fmt.Printf("Invalid config key: %v\n", key)
		}
		fmt.Printf("%s = %v\n", key, value)
	}

}

func Config() {
	config.LoadConfig(".")

	var isSet bool
	pflag.BoolVarP(&isSet, "set", "s", false, "set config item if specific, nor get item")
	pflag.CommandLine.Parse(os.Args[2:])

	left := pflag.Args()

	if isSet {
		if len(left) != 2 {
			fmt.Printf("Please Set config one by one, in format of KEY VALUE")
			os.Exit(1)
		}
		set(left[0], left[1])
	} else {
		get(left)
	}
}
