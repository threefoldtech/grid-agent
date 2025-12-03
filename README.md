# Grid Agent

AI-powered agent framework and tools for interacting with the ThreeFold Grid through natural language.

<div align="center">
  <img src="grid-agent-gui/build/appicon.png" alt="ThreeFold Grid Agent" width="400px">
</div>

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Overview

The **Grid Agent** repository provides a complete suite of tools for intelligent automation and management of ThreeFold Grid infrastructure:

- **🤖 Agent Framework** - Reusable AI agent core with LLM provider abstraction
- **🖥️ Grid Agent GUI** - Modern desktop application built with Wails
- **⚡ Grid CLI (tfcmd)** - Command-line interface for Grid operations

## Components

### 1. 📦 Agent (`/agent`)

A standalone, reusable AI agent framework for building intelligent conversational assistants.

**Key Features:**
- LLM provider abstraction (Google Gemini, extensible to OpenAI, etc.)
- Extensible tool system for CLI integration
- Real-time streaming command execution
- Automatic retry logic and error handling
- Structured JSON response parsing

**Package:** `github.com/threefoldtech/grid-agent/agent`

[📖 Read the full Agent documentation →](./agent/README.md)

---

### 2. 🖥️ Grid Agent GUI (`/grid-agent-gui`)

A modern, cross-platform desktop application providing a chat interface to interact with the ThreeFold Grid using natural language.

**Key Features:**
- AI-powered chat interface with Google Gemini
- Beautiful dark/light theme support
- Secure onboarding flow
- Real-time command execution with streaming output
- Multi-network support (mainnet, testnet, devnet)

**Technologies:** Wails v2, Svelte, TypeScript, Go

[📖 Read the full GUI documentation →](./grid-agent-gui/README.md)

---

### 3. ⚡ Grid CLI (`/grid-cli`)

Command-line interface for ThreeFold Grid operations, providing the foundation for agent schema generation.

**Key Features:**
- Deploy and manage VMs, Kubernetes clusters, databases
- Gateway management (FQDN, Name)
- Contract operations
- Schema export for AI agent integration

**Package:** `github.com/threefoldtech/grid-agent/grid-cli`

[📖 Read the full CLI documentation →](./grid-cli/README.md)

---

## Quick Start

### Prerequisites

- **Go** 1.23 or higher
- **Node.js** 18 or higher (for GUI)
- **Wails CLI** v2.11.0+ (for GUI development)
- **System Dependencies** (Linux only): Run `make install-deps`

### Installation

Install all components with a single command:

```bash
make install
```

Or install specific components:

```bash
# Install Grid CLI only
make install-grid-cli

# Install Grid Agent GUI only
make install-grid-agent-gui
```

This will install binaries to:
- **Linux:** `~/.local/bin`
- **macOS:** `/usr/local/bin`

> **Note:** Make sure the installation directory is in your `PATH`.

### Using the GUI

After installation, launch the Grid Agent GUI:

```bash
# From terminal
grid-agent-gui

# Or use your application launcher
# Search for "ThreeFold Grid Agent" in your app menu
```

On first launch, you'll be guided through:
1. Entering your mnemonic phrase
2. Selecting a network (mainnet/testnet/devnet)
3. Providing a Gemini API key

Get a free Gemini API key at: https://aistudio.google.com/app/apikey

---

## Development

This repository uses Go workspaces to manage multiple modules.

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/threefoldtech/grid-agent.git
cd grid-agent

# Check dependencies
make check-deps

# Install development tools
make install-tools

# Sync workspace
make work-sync
```

### Build from Source

```bash
# Build all components
make build

# Build specific components
make build-agent          # Build agent library
make build-grid-cli       # Build tfcmd binary
make build-grid-agent-gui # Build GUI application
```

### Running Tests

```bash
# Run all tests
make test

# Run specific component tests
make test-agent
make test-grid-cli
make test-grid-agent-gui

# Generate coverage reports
make coverage
```

### Linting

```bash
# Lint all components
make lint

# Lint specific components
make lint-agent
make lint-grid-cli
make lint-grid-agent-gui
```

### GUI Development

Run the GUI in development mode with hot reload:

```bash
make dev-gui

# Or directly
cd grid-agent-gui
wails dev
```

This starts:
- Vite dev server with hot reload
- Go backend
- Dev server at http://localhost:34115

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Grid Agent GUI (Wails)                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Frontend (Svelte + TypeScript)                        │ │
│  │  - Chat Interface                                      │ │
│  │  - Settings Management                                 │ │
│  │  - Real-time Output Display                           │ │
│  └──────────────────────┬─────────────────────────────────┘ │
│                         │ Wails Runtime Bridge              │
│  ┌──────────────────────▼─────────────────────────────────┐ │
│  │  Backend (Go)                                          │ │
│  │  - Agent Integration                                   │ │
│  │  - tfcmd Schema Generation                            │ │
│  │  - Configuration Management                           │ │
│  └──────────────────────┬─────────────────────────────────┘ │
└─────────────────────────┼───────────────────────────────────┘
                          │
            ┌─────────────▼─────────────┐
            │   Agent Framework Core    │
            │  ┌─────────────────────┐  │
            │  │  LLM Provider       │  │
            │  │  (Gemini)           │  │
            │  └──────────┬──────────┘  │
            │             │              │
            │  ┌──────────▼──────────┐  │
            │  │  Tool Registry      │  │
            │  │  - Command Tool     │  │
            │  │  - URL Fetch Tool   │  │
            │  │  - tfcmd Tool       │  │
            │  └──────────┬──────────┘  │
            │             │              │
            │  ┌──────────▼──────────┐  │
            │  │  Workflow Processor │  │
            │  └─────────────────────┘  │
            └─────────────┬─────────────┘
                          │
            ┌─────────────▼─────────────┐
            │   Grid CLI (tfcmd)        │
            │  - VM Management          │
            │  - Kubernetes Deployment  │
            │  - Gateway Configuration  │
            │  - Contract Operations    │
            └───────────────────────────┘
                          │
            ┌─────────────▼─────────────┐
            │   ThreeFold Grid          │
            │   (Blockchain + Nodes)    │
            └───────────────────────────┘
```

