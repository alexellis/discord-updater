.PHONY: all build

# Build flags for smaller binary
LDFLAGS := -w -s

# Default target
all: build

# Build the main binary with optimizations
build:
	go build -ldflags="$(LDFLAGS)" -o bin/discord-updater main.go launcher.go daemon.go

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build optimized binary with -w -s flags"
