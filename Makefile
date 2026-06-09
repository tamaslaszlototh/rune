.DEFAULT_GOAL := build

.PHONY: build install clean release snapshot

build:
	go build -o rune ./cmd/rune

install:
	go install ./cmd/rune

clean:
	rm -f rune rune.exe

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean
