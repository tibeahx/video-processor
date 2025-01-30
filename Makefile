.PHONY: build run clean

build:
	go build -o bin/video-processor *.go

run: build
	./bin/video-processor

clean:
	rm -rf bin/ audio/ video/ 