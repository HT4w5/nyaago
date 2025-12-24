BINARY_NAME=nyaago
GO_FILES=$(shell find . -name "*.go")

VERSION=$(shell cat .version)
COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u)
GO_VERSION=$(shell go version | cut -d " " -f 3)

LDFLAGS_BASE=-X github.com/HT4w5/nyaago/pkg/meta.Version=$(VERSION) -X github.com/HT4w5/nyaago/pkg/meta.CommitHash=$(COMMIT) -X github.com/HT4w5/nyaago/pkg/meta.GoVersion=$(GO_VERSION) -X 'github.com/HT4w5/nyaago/pkg/meta.BuildDate=$(BUILD_DATE)'

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build run test clean build-all

all: build

# Build for current native platform
build:
	@echo "Building $(BINARY_NAME) for host..."
	go build -ldflags "$(LDFLAGS_BASE) -X github.com/HT4w5/nyaago/pkg/meta.Platform=$(shell go env GOOS)/$(shell go env GOARCH)" -o bin/$(BINARY_NAME) cmd/nyaago/nyaago.go

# Multi-arch build target
build-all:
	@$(foreach PLATFORM,$(PLATFORMS), \
		OS=$(word 1,$(subst /, ,$(PLATFORM))); \
		ARCH=$(word 2,$(subst /, ,$(PLATFORM))); \
		SUFFIX=""; \
		if [ "$$OS" = "windows" ]; then SUFFIX=".exe"; fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build \
			-ldflags "$(LDFLAGS_BASE) -X github.com/HT4w5/nyaago/pkg/meta.Platform=$$OS/$$ARCH" \
			-o bin/$(BINARY_NAME)-$$OS-$$ARCH$$SUFFIX cmd/nyaago/nyaago.go; \
	)

run: build
	./bin/$(BINARY_NAME)

test:
	@echo "Running tests..."
	go test -v -race -cover ./...

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	go clean
