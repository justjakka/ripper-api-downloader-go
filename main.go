package main

import (
	"fmt"
	"os"

	"github.com/justjakka/ripper-api-downloader-go/downloader"
	"github.com/urfave/cli/v2"
)

func Rip(ctx *cli.Context) error {
	_, err := downloader.CheckConfig()
	if err != nil {
		return err
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
		fmt.Println(err.Error())
	}
}

func main() {
	Start()
}
