.PHONY: all build test clean install help
.PHONY: build-agent build-grid-cli build-grid-agent-gui build-grid-agent-gui-dev
.PHONY: install-agent install-grid-cli install-grid-agent-gui
.PHONY: test-agent test-grid-cli test-grid-agent-gui
.PHONY: lint lint-agent lint-grid-cli lint-grid-agent-gui
.PHONY: tidy dev-gui dev-gui-browser
.PHONY: coverage work-sync
.PHONY: install-deps install-tools check-deps
.PHONY: build-cli-all build-gui-all build-all-platforms

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Go bin directory (where go install puts tools like wails)
GO_BIN := $(shell go env GOBIN 2>/dev/null || echo "$(shell go env GOPATH)/bin")

# Build information
VERSION := $(shell git tag --sort=-version:refname 2>/dev/null | head -n 1)
ifeq ($(VERSION),)
    VERSION := dev
endif
COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Directories
AGENT_DIR=agent
GRID_CLI_DIR=grid-cli
GUI_DIR=grid-agent-gui
CLI_BIN_DIR=$(GRID_CLI_DIR)/build/bin
DIST_DIR=dist

# Detect OS for platform-specific settings
HOST_OS := $(shell uname -s 2>/dev/null)

# Install paths (platform-specific)
ifeq ($(HOST_OS),Darwin)
    INSTALL_PREFIX=/usr/local
else ifeq ($(HOST_OS),Linux)
    INSTALL_PREFIX=$(HOME)/.local
endif

INSTALL_BIN=$(INSTALL_PREFIX)/bin

# Colors for terminal output
CYAN=\033[0;36m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

all: build ## Build all components

help: ## Show this help message
	@echo "$(CYAN)Grid Agent Makefile Commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-25s$(NC) %s\n", $$1, $$2}'
	@echo ""

#########################
# Build Targets
#########################

build: build-agent build-grid-cli build-grid-agent-gui ## Build all components

build-agent: ## Build agent package
	@echo "$(CYAN)Building agent...$(NC)"
	@cd $(AGENT_DIR) && $(GOBUILD) ./...
	@echo "$(GREEN)✓ Agent built successfully$(NC)"

build-grid-cli: ## Build grid-cli (tfcmd)
	@echo "$(CYAN)Building grid-cli (tfcmd)...$(NC)"
	@mkdir -p $(CLI_BIN_DIR)
	@cd $(GRID_CLI_DIR) && $(GOBUILD) -ldflags \
		"-X github.com/threefoldtech/grid-agent/grid-cli/cmd.commit=$(COMMIT) \
		 -X github.com/threefoldtech/grid-agent/grid-cli/cmd.version=$(VERSION) \
		 -X github.com/threefoldtech/grid-agent/grid-cli/cmd.date=$(BUILD_DATE)" \
		-o ../$(CLI_BIN_DIR)/tfcmd main.go
	@echo "$(GREEN)✓ Grid CLI built successfully → $(CLI_BIN_DIR)/tfcmd$(NC)"

build-grid-agent-gui: ## Build grid-agent-gui (Wails application)
	@echo "$(CYAN)Building grid-agent-gui...$(NC)"
	@cd $(GUI_DIR) && PATH="$(GO_BIN):$$PATH" wails build -tags webkit2_41 -ldflags "-X main.Version=$(VERSION)"
	@echo "$(GREEN)✓ Grid Agent GUI built successfully$(NC)"

build-grid-agent-gui-dev: ## Build grid-agent-gui in development mode
	@echo "$(CYAN)Building grid-agent-gui (dev mode)...$(NC)"
	@cd $(GUI_DIR) && wails build -debug -tags webkit2_41 -ldflags "-X main.Version=$(VERSION)"
	@echo "$(GREEN)✓ Grid Agent GUI built successfully (dev mode)$(NC)"

#########################
# Install Targets
#########################

install: install-grid-cli install-grid-agent-gui ## Install all components

install-agent: build-agent ## Install agent package
	@echo "$(CYAN)Installing agent...$(NC)"
	@echo "$(YELLOW)Note: Agent is a library package, no binary to install$(NC)"

