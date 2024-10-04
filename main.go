package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/justjakka/ripper-api-downloader-go/converter"
	"github.com/justjakka/ripper-api-downloader-go/downloader"
	"github.com/urfave/cli/v2"
)

func Rip(_ *cli.Context) error {
	config, err := downloader.CheckConfig()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	client := &http.Client{}

	for i, line := range os.Args[1:] {
		Zipname, err := downloader.Download(config, line, client, i+1, len(os.Args)-1)
		if err != nil {
			color.Red("Error while downloading %s: %s", line, err.Error())
		}
		if config.Unarchive {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := converter.Unzip(Zipname); err != nil {
					color.Red("Error while unzipping %s: %s", Zipname, err.Error())
				}
				if err := os.Remove(Zipname); err != nil {
					color.Red("Error while removing %s: %s", Zipname, err.Error())
				}
			}()
		}
	}

	wg.Wait()
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
