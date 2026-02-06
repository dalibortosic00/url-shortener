BINARY_NAME=url-shortener
BUILD_DIR=bin

.PHONY: all build run clean test test/cover help

all: build

## build: Build the binary for the current OS
build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

## compile: Build binaries for multiple platforms
compile:
	@echo "Compiling for multiple platforms..."
	# 64-bit Linux
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	# 64-bit Windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)
	# macOS (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	@echo "Done! Check the $(BUILD_DIR) folder."

## run: Build and run the application
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test/cover: Run tests and open the coverage report in the browser
test/cover:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out
	@rm coverage.out

## clean: Remove build artifacts and coverage files
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out

## help: Show this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'