install-grid-cli: build-grid-cli ## Install grid-cli to system
	@echo "$(CYAN)Installing tfcmd to $(INSTALL_BIN)...$(NC)"
	@mkdir -p $(INSTALL_BIN)
	@cp $(CLI_BIN_DIR)/tfcmd $(INSTALL_BIN)/tfcmd
	@chmod +x $(INSTALL_BIN)/tfcmd
	@echo "$(GREEN)✓ tfcmd installed successfully to $(INSTALL_BIN)/tfcmd$(NC)"
	@echo "$(YELLOW)Make sure $(INSTALL_BIN) is in your PATH$(NC)"

install-grid-agent-gui: build-grid-agent-gui ## Install grid-agent-gui to system
ifeq ($(HOST_OS),Darwin)
	@echo "$(CYAN)Installing grid-agent-gui.app to /Applications...$(NC)"
	@rm -rf "/Applications/TF Grid Agent.app"
	@cp -r $(GUI_DIR)/build/bin/grid-agent-gui.app "/Applications/TF Grid Agent.app"
	@echo "$(GREEN)✓ Grid Agent GUI installed successfully to /Applications/TF Grid Agent.app$(NC)"
else ifeq ($(HOST_OS),Linux)
	@echo "$(CYAN)Installing grid-agent-gui to $(INSTALL_BIN)...$(NC)"
	@mkdir -p $(INSTALL_BIN)
	@cp $(GUI_DIR)/build/bin/grid-agent-gui $(INSTALL_BIN)/grid-agent-gui
	@chmod +x $(INSTALL_BIN)/grid-agent-gui
	@echo "$(CYAN)Installing desktop entry...$(NC)"
	@mkdir -p $(HOME)/.local/share/applications
	@mkdir -p $(HOME)/.local/share/icons/hicolor/512x512/apps
	@cp $(GUI_DIR)/build/appicon.png $(HOME)/.local/share/icons/hicolor/512x512/apps/grid-agent.png
	@echo "[Desktop Entry]\n\
Type=Application\n\
Name=ThreeFold Grid Agent\n\
Comment=AI-powered interface for ThreeFold Grid\n\
Exec=$(INSTALL_BIN)/grid-agent-gui\n\
Icon=grid-agent\n\
Terminal=false\n\
Categories=Utility;Development;" > $(HOME)/.local/share/applications/grid-agent.desktop
	@chmod +x $(HOME)/.local/share/applications/grid-agent.desktop
	@gtk-update-icon-cache $(HOME)/.local/share/icons/hicolor/ -f 2>/dev/null || true
	@update-desktop-database $(HOME)/.local/share/applications/ 2>/dev/null || true
	@echo "$(GREEN)✓ Grid Agent GUI installed successfully to $(INSTALL_BIN)/grid-agent-gui$(NC)"
endif

#########################
# Test Targets
#########################

test: test-agent test-grid-cli test-grid-agent-gui ## Run all tests
	@echo "$(GREEN)✓ All tests passed$(NC)"

test-agent: ## Test agent package
	@echo "$(CYAN)Testing agent...$(NC)"
	@cd $(AGENT_DIR) && $(GOTEST) -v ./...

test-grid-cli: ## Test grid-cli
	@echo "$(CYAN)Testing grid-cli...$(NC)"
	@cd $(GRID_CLI_DIR) && $(GOTEST) -v ./...

test-grid-agent-gui: ## Test grid-agent-gui
	@echo "$(CYAN)Testing grid-agent-gui...$(NC)"
	@cd $(GUI_DIR) && $(GOTEST) -v ./...

#########################
# Lint Targets
#########################

lint: lint-agent lint-grid-cli lint-grid-agent-gui ## Lint all components
	@echo "$(GREEN)✓ All components linted$(NC)"

lint-agent: ## Lint agent package
	@echo "$(CYAN)Linting agent...$(NC)"
	@cd $(AGENT_DIR) && golangci-lint run

lint-grid-cli: ## Lint grid-cli
	@echo "$(CYAN)Linting grid-cli...$(NC)"
	@cd $(GRID_CLI_DIR) && golangci-lint run

lint-grid-agent-gui: ## Lint grid-agent-gui
	@echo "$(CYAN)Linting grid-agent-gui...$(NC)"
	@cd $(GUI_DIR) && golangci-lint run

#########################
# Development Targets
#########################

