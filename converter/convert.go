package converter

import (
	"errors"
	"github.com/fatih/color"
	"github.com/mewkiz/pkg/pathutil"
	"github.com/wtolson/go-taglib"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func CopyTags(AlacPath string) error {
	FlacPath := pathutil.TrimExt(AlacPath) + ".flac"

	AlacFile, err := taglib.Read(AlacPath)
	if err != nil {
		return err
	}
	defer AlacFile.Close()

	FlacFile, err := taglib.Read(FlacPath)
	if err != nil {
		return err
	}

	defer FlacFile.Close()

	FlacFile.SetAlbum(AlacFile.Album())
	FlacFile.SetArtist(AlacFile.Artist())
	FlacFile.SetTitle(AlacFile.Title())
	FlacFile.SetYear(AlacFile.Year())
	FlacFile.SetTrack(AlacFile.Track())
	FlacFile.SetGenre(AlacFile.Genre())

	if err := FlacFile.Save(); err != nil {
		return err
	}

	return nil
}

func ConvertFile(path string) error {
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
	if err := exec.Command("flac", "-8", "-f", WavName).Run(); err != nil {
		return err
	}

	if err := os.Remove(WavName); err != nil {
		return err
	}

	if err := CopyTags(path); err != nil {
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

	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ConvertFile(filepath.Join(dirname, file.Name())); err != nil {
				color.Red("Error while processing %s: %s", file.Name(), err.Error())
			}
		}()
	}

	wg.Wait()
	return nil
}
