package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var folders = []string{"audio", "video"}

func main() {
	sourcePath := "source/IMG_8435.MOV"
	if err := processInput(sourcePath); err != nil {
		log.Fatal(err)
	}
}

func processInput(sourcePath string) error {
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

	processor, err := newVideoProcessor(videoOut, 7)
	if err != nil {
		return fmt.Errorf("failed to create video processor: %w", err)
	}

	if err := processor.process(); err != nil {
		return fmt.Errorf("failed to process video: %w", err)
	}

	if err := os.Remove(videoOut); err != nil {
		return fmt.Errorf("failed to remove temporary video file: %w", err)
	}

	return nil
}
