BIN_NAME=otori
BIN_DIR=bin
CMD_DIR=./cmd/otori
INSTALL_DIR=$(HOME)/.local/bin
OTORI_DIR=$(HOME)/.otori

.PHONY: all build run clean install uninstall help

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BIN_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)

clean:
	rm -rf $(BIN_DIR)

install: build
	@echo "Installing otori..."
	@mkdir -p $(OTORI_DIR)
	@mkdir -p $(OTORI_DIR)/profiles
	@if [ -d "cowrie-honeyfs-base" ]; then \
		cp -r cowrie-honeyfs-base $(OTORI_DIR)/; \
		echo "  [+] Copied cowrie-honeyfs-base to $(OTORI_DIR)/"; \
	fi
	@mkdir -p $(INSTALL_DIR)
	@cp $(BIN_DIR)/$(BIN_NAME) $(INSTALL_DIR)/
	@echo "  [+] Installed binary to $(INSTALL_DIR)/$(BIN_NAME)"
	@echo ""
	@echo "Installation complete!"
	@echo ""
	@if echo "$$PATH" | grep -q "$(INSTALL_DIR)"; then \
		echo "You can now use 'otori' from anywhere."; \
	else \
		echo "Add this to your ~/.bashrc or ~/.zshrc:"; \
		echo "  export PATH=\"$(INSTALL_DIR):\$$PATH\""; \
		echo ""; \
		echo "Then run: source ~/.bashrc (or restart your terminal)"; \
	fi

uninstall:
	@echo "Uninstalling otori..."
	@rm -f $(INSTALL_DIR)/$(BIN_NAME)
	@echo "  [+] Removed $(INSTALL_DIR)/$(BIN_NAME)"
	@echo ""
	@echo "Note: $(OTORI_DIR) was kept (contains your profiles)."
	@echo "To remove completely: rm -rf $(OTORI_DIR)"

help:
	@echo "Available commands:"
	@echo "  build     - Build the otori binary"
	@echo "  run       - Run otori with go run"
	@echo "  clean     - Remove build artifacts"
	@echo "  install   - Build and install otori to ~/.local/bin"
	@echo "  uninstall - Remove otori from ~/.local/bin"

