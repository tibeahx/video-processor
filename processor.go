package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type outputConfig struct {
	numElements         float64
	allElements         float64
	lastElementDuration float64
}

type videoChunk struct {
	StartTime  int
	Duration   float64
	OutputPath string
}

type processorParams struct {
	sourcePath string
	chunkSize  int
	baseDir    string
	baseName   string
}

type videoProcessor struct {
	params processorParams
	config outputConfig
}

func newOutputParams(videoPath string, chunkSize int) (outputConfig, error) {
	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return outputConfig{}, err
	}

	numElements := math.Floor(duration / float64(chunkSize))
	lastElementDuration := duration - (numElements * float64(chunkSize))
	allElements := numElements

	if lastElementDuration > 0 {
		allElements++
	}

	return outputConfig{
		numElements:         numElements,
		allElements:         allElements,
		lastElementDuration: lastElementDuration,
	}, nil
}

func newVideoProcessor(sourcePath string, chunkSize int) (*videoProcessor, error) {
	if _, err := os.Stat(sourcePath); err != nil {
		return nil, fmt.Errorf("source file error: %w", err)
	}

	baseDir := filepath.Dir(sourcePath)
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))

	config, err := newOutputParams(sourcePath, chunkSize)
	if err != nil {
		return nil, err
	}

	params := processorParams{
		sourcePath: sourcePath,
		chunkSize:  chunkSize,
		baseDir:    baseDir,
		baseName:   baseName,
	}

	return &videoProcessor{
		params: params,
		config: config,
	}, nil
}

func (vp *videoProcessor) createChunk(chunk videoChunk) error {
	return ffmpeg.Input(vp.params.sourcePath).
		Output(chunk.OutputPath, ffmpeg.KwArgs{
			"ss": chunk.StartTime,
			"t":  chunk.Duration,
			"c":  "copy",
		}).
		OverWriteOutput().
		Run()
}

func (vp *videoProcessor) generateChunkPath(index int) string {
	return filepath.Join(vp.params.baseDir, fmt.Sprintf("%s_chunk_%03d.mp4", vp.params.baseName, index))
}

func (vp *videoProcessor) processRegularChunks() error {
	for i := 0; i < int(vp.config.numElements); i++ {
		chunk := videoChunk{
			StartTime:  i * vp.params.chunkSize,
			Duration:   float64(vp.params.chunkSize),
			OutputPath: vp.generateChunkPath(i),
		}

		if err := vp.createChunk(chunk); err != nil {
			return fmt.Errorf("failed to create chunk %d: %w", i, err)
		}
	}
	return nil
}

func (vp *videoProcessor) processLastChunk() error {
	if vp.config.lastElementDuration <= 0 {
		return nil
	}

	chunk := videoChunk{
		StartTime:  int(vp.config.numElements) * vp.params.chunkSize,
		Duration:   vp.config.lastElementDuration,
		OutputPath: vp.generateChunkPath(int(vp.config.numElements)),
	}

	if err := vp.createChunk(chunk); err != nil {
		return fmt.Errorf("failed to create final chunk: %w", err)
	}
	return nil
}

func (vp *videoProcessor) process() error {
	if err := vp.processRegularChunks(); err != nil {
		return err
	}
	return vp.processLastChunk()
}
