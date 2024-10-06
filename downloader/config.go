package downloader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
)

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
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil && os.IsNotExist(err) {
		return nil, err
	}

	configfile := filepath.Join(path, "ripper-config.toml")

	if _, err = os.Stat(configfile); os.IsNotExist(err) {
		fh, err := os.Create(configfile)
		if err != nil {
			return nil, err
		}

		defer func(fh *os.File) {
			err := fh.Close()
			if err != nil {
				fmt.Printf("error while closing: %v\n", err.Error())
			}
		}(fh)

		if runtime.GOOS == "windows" {
			err = toml.NewEncoder(bufio.NewWriter(fh)).Encode(Config{Path: "C:\\Users\\Test\\Downloads\\", Url: "https://test.dev/", ApiKey: "test123", Unarchive: false, Convert: false})
		} else {
			err = toml.NewEncoder(bufio.NewWriter(fh)).Encode(Config{Path: "/home/user/Downloads/", Url: "https://test.dev/", ApiKey: "test123", Unarchive: false, Convert: false})
		}
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

	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(config.Url, "/") {
			config.Url = fmt.Sprintf("%v/", config.Url)
		}

		if !strings.HasSuffix(config.Path, "\\") {
			config.Path = fmt.Sprintf("%v\\", config.Path)
		}

		err := os.MkdirAll(config.Path, os.ModePerm)
		if err != nil && os.IsNotExist(err) {
			return nil, err
		}
	} else {
		if !strings.HasSuffix(config.Url, "/") {
			config.Url = fmt.Sprintf("%v/", config.Url)
		}

		if !strings.HasSuffix(config.Path, "/") {
			config.Path = fmt.Sprintf("%v/", config.Path)
		}

		err := os.MkdirAll(config.Path, os.ModePerm)
		if err != nil && os.IsNotExist(err) {
			return nil, err
		}
	}

	if !config.Unarchive && config.Convert {
		color.Red("Invalid configuration! Convert must be used with unarchive")
	}

	return &config, nil
}
