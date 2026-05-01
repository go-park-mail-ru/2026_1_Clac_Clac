BINARY_NAME=main
MAIN_PATH=./cmd/api/main.go
BUILD_DIR=bin

.PHONY: build run

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)
