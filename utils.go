package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func runFFmpeg(input, output string, args ffmpeg.KwArgs) error {
	return ffmpeg.Input(input).
		Output(output, args).
		OverWriteOutput().
		Run()
}

func extractAudio(source, output string) error {
	return runFFmpeg(source, output, ffmpeg.KwArgs{
		"vn":     "",
		"acodec": "libmp3lame",
	})
}

func extractVideo(source, output string) error {
	return runFFmpeg(source, output, ffmpeg.KwArgs{
		"c:v": "copy",
		"an":  "",
	})
}

func getVideoDuration(videoPath string) (float64, error) {
	probe, err := ffmpeg.Probe(videoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to probe video: %w", err)
	}
	if probe == "" {
		return 0, errors.New("empty probe result")
	}

	var probeWrapper struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	if err := json.Unmarshal([]byte(probe), &probeWrapper); err != nil {
		return 0, fmt.Errorf("failed to parse probe data: %w", err)
	}

	duration, err := strconv.ParseFloat(probeWrapper.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

func prepare(folders []string) error {
	if folders == nil {
		return errors.New("nil folders")
	}

	for _, folder := range folders {
		if err := os.MkdirAll(folder, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", folder, err)
		}
	}

	return nil
}

func cleanUp(dirsToClean []string) error {
	for _, dir := range dirsToClean {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", dir, err)
		}
	}

	return nil
}
