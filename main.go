package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var folders = []string{"audio", "video"}

func ProcessVideo(sourcePath string) error {
	if err := cleanUp(folders); err != nil {
		return err
	}

	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("source file error: %w", err)
	}

	if err := prepare(folders); err != nil {
		return fmt.Errorf("failed to prepare folders: %w", err)
	}

	audioOut := filepath.Join("audio", strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))+".mp3")
	videoOut := filepath.Join("video", "temp_"+filepath.Base(sourcePath))

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := extractAudio(sourcePath, audioOut); err != nil {
			errChan <- fmt.Errorf("failed to extract audio: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := extractVideo(sourcePath, videoOut); err != nil {
			errChan <- fmt.Errorf("failed to extract video: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	if err := splitVideoIntoChunks(videoOut, 5); err != nil {
		return fmt.Errorf("failed to split video: %w", err)
	}

	os.Remove(videoOut)
	return nil
}

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

func splitVideoIntoChunks(videoPath string, chunkSize int) error {
	probe, err := ffmpeg.Probe(videoPath)
	if err != nil {
		return fmt.Errorf("failed to probe video file: %w", err)
	}
	if probe == "" {
		return errors.New("failed to probe video file: empty probe result")
	}

	baseDir := filepath.Dir(videoPath)
	baseName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	chunkPattern := filepath.Join(baseDir, fmt.Sprintf("%s_chunk_%%03d.mp4", baseName))

	return ffmpeg.Input(videoPath).
		Output(chunkPattern, ffmpeg.KwArgs{
			"c":                   "copy",
			"f":                   "segment",
			"segment_time":        chunkSize,
			"reset_timestamps":    1,
			"segment_format":      "mp4",
			"break_non_keyframes": 1,
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

func main() {
	sourcePath := "source/IMG_8435.MOV"
	if err := ProcessVideo(sourcePath); err != nil {
		log.Fatal(err)
	}
}
