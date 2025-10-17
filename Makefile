Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X github.com/alexellis/discord-updater/pkg.Version=$(Version) -X github.com/alexellis/discord-updater/pkg.GitCommit=$(GitCommit)"
PLATFORM := $(shell ./hack/platform-tag.sh)
SOURCE_DIRS = main.go launcher.go daemon.go
export GO111MODULE=on

.PHONY: all
all: gofmt build dist hash

.PHONY: build
build:
	go build -ldflags="$(LDFLAGS)" -o bin/discord-updater main.go launcher.go daemon.go

.PHONY: gofmt
gofmt:
	@test -z $(shell gofmt -l -s $(SOURCE_DIRS) ./ |grep -v vendor/| tee /dev/stderr) || (echo "[WARN] Fix formatting issues with 'make gofmt'" && exit 1)

.PHONY: dist
dist:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o bin/discord-updater
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -o bin/discord-updater-arm64

.PHONY: hash
hash:
	rm -rf bin/*.sha256 && ./hack/hashgen.sh

# Show help
help:
	@echo "Available targets:"
	@echo "  all      - Run gofmt, build, dist, hash"
	@echo "  build    - Build optimized binary with -w -s flags"
	@echo "  gofmt    - Check code formatting"
	@echo "  dist     - Build binaries for multiple platforms"
	@echo "  hash     - Generate SHA256 hashes for binaries"
