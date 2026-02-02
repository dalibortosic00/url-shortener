BINARY_NAME=url-shortener
BUILD_DIR=bin

.PHONY: all build run clean test

all: build

build:
	@echo "Building for local host..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

compile:
	@echo "Compiling for multiple platforms..."
	# 64-bit Linux
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	# 64-bit Windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)
	# macOS (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	@echo "Done! Check the $(BUILD_DIR) folder."

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)