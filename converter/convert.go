package converter

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/mewkiz/pkg/pathutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func GenerateArgs(tags *Tags, total string) []string {
	args := []string{"-f8"}
	if tags.Artist != "" {
		args = append(args, "-T")
		args = append(args, fmt.Sprintf("ARTIST=%s", tags.Artist))
	}
	if tags.Album != "" {
		args = append(args, "-T")
		args = append(args, fmt.Sprintf("ALBUM=%s", tags.Album))
	}
	if tags.AlbumArtist != "" {
		args = append(args, "-T")
		args = append(args, fmt.Sprintf("ALBUMARTIST=%s", tags.AlbumArtist))
	}
	if tags.Title != "" {
		args = append(args, "-T")
		args = append(args, fmt.Sprintf("TITLE=%s", tags.Title))
	}
	if tags.Year != "" {
		args = append(args, "-T")
		args = append(args, fmt.Sprintf("DATE=%s", tags.Year))
	}
	if tags.TrackNumber != "" {
		args = append(args, "-T")
		args = append(args, fmt.Sprintf("TRACKNUMBER=%s", tags.TrackNumber))
	}
	if tags.DiscNumber != "" {
		args = append(args, "-T")
		args = append(args, fmt.Sprintf("DISCNUMBER=%s", tags.DiscNumber))
	} else {
		args = append(args, "-T")
		args = append(args, "DISCNUMBER=1")
	}

	args = append(args, "-T")
	args = append(args, fmt.Sprintf("TOTALTRACKS=%s", total))
	args = append(args, "-T")
	args = append(args, fmt.Sprintf("TRACKTOTAL=%s", total))

	args = append(args, "-T")
	args = append(args, "TOTALDISCS=1")
	args = append(args, "-T")
	args = append(args, "DISCTOTAL=1")

	return args
}

func ConvertFile(path string, total string) error {
	if filepath.Ext(path) != ".m4a" {
		return nil
	}

	re := regexp.MustCompile("bits_per_raw_sample=[0-9]+")

	result, err := exec.Command("ffprobe", "-loglevel", "panic", "-show_entries", "stream=bits_per_raw_sample", path).CombinedOutput()
	if err != nil {
		return err
	}
	sample, found := strings.CutPrefix(string(re.Find(result)), "bits_per_raw_sample=")
	if !found {
		return errors.New("could not find bits_per_raw_sample")
	}

	BitDepth, err := strconv.Atoi(sample)
	if err != nil {
		return err
	}

	WavName := pathutil.TrimExt(path) + ".wav"

	if BitDepth == 32 {
		if err := exec.Command("ffmpeg", "-i", path, "-acodec", "pcm_s32le", WavName).Run(); err != nil {
			return err
		}
	} else {
		if err := exec.Command("ffmpeg", "-i", path, WavName).Run(); err != nil {
			return err
		}
	}

	tags, err := ParseALACTags(path)
	if err != nil {
		return err
	}

	args := GenerateArgs(tags, total)
	args = append(args, WavName)

	if err := exec.Command("flac", args...).Run(); err != nil {
		return err
	}

	if err := os.Remove(WavName); err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return nil
}

func ProcessFiles(dirname string) error {
	files, err := os.ReadDir(dirname)
	if err != nil {
		return err
	}

	counter := 0
	err = filepath.Walk(dirname, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(f.Name()) == ".m4a" {
				counter += 1
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	total := fmt.Sprint(counter)

	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ConvertFile(filepath.Join(dirname, file.Name()), total); err != nil {
				color.Red("Error while processing %s: %s", file.Name(), err.Error())
			}
		}()
	}

	wg.Wait()
	return nil
}
