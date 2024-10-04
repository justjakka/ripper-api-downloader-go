package main

import (
	"fmt"
	"os/exec"
)

func main() {
	result, err := exec.Command("ffprobe", "-loglevel panic", "-show_entries stream=bits_per_raw_sample", "-select_streams a", "1.m4a").Output()
	output := string(result)
	if err != nil {
		fmt.Println(output)
		fmt.Println(err.Error())
	}

	fmt.Println(output)
}
