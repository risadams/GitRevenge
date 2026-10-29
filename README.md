<img src="build/appicon.svg" alt="" width="96" align="right">

# Git Revenge

Git Revenge is a fast, local-only desktop git client for Windows and macOS. It has a Go backend and is inspired by GitKraken, minus the AI, Launchpad and cloud features.

> **Status:** early development. [`0.5.0`](https://github.com/risadams/GitRevenge/releases/tag/v0.5.0) is the latest release, and `0.6.0` (visual design and release polish) is in progress.

## Goals

- **Fast on very large repos.** The reference is the Linux kernel: the first screen of the graph appears in under 1s, and the graph scrolls at 60fps.
- **Powerful.** Commit graph, staging at file, hunk and line level, interactive rebase, a 3-pane conflict tool, stash, LFS, submodules, blame, and undo/redo.
- **Multi-repo.** Repos open in tabs, and you can create local project groups with a dashboard and batch fetch/pull.
- **Local only.** No accounts, cloud or telemetry. It uses your existing git config, credential helpers and SSH agent.
- **Comfortable to live in.** Light and dark themes, a timeline graph with a compact mode, a command palette (<kbd>Ctrl</kbd>+<kbd>K</kbd>) and keyboard shortcuts for common actions (<kbd>Ctrl</kbd>+<kbd>/</kbd> lists them; <kbd>⌘</kbd> on macOS).

## Install

Download the Windows installer or the macOS disk image from the [latest release](https://github.com/risadams/GitRevenge/releases/latest).

## Requirements

- `git` 2.40 or newer on your `PATH`.
- `git-lfs` if you want the LFS features.

## Development

You need Go 1.27.3+, Node 24+ and the [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation), pinned to the version in `go.mod`:

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
```

```sh
wails dev        # run with hot reload
wails build      # produce build/bin/gitrevenge(.exe)
```

The Go binary embeds `frontend/dist`. On a fresh clone, build the frontend once before running Go tools directly:

```sh
cd frontend && npm ci && npm run build && cd ..
go test ./...
```

Frontend checks, as CI runs them:

```sh
cd frontend && npm run check && npm run lint && npm test
```

The TypeScript bindings in `frontend/wailsjs` are generated from the Go API (`wails dev`, `wails build` or `wails generate module`). Commit them exactly as generated.

## Docs

- [Plan](docs/PLAN.md): scope, architecture, performance targets, milestones and workflow.
- [Architecture decision records](docs/adr/).
- Progress is tracked in [issues](https://github.com/risadams/GitRevenge/issues), grouped by [milestone](https://github.com/risadams/GitRevenge/milestones).

## License

[Apache-2.0](LICENSE)
