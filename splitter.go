package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type chunkOptions struct {
	totalChunks       float64
	regularChunks     float64
	lastChunkDuration float64
}

type videoSplitter struct {
	params splitterParams
	opts   chunkOptions
}

func newSplitConfig(videoPath string, chunkSize int) (chunkOptions, error) {
	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return chunkOptions{}, err
	}

	regularChunks := math.Floor(duration / float64(chunkSize))
	lastChunkDuration := duration - (regularChunks * float64(chunkSize))
	totalChunks := regularChunks

	if lastChunkDuration > 0 {
		totalChunks++
	}

	return chunkOptions{
		totalChunks:       totalChunks,
		regularChunks:     regularChunks,
		lastChunkDuration: lastChunkDuration,
	}, nil
}

type splitterParams struct {
	sourcePath string
	chunkSize  int
	baseDir    string
	baseName   string
}

func newVideoSplitter(sourcePath string, chunkSize int) (*videoSplitter, error) {
	if _, err := os.Stat(sourcePath); err != nil {
		return nil, fmt.Errorf("source file error: %w", err)
	}

	baseDir := filepath.Dir(sourcePath)
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))

	splitCfg, err := newSplitConfig(sourcePath, chunkSize)
	if err != nil {
		return nil, err
	}

	p := splitterParams{
		sourcePath: sourcePath,
		chunkSize:  chunkSize,
		baseDir:    baseDir,
		baseName:   baseName,
	}

	return &videoSplitter{
		params: p,
		opts:   splitCfg,
	}, nil
}

type videoChunk struct {
	StartTime  int
	Duration   float64
	OutputPath string
}

func (vs *videoSplitter) splitChunk(chunk videoChunk) error {
	return runFFmpeg(vs.params.sourcePath, chunk.OutputPath, ffmpeg.KwArgs{
		"ss": chunk.StartTime,
		"t":  chunk.Duration,
		"c":  "copy",
	})
}

func (vs *videoSplitter) generateChunkPath(index int) string {
	return filepath.Join(vs.params.baseDir, fmt.Sprintf("%s_chunk_%03d.mp4", vs.params.baseName, index))
}

func (vs *videoSplitter) splitRegularChunks() error {
	for i := 0; i < int(vs.opts.regularChunks); i++ {
		chunk := videoChunk{
			StartTime:  i * vs.params.chunkSize,
			Duration:   float64(vs.params.chunkSize),
			OutputPath: vs.generateChunkPath(i),
		}

		if err := vs.splitChunk(chunk); err != nil {
			return fmt.Errorf("failed to split chunk %d: %w", i, err)
		}
	}
	return nil
}

func (vs *videoSplitter) splitLastChunk() error {
	if vs.opts.lastChunkDuration <= 0 {
		return nil
	}

	chunk := videoChunk{
		StartTime:  int(vs.opts.regularChunks) * vs.params.chunkSize,
		Duration:   vs.opts.lastChunkDuration,
		OutputPath: vs.generateChunkPath(int(vs.opts.regularChunks)),
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
