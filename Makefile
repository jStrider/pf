.PHONY: build run test clean lint

BINARY_NAME=pf
MAIN_PATH=cmd/pf/main.go

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

lint:
	golangci-lint run

install:
	go build -o ~/go/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary installed to ~/go/bin/$(BINARY_NAME)"
	@echo ""
	@echo "Installing shell completions..."
	@# Install zsh completion
	@mkdir -p ~/.oh-my-zsh/custom/completions 2>/dev/null || true
	@~/go/bin/$(BINARY_NAME) completion zsh > ~/.oh-my-zsh/custom/completions/_$(BINARY_NAME) 2>/dev/null || \
		(mkdir -p ~/.zsh/completions && ~/go/bin/$(BINARY_NAME) completion zsh > ~/.zsh/completions/_$(BINARY_NAME))
	@# Install bash completion
	@if [ -d ~/.bash_completion.d ]; then \
		~/go/bin/$(BINARY_NAME) completion bash > ~/.bash_completion.d/$(BINARY_NAME); \
	elif [ -d /usr/local/etc/bash_completion.d ]; then \
		~/go/bin/$(BINARY_NAME) completion bash > /usr/local/etc/bash_completion.d/$(BINARY_NAME) 2>/dev/null || true; \
	fi
	@echo "✓ Shell completions installed"
	@echo ""
	@echo "To use 'pf' command, add ~/go/bin to your PATH:"
	@echo "  echo 'export PATH=\$$PATH:~/go/bin' >> ~/.zshrc"
	@echo "  source ~/.zshrc"
	@echo ""
	@echo "For zsh completions to work, restart your shell or run:"
	@echo "  source ~/.zshrc"