---

## Project Structure

```
grid-agent/
├── agent/                  # AI agent framework (library)
│   ├── pkg/
│   │   ├── core/          # Agent orchestration
│   │   ├── llm/           # LLM provider implementations
│   │   ├── tools/         # Tool system and built-ins
│   │   └── workflow/      # Workflow processing
│   └── go.mod
│
├── grid-cli/              # ThreeFold Grid CLI
│   ├── cmd/               # CLI commands
│   ├── internal/          # Internal packages
│   ├── docs/              # Command documentation
│   └── go.mod
│
├── grid-agent-gui/        # Desktop GUI application
│   ├── frontend/          # Svelte frontend
│   │   └── src/
│   ├── internal/          # Go backend logic
│   │   ├── config/       # Configuration management
│   │   └── tfcmd/        # CLI integration
│   ├── build/            # Build assets
│   ├── app.go            # Main app logic
│   └── go.mod
│
├── Makefile              # Build automation
├── go.work               # Go workspace
└── README.md             # This file
```

---

## Makefile Commands

Run `make help` to see all available commands:

```bash
# Building
make build                    # Build all components
make build-grid-cli          # Build tfcmd only
make build-grid-agent-gui    # Build GUI only

# Installation
make install                 # Install all components
make install-grid-cli        # Install tfcmd
make install-grid-agent-gui  # Install GUI with desktop entry

# Development
make install-deps            # Install system dependencies (Linux)
make dev-gui                 # Run GUI in dev mode with hot reload
make tidy                    # Tidy all go.mod files
make work-sync              # Sync Go workspace

# Testing
make test                    # Run all tests
make coverage               # Generate coverage reports

# Linting
make lint                    # Lint all components

# Cleanup
make clean                   # Remove build artifacts
```

---

## Configuration

### Grid Agent GUI

Settings are stored in `~/.config/grid-agent/settings.json`:

```json
{
  "mnemonics": "your twelve word mnemonic phrase here",
  "network": "mainnet",
  "geminiApiKey": "your-gemini-api-key",
  "theme": "dark",
  "isConfigured": true
}
```

### Grid CLI

Configuration is stored in `.tfgridconfig` in your system's config directory.

---

## Contributing

We welcome contributions! Please follow these guidelines:

### Getting Started

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**
4. **Run tests**: `make test`
5. **Run linter**: `make lint`
6. **Commit changes**: `git commit -m 'Add amazing feature'`
7. **Push to branch**: `git push origin feature/amazing-feature`
8. **Open a Pull Request**

### Code Standards

- Follow Go best practices and conventions
- Add tests for new features
- Update documentation for API changes
- Ensure all tests pass
- Run `make lint` before committing
- Use meaningful commit messages

### Development Workflow

```bash
# 1. Sync dependencies
make tidy

# 2. Make your changes

# 3. Test
make test

# 4. Lint
make lint

# 5. Build
make build
```

---

## Troubleshooting

### GUI: "tfcmd not found"

Ensure tfcmd is installed and in your PATH:

```bash
make install-grid-cli
which tfcmd
```

### GUI: Icon not showing (Linux)

Update icon cache and desktop database:

```bash
gtk-update-icon-cache ~/.local/share/icons/hicolor/ -f
update-desktop-database ~/.local/share/applications/
```

Then log out and log back in.

### Build: "wails not found"

Install Wails CLI:

```bash
make install-tools
# or
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Module: "package not found"

Sync the workspace:

```bash
make work-sync
make tidy
```

---

## License

Apache License 2.0

---

## Support

For issues and questions:

- **GitHub Issues:** [github.com/threefoldtech/grid-agent/issues](https://github.com/threefoldtech/grid-agent/issues)
- **ThreeFold Forum:** [forum.threefold.io](https://forum.threefold.io/)
- **Documentation:** [manual.grid.tf](https://manual.grid.tf)

---

## Credits

- **Author:** Sameh Abouel-saad
- **Framework:** [Wails](https://wails.io/)
- **AI Provider:** [Google Gemini](https://ai.google.dev/)
- **ThreeFold:** [threefold.io](https://threefold.io/)

---

## Related Projects

- [tfgrid-sdk-go](https://github.com/threefoldtech/tfgrid-sdk-go) - ThreeFold Grid SDK for Go (original monorepo)
- [ThreeFold Grid Proxy](https://github.com/threefoldtech/tfgrid-sdk-go/tree/development/grid-proxy) - Grid indexer and query service
- [ThreeFold Manual](https://manual.grid.tf) - Comprehensive Grid documentation


<parameter name="Complexity">7
