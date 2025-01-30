package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func extractAudio(source, output string) error {
	return ffmpeg.Input(source).
		Output(output, ffmpeg.KwArgs{
			"vn":     "",
			"acodec": "libmp3lame",
		}).
		OverWriteOutput().
		Run()
}

func extractVideo(source, output string) error {
	return ffmpeg.Input(source).
		Output(output, ffmpeg.KwArgs{
			"c:v": "copy",
			"an":  "",
		}).
		OverWriteOutput().
		Run()
}

func getVideoDuration(videoPath string) (float64, error) {
	probe, err := ffmpeg.Probe(videoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to probe video: %w", err)
	}
	if probe == "" {
		return 0, errors.New("empty probe result")
	}

	var probeData struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	if err := json.Unmarshal([]byte(probe), &probeData); err != nil {
		return 0, fmt.Errorf("failed to parse probe data: %w", err)
	}

	duration, err := strconv.ParseFloat(probeData.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}
