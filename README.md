# gh-grass

A CLI tool to check GitHub contribution counts from the terminal with an interactive TUI.

## Features

- View GitHub contribution graphs in your terminal
- Check your own contributions
- Search and view other users' contributions
- View organization members' contributions
- Beautiful TUI interface powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)

## Installation

### From source

```bash
git clone https://github.com/umekikazuya/gh-grass.git
cd gh-grass
go install ./cmd/gh-grass
```

Or using `go install`:

```bash
go install github.com/umekikazuya/gh-grass/cmd/gh-grass@latest
```

## Usage

Simply run:

```bash
gh-grass
```

The interactive TUI will guide you through:

1. Viewing your own contributions
2. Searching for other users
3. Viewing organization members

## Requirements

- Go 1.25.5 or later (for building from source)
- [GitHub CLI](https://cli.github.com/) (gh command)
- GitHub authentication via `gh auth login`

## Development

### Prerequisites

- Go 1.25.5 or later
- GitHub personal access token

### Build

```bash
go build -o gh-grass ./cmd/gh-grass
```

### Run locally

```bash
go run ./cmd/gh-grass
```

## Architecture

The project follows clean architecture principles:

```
internal/
├── domain/         # Domain models and interfaces
├── infrastructure/ # GitHub API client and authentication
├── ui/            # TUI and command-line interface
└── usecase/       # Business logic
```

## License

See [LICENSE](LICENSE) file for details.
