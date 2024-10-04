package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/justjakka/ripper-api-downloader-go/downloader"
	"github.com/urfave/cli/v2"
	"net/http"
	"os"
	"strings"
)

func Rip(_ *cli.Context) error {
	config, err := downloader.CheckConfig()
	if err != nil {
		return err
	}
	client := &http.Client{}

	var links []string

	if len(os.Args) < 2 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("URL: ")
		url, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		url = strings.TrimSpace(url)
		links = append(links, url)
	} else {
		links = os.Args[1:]
	}

	if err := downloader.ProcessDownloads(config, client, links); err != nil {
		return err
	}
	return nil
}

func Start() {
	app := &cli.App{
		Name:        "ripper-downloader",
		Usage:       "downloader written in go",
		UsageText:   "ripper-downloader link1 link2 ...",
		Description: "downloader for my ripper-api implementation written in golang",
		Args:        true,
		ArgsUsage:   "specify amusic links to download",
		Action:      Rip,
	}
	if err := app.Run(os.Args); err != nil {
		color.Red(err.Error())
	}
}

func main() {
	Start()
}
