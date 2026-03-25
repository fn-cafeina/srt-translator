.PHONY: build run clean test help

BINARY_NAME=srt-translator
BIN_DIR=bin

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/srt-translator

run: build
	@./$(BIN_DIR)/$(BINARY_NAME) $(ARGS)

clean:
	rm -rf $(BIN_DIR)

test:
	go test ./internal/...

help:
	@echo "Available commands:"
	@echo "  build   : Build the binary"
	@echo "  run     : Build and run the binary (use ARGS=\"-input ...\" to pass arguments)"
	@echo "  clean   : Remove the bin directory"
	@echo "  test    : Run tests"
	@echo "  help    : Show this help"
