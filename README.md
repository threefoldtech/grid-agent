# Grid Agent

An AI-powered agent framework and command-line tool suite for managing decentralized infrastructure deployments through natural language.

## What this is

Grid Agent provides a complete set of tools for intelligent automation and management of grid infrastructure. It includes a reusable AI agent core with LLM provider abstraction, a modern desktop application, and a command-line interface for infrastructure operations.

The agent interprets natural language requests, translates them into structured operations, and executes them against a grid backend. This makes it possible to deploy virtual machines, Kubernetes clusters, databases, and gateways using conversational commands rather than manual API calls.

## What this repository contains

- **Agent Framework** (`/agent`) — Reusable AI agent core with LLM provider abstraction and extensible tool system
- **Grid Agent GUI** (`/grid-agent-gui`) — Cross-platform desktop application built with Wails and Svelte
- **Grid CLI** (`/grid-cli`, `tfcmd`) — Command-line interface for grid operations and schema export

## Role in the stack

Grid Agent sits at the user-facing layer of the infrastructure stack. It communicates with the grid's management plane to execute workloads, report status, and handle lifecycle operations. The CLI provides the foundational operations that the agent framework orchestrates, while the GUI offers a chat-based interface for non-technical users.

## Relation to ThreeFold

This technology is used within the ThreeFold ecosystem and was first deployed on the ThreeFold Grid. The component itself is designed as reusable infrastructure technology and should be understood by its technical function first, independent of any specific deployment.

## Ownership

This repository is owned and maintained by TF-Tech NV, a Belgian company responsible for the development and maintenance of this technology.

## Components

### 1. Agent (`/agent`)

A standalone, reusable AI agent framework for building intelligent conversational assistants.

**Key features:**
- LLM provider abstraction (Google Gemini, extensible to OpenAI and others)
- Extensible tool system for CLI integration
- Real-time streaming command execution
- Automatic retry logic and error handling
- Structured JSON response parsing

**Package:** `github.com/threefoldtech/grid-agent/agent`

[Read the full Agent documentation →](./agent/README.md)

---

### 2. Grid Agent GUI (`/grid-agent-gui`)

A modern, cross-platform desktop application providing a chat interface for grid management.

**Key features:**
- AI-powered chat interface with Google Gemini
- Dark and light theme support
- Secure onboarding flow
- Real-time command execution with streaming output
- Multi-network support (mainnet, testnet, devnet)

**Technologies:** Wails v2, Svelte, TypeScript, Go

[Read the full GUI documentation →](./grid-agent-gui/README.md)

---

### 3. Grid CLI (`/grid-cli`)

Command-line interface for grid operations, providing the foundation for agent schema generation.

**Key features:**
- Deploy and manage VMs, Kubernetes clusters, and databases
- Gateway management (FQDN, Name)
- Contract operations
- Schema export for AI agent integration

**Package:** `github.com/threefoldtech/grid-agent/grid-cli`

[Read the full CLI documentation →](./grid-cli/README.md)

---

## Quick Start

You have two options to start using Grid Agent:

1. **Download prebuilt binaries** (simplest)
2. **Build and install from source** (for development)

### 1. Download prebuilt binaries

Prebuilt binaries for the latest tagged version are available on GitHub Releases:

- <https://github.com/threefoldtech/grid-agent/releases/latest>

#### Linux

1. Download the latest `tfcmd-linux-amd64` and `grid-agent-gui-linux-amd64`.
2. Make the binaries executable and move them into your local bin directory:

    ```bash
    chmod +x tfcmd-linux-amd64
    mv tfcmd-linux-amd64 ~/.local/bin/tfcmd

    chmod +x grid-agent-gui-linux-amd64
    mv grid-agent-gui-linux-amd64 ~/.local/bin/grid-agent-gui
    ```

3. Ensure `~/.local/bin` is in your `PATH`.

#### macOS

1. Download the appropriate CLI and GUI artifacts for your architecture (`amd64` or `arm64`).
2. Install the CLI:

    ```bash
    chmod +x tfcmd-darwin-*
    sudo mv tfcmd-darwin-amd64 /usr/local/bin/tfcmd   # Intel (amd64)
    # or
    sudo mv tfcmd-darwin-arm64 /usr/local/bin/tfcmd   # Apple Silicon (arm64)
    ```

3. For the GUI, download the `.app.zip` for your architecture and extract it. This produces `grid-agent-gui.app`.

    ```bash
    mv grid-agent-gui.app /Applications/Grid\ Agent.app
    ```

4. Launch from Spotlight or Finder. If macOS blocks it as an unidentified developer, allow it under **System Settings → Privacy & Security**.

#### Windows

1. Download `tfcmd-windows-amd64.exe` and `grid-agent-gui-windows-amd64.exe`.
2. Optionally rename and add the CLI to your `PATH`:

    ```powershell
    Rename-Item .\tfcmd-windows-amd64.exe tfcmd.exe
    # Then add its directory to the PATH environment variable
    ```

3. Run `grid-agent-gui-windows-amd64.exe` to start the GUI.

### 2. Build and install from source

#### Prerequisites

- **Go** 1.23 or higher
- **Node.js** 18 or higher (for GUI)
- **Wails CLI** v2.11.0+ (for GUI development)
- **System Dependencies** (Linux only): Run `make install-deps`

#### Installation from source

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

This installs binaries to:

- **Linux:** `~/.local/bin`
- **macOS:** `/usr/local/bin`

> **Note:** Make sure the installation directory is in your `PATH`.

### Using the GUI

After installation, launch the Grid Agent GUI:

```bash
# From terminal
grid-agent-gui

# Or use your application launcher
```

On first launch, you will be guided through:

1. Entering your mnemonic phrase
2. Selecting a network (mainnet/testnet/devnet)
3. Providing a Gemini API key

Get a free Gemini API key at: <https://aistudio.google.com/app/apikey>

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
- Dev server at <http://localhost:34115>

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
            │   Grid Backend            │
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
├── grid-cli/              # Grid CLI
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

We welcome contributions. Please follow these guidelines:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit changes: `git commit -m 'Add amazing feature'`
7. Push to branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

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

This project is licensed under the Apache License 2.0 — see the [LICENSE](LICENSE) file for details.
---

## Related Projects

- [zos_sdk_go](https://github.com/threefoldtech/tfgrid-sdk-go) — Grid SDK for Go
- [Grid Proxy](https://github.com/threefoldtech/tfgrid-sdk-go/tree/development/grid-proxy) — Grid indexer and query service
