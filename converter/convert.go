package converter

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func ConvertFile(path string) error {
	//basename := strings.TrimSuffix(path, filepath.Ext(path))

	ffprobeBitdepth := exec.Command(fmt.Sprintf("ffprobe -loglevel panic -show_entries stream=bits_per_raw_sample -select_streams a %s", path))
	if errors.Is(ffprobeBitdepth.Err, exec.ErrDot) {
		ffprobeBitdepth.Err = nil
	}

	result, err := ffprobeBitdepth.Output()
	if err != nil {
		return err
	}

	output := string(result)

	fmt.Println(output)
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
			err := ConvertFile(filepath.Join(dirname, file.Name()))
			if err != nil {
				color.Red("Error while processing %s: %s", file.Name(), err.Error())
			}
		}()
	}

	wg.Wait()
	return nil
}
