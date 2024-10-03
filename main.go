package main

import (
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/justjakka/ripper-api-downloader-go/downloader"
	"github.com/urfave/cli/v2"
)

func Rip(ctx *cli.Context) error {
	config, err := downloader.CheckConfig()
	if err != nil {
		return err
	}

	client := &http.Client{}

	for _, line := range os.Args[1:] {
		err := downloader.Download(config, line, client)
		if err != nil {
			color.Red("Error while downloading %v: %v", line, err.Error())
		}
	}

	return nil
}

func Start() {
	app := &cli.App{
		Name:        "ripper-downloader",
		Usage:       "downloader writter in go",
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
