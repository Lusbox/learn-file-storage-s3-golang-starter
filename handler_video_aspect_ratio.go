package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)


func getVideoAspectRatio(filepath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)

	var buffer bytes.Buffer
	cmd.Stdout = &buffer

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("unable to run command")
	}

	type dimensions struct {
		Streams []struct {
			Width int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	videoDimensions := dimensions{}
	if err := json.Unmarshal(buffer.Bytes(), &videoDimensions); err != nil {
		return "", fmt.Errorf("unable to unmarshal data")
	}

	aspectRatio := float64(videoDimensions.Streams[0].Width)/float64(videoDimensions.Streams[0].Height)

	if aspectRatio > 1 {
		return "16:9", nil
	} else if aspectRatio < 1 {
		return "9:16", nil
	} else {
		return "other", nil
	}
	
}