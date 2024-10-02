package downloader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/kirsle/configdir"
)

type Config struct {
	Path   string `toml:"path"`
	Url    string `toml:"url"`
	ApiKey string `toml:"apikey"`
}

func CheckConfig() (*Config, error) {
	path := configdir.LocalConfig()
	err := configdir.MakePath(path)
	if err != nil {
		panic(err)
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
