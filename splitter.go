package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type splitConfig struct {
	totalChunks       float64
	regularChunks     float64
	lastChunkDuration float64
}

type videoChunk struct {
	StartTime  int
	Duration   float64
	OutputPath string
}

type splitterConfig struct {
	sourcePath string
	chunkSize  int
	baseDir    string
	baseName   string
}

type videoSplitter struct {
	config   splitterConfig
	splitCfg splitConfig
}

func crateSplitConfig(videoPath string, chunkSize int) (splitConfig, error) {
	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return splitConfig{}, err
	}

	regularChunks := math.Floor(duration / float64(chunkSize))
	lastChunkDuration := duration - (regularChunks * float64(chunkSize))
	totalChunks := regularChunks

	if lastChunkDuration > 0 {
		totalChunks++
	}

	return splitConfig{
		totalChunks:       totalChunks,
		regularChunks:     regularChunks,
		lastChunkDuration: lastChunkDuration,
	}, nil
}

func newVideoSplitter(sourcePath string, chunkSize int) (*videoSplitter, error) {
	if _, err := os.Stat(sourcePath); err != nil {
		return nil, fmt.Errorf("source file error: %w", err)
	}

	baseDir := filepath.Dir(sourcePath)
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))

	splitCfg, err := crateSplitConfig(sourcePath, chunkSize)
	if err != nil {
		return nil, err
	}

	config := splitterConfig{
		sourcePath: sourcePath,
		chunkSize:  chunkSize,
		baseDir:    baseDir,
		baseName:   baseName,
	}

	return &videoSplitter{
		config:   config,
		splitCfg: splitCfg,
	}, nil
}

func (vs *videoSplitter) splitChunk(chunk videoChunk) error {
	return runFFmpeg(vs.config.sourcePath, chunk.OutputPath, ffmpeg.KwArgs{
		"ss": chunk.StartTime,
		"t":  chunk.Duration,
		"c":  "copy",
	})
}

func (vs *videoSplitter) generateChunkPath(index int) string {
	return filepath.Join(vs.config.baseDir, fmt.Sprintf("%s_chunk_%03d.mp4", vs.config.baseName, index))
}

func (vs *videoSplitter) splitRegularChunks() error {
	for i := 0; i < int(vs.splitCfg.regularChunks); i++ {
		chunk := videoChunk{
			StartTime:  i * vs.config.chunkSize,
			Duration:   float64(vs.config.chunkSize),
			OutputPath: vs.generateChunkPath(i),
		}

		if err := vs.splitChunk(chunk); err != nil {
			return fmt.Errorf("failed to split chunk %d: %w", i, err)
		}
	}
	return nil
}

func (vs *videoSplitter) splitLastChunk() error {
	if vs.splitCfg.lastChunkDuration <= 0 {
		return nil
	}

	chunk := videoChunk{
		StartTime:  int(vs.splitCfg.regularChunks) * vs.config.chunkSize,
		Duration:   vs.splitCfg.lastChunkDuration,
		OutputPath: vs.generateChunkPath(int(vs.splitCfg.regularChunks)),
	}

	if err := vs.splitChunk(chunk); err != nil {
		return fmt.Errorf("failed to split final chunk: %w", err)
	}
	return nil
}

func (vs *videoSplitter) split() error {
	if err := vs.splitRegularChunks(); err != nil {
		return err
	}
	return vs.splitLastChunk()
}
