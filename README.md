# gh-grass

[![Go Reference](https://pkg.go.dev/badge/github.com/umekikazuya/gh-grass.svg)](https://pkg.go.dev/github.com/umekikazuya/gh-grass)
[![CI](https://github.com/umekikazuya/gh-grass/actions/workflows/ci.yml/badge.svg)](https://github.com/umekikazuya/gh-grass/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/umekikazuya/gh-grass)](https://goreportcard.com/report/github.com/umekikazuya/gh-grass)
[![Release](https://img.shields.io/github/v/release/umekikazuya/gh-grass?display_name=tag)](https://github.com/umekikazuya/gh-grass/releases)
[![License](https://img.shields.io/github/license/umekikazuya/gh-grass)](LICENSE)

`gh-grass` is a GitHub CLI extension for quickly checking GitHub contribution activity in your terminal.

It is for people who use `gh` daily and want to check their own graph, other users, or organization members without opening the browser.

![gh-grass demo](https://raw.githubusercontent.com/umekikazuya/gh-grass/refs/heads/assets/demo.gif)

## At a glance

- Command: `gh grass`
- Interface: interactive TUI (keyboard-driven)
- Data source: your authenticated GitHub session (`gh auth login`)

## Quick start (recommended)

```bash
gh extension install umekikazuya/gh-grass
gh grass
```

If the same command is already installed:

```bash
gh extension install umekikazuya/gh-grass --force
```

## What you can do

- View your own contribution graph
- Search and view another user's contributions
- View organization members' contributions

## Requirements

- [GitHub CLI](https://cli.github.com/) (`gh`)
- Authenticated GitHub session (`gh auth login`)

## Alternative installation

### Download binary from Releases

Download your platform binary from [Releases](https://github.com/umekikazuya/gh-grass/releases), make it executable, and place it in your `PATH`.

Then run:

```bash
gh-grass
```

### Build from source

```bash
go install github.com/umekikazuya/gh-grass/cmd/gh-grass@latest
```

## License

See [LICENSE](LICENSE) for details.
