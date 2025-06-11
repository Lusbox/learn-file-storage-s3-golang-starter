package main

import (
	"fmt"
	"os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := filePath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("unable to run command")
	}

	return outputPath, nil
}