dev-gui: ## Run grid-agent-gui in development mode with hot reload
	@echo "$(CYAN)Starting grid-agent-gui in dev mode...$(NC)"
	@cd $(GUI_DIR) && PATH="$(GO_BIN):$$PATH" wails dev -tags webkit2_41 -ldflags "-X main.Version=$(VERSION)"

dev-gui-browser: ## Run grid-agent-gui in development mode with hot reload in the browser
	@echo "$(CYAN)Starting grid-agent-gui in dev mode...$(NC)"
	@cd $(GUI_DIR) && PATH="$(GO_BIN):$$PATH" wails dev -browser -tags webkit2_41 -ldflags "-X main.Version=$(VERSION)"

tidy: ## Tidy all go.mod files
	@echo "$(CYAN)Tidying all modules...$(NC)"
	@cd $(AGENT_DIR) && $(GOMOD) tidy
	@cd $(GRID_CLI_DIR) && $(GOMOD) tidy
	@cd $(GUI_DIR) && $(GOMOD) tidy
	@$(GOCMD) work sync
	@echo "$(GREEN)✓ All modules tidied$(NC)"

#########################
# Coverage Targets
#########################

coverage: ## Generate coverage report for all modules
	@echo "$(CYAN)Generating coverage reports...$(NC)"
	@mkdir -p coverage
	@cd $(AGENT_DIR) && $(GOTEST) -v -coverprofile=../coverage/agent-coverage.out ./...
	@cd $(GRID_CLI_DIR) && $(GOTEST) -v -coverprofile=../coverage/grid-cli-coverage.out ./...
	@cd $(GUI_DIR) && $(GOTEST) -v -coverprofile=../coverage/gui-coverage.out ./...
	@$(GOCMD) tool cover -html=coverage/agent-coverage.out -o coverage/agent-coverage.html
	@$(GOCMD) tool cover -html=coverage/grid-cli-coverage.out -o coverage/grid-cli-coverage.html
	@$(GOCMD) tool cover -html=coverage/gui-coverage.out -o coverage/gui-coverage.html
	@echo "$(GREEN)✓ Coverage reports generated in coverage/$(NC)"

#########################
# Clean Targets
#########################

clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	@rm -rf $(CLI_BIN_DIR)
	@rm -rf coverage
	@rm -rf $(GUI_DIR)/build/bin
	@rm -f $(GUI_DIR)/grid-agent-gui
	@rm -rf $(DIST_DIR)
	@echo "$(GREEN)✓ Clean complete$(NC)"

#########################
# Workspace Targets
#########################

work-sync: ## Sync Go workspace
	@echo "$(CYAN)Syncing Go workspace...$(NC)"
	@$(GOMOD) work sync
	@echo "$(GREEN)✓ Workspace synced$(NC)"

#########################
# Setup Targets
#########################

install-deps: ## Install system dependencies (Linux/Ubuntu)
ifeq ($(HOST_OS),Linux)
	@echo "$(CYAN)Installing system dependencies...$(NC)"
	@if command -v apt >/dev/null; then \
		sudo apt update && sudo apt install -y \
		libgtk-3-dev \
		libwebkit2gtk-4.1-dev \
		libsoup-3.0-dev \
		libjavascriptcoregtk-4.1-dev \
		build-essential \
		pkg-config; \
		echo "$(GREEN)✓ System dependencies installed$(NC)"; \
	else \
		echo "$(YELLOW)Warning: 'apt' not found. Please install dependencies manually for your distribution.$(NC)"; \
	fi
else
	@echo "$(YELLOW)System dependency installation is currently only supported for Linux (apt).$(NC)"
endif

install-tools: ## Install development tools
	@echo "$(CYAN)Installing development tools...$(NC)"
	@echo "Installing Wails v2.11.0..."
	@$(GOINSTALL) github.com/wailsapp/wails/v2/cmd/wails@v2.11.0
	@echo "Installing golangci-lint..."
	@$(GOINSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)✓ Development tools installed$(NC)"

