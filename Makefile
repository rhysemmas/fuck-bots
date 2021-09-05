APP  = fuck-bots

BIN = bin/$(APP)

BIN_LINUX_AMD64  = $(BIN)-linux-amd64
BIN_LINUX_ARM32  = $(BIN)-linux-arm32
BIN_DARWIN_AMD64 = $(BIN)-darwin-amd64

IMAGE   = localhost/$(APP)
CMD_SRC = cmd/$(APP)/main.go

SOURCES = $(shell find . -type f -iname "*.go")

.PHONY: all build vet fmt test run image clean private

all: test build

$(BIN_DARWIN_AMD64): $(SOURCES)
	GOARCH=amd64 GOOS=darwin go build -o $(BIN_DARWIN_AMD64) $(CMD_SRC)

$(BIN_LINUX_ARM32): $(SOURCES)
	GOARCH=arm GOOS=linux CGO_ENABLED=0 go build -o $(BIN_LINUX_ARM32) $(CMD_SRC)

$(BIN_LINUX_AMD64): $(SOURCES)
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o $(BIN_LINUX_AMD64) $(CMD_SRC)

build: $(BIN_DARWIN_AMD64) $(BIN_LINUX_ARM32) $(BIN_LINUX_AMD64) fmt vet

private:
	go env -w GOPRIVATE=github.com/rhysemmas/fuck-bots

vet:
	go vet ./...

fmt: private
	go fmt ./...

test: fmt vet
	go test ./... -coverprofile cover.out

run: fmt vet
	go run $(CMD_SRC) --debug

image: $(BIN_LINUX_AMD64) $(BIN_LINUX_ARM32)
	docker buildx build --platform linux/amd64 -t $(IMAGE):amd64 . -f Dockerfile.amd64
	docker buildx build --platform linux/arm/v7 -t $(IMAGE):arm32 . -f Dockerfile.arm32

clean:
	rm -rf bin/
