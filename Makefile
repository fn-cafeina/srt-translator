.PHONY: build run clean test

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