check-deps: ## Check if required dependencies are installed
	@echo "$(CYAN)Checking dependencies...$(NC)"
	@command -v go > /dev/null 2>&1 || { echo "$(YELLOW)✗ Go is not installed$(NC)"; exit 1; }
	@echo "$(GREEN)✓ Go $(shell go version)$(NC)"
	@command -v wails > /dev/null 2>&1 || { echo "$(YELLOW)✗ Wails is not installed. Run 'make install-tools'$(NC)"; exit 1; }
	@echo "$(GREEN)✓ Wails $(shell wails version)$(NC)"
	@command -v node > /dev/null 2>&1 || { echo "$(YELLOW)✗ Node.js is not installed$(NC)"; exit 1; }
	@echo "$(GREEN)✓ Node.js $(shell node --version)$(NC)"
	@echo "$(GREEN)All dependencies are installed!$(NC)"

#########################
# Cross-Compilation Targets
#########################

build-cli-all: ## Build CLI for all platforms
	@echo "$(CYAN)Building CLI for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@cd $(GRID_CLI_DIR) && \
		LDFLAGS="-X github.com/threefoldtech/grid-agent/grid-cli/cmd.commit=$(COMMIT) \
		         -X github.com/threefoldtech/grid-agent/grid-cli/cmd.version=$(VERSION) \
		         -X github.com/threefoldtech/grid-agent/grid-cli/cmd.date=$(BUILD_DATE)"; \
		echo "  Building linux/amd64..."; \
		GOOS=linux GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o ../$(DIST_DIR)/tfcmd-linux-amd64 .; \
		echo "  Building windows/amd64..."; \
		GOOS=windows GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o ../$(DIST_DIR)/tfcmd-windows-amd64.exe .; \
		echo "  Building darwin/amd64..."; \
		GOOS=darwin GOARCH=amd64 go build -ldflags "$$LDFLAGS" -o ../$(DIST_DIR)/tfcmd-darwin-amd64 .; \
		echo "  Building darwin/arm64..."; \
		GOOS=darwin GOARCH=arm64 go build -ldflags "$$LDFLAGS" -o ../$(DIST_DIR)/tfcmd-darwin-arm64 .
	@echo "$(GREEN)✓ CLI built for all platforms → $(DIST_DIR)/$(NC)"

build-gui-all: ## Build GUI for all platforms (requires appropriate OS)
	@echo "$(CYAN)Building GUI for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@cd $(GUI_DIR) && \
		echo "  Building windows/amd64..."; \
		PATH="$(GO_BIN):$$PATH" wails build -platform windows/amd64 -tags webkit2_41 -ldflags "-X main.Version=$(VERSION)" -o ../../../$(DIST_DIR)/grid-agent-gui-windows-amd64.exe
ifeq ($(HOST_OS),Darwin)
	@cd $(GUI_DIR) && \
		echo "  Building darwin/amd64..."; \
		PATH="$(GO_BIN):$$PATH" wails build -platform darwin/amd64 -tags webkit2_41 -ldflags "-X main.Version=$(VERSION)" -o ../../../$(DIST_DIR)/grid-agent-gui-darwin-amd64; \
		cp -r build/bin/grid-agent-gui.app ../dist/grid-agent-gui-darwin-amd64.app
	@cd $(GUI_DIR) && \
		echo "  Building darwin/arm64..."; \
		PATH="$(GO_BIN):$$PATH" wails build -platform darwin/arm64 -tags webkit2_41 -ldflags "-X main.Version=$(VERSION)" -o ../../../$(DIST_DIR)/grid-agent-gui-darwin-arm64; \
		cp -r build/bin/grid-agent-gui.app ../dist/grid-agent-gui-darwin-arm64.app
endif
ifeq ($(HOST_OS),Linux)
		echo "  Building linux/amd64..."; \
		PATH="$(GO_BIN):$$PATH" wails build -platform linux/amd64 -tags webkit2_41 -ldflags "-X main.Version=$(VERSION)" -o ../../../$(DIST_DIR)/grid-agent-gui-linux-amd64
endif
	@echo "$(GREEN)✓ GUI built for available platforms → $(DIST_DIR)/$(NC)"

build-all-platforms: build-cli-all build-gui-all ## Build all binaries for all platforms
	@echo "$(GREEN)✓ All binaries built successfully$(NC)"
	@echo "$(CYAN)Binaries available in $(DIST_DIR)/:$(NC)"
	@ls -lh $(DIST_DIR)/
