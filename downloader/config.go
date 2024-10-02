package downloader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Path   string `toml:"path"`
	Url    string `toml:"url"`
	ApiKey string `toml:"apikey"`
}

func GetConfigDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("APPDATA")
	} else if runtime.GOOS == "linux" {
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			return os.Getenv("XDG_CONFIG_HOME")
		} else {
			return filepath.Join(os.Getenv("HOME"), ".config")
		}
	} else if runtime.GOOS == "darwin" {
		return os.Getenv("HOME") + "/Library/Application Support"
	} else {
		return ""
	}

}

func CheckConfig() (*Config, error) {
	path := GetConfigDir()
	if path == "" {
		return nil, errors.New("invalid OS")
	}
	err := os.Mkdir(path, os.ModePerm)
	if err != nil && os.IsNotExist(err) {
		return nil, err
	}

	configfile := filepath.Join(path, "ripper-config.toml")

	if _, err = os.Stat(configfile); os.IsNotExist(err) {
		fh, err := os.Create(configfile)
		if err != nil {
			return nil, err
		}

		defer fh.Close()

		err = toml.NewEncoder(bufio.NewWriter(fh)).Encode(Config{Path: "/home/user/Downloads/", Url: "https://test.dev", ApiKey: "test123"})
		if err != nil {
			return nil, err
		}

		fmt.Println("config template created. fill the required fields for usage")
		os.Exit(0)
	}

	var config Config

	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		return nil, err
	}

	if config.ApiKey == "" || config.Path == "" || config.Url == "" {
		return nil, errors.New("invalid configuration")
	}

	return &config, nil
}
