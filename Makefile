BIN_NAME=otori
BIN_DIR=bin
CMD_DIR=./cmd/otori

.PHONY: all build run clean help

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BIN_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)

clean:
	rm -rf $(BIN_DIR)

help:
	@echo "Available commands:"
	@echo " build     - Build the otori binary"
	@echo " run       - Run otori with go run"
	@echo " clean     - Remove build artifact"

