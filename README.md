<img src="build/appicon.svg" alt="" width="96" align="right">

# Git Revenge

Git Revenge is a fast, local-only desktop git client for Windows and macOS, built to stay quick on very large repositories. It has a Go backend that drives your installed `git`, and no accounts, cloud features or AI.

> **Status:** early development. [`0.5.0`](https://github.com/risadams/GitRevenge/releases/tag/v0.5.0) is the latest release, and `0.6.0` (visual design and release polish) is in progress.

## Features

- **A commit graph for big histories.** History loads just ahead of where you're looking rather than all at once. The aim on the Linux kernel's 1.3M commits is a first screen in under a second and scrolling at 60fps. Search, jump to any commit, and filter branches and tags.
- **Everyday git.** Stage files, hunks or single lines, write and amend commits (signed if your git config says so), stash, branch, tag, merge, rebase, cherry-pick, reset and revert, with undo and redo.
- **History tools.** Interactive rebase, a 3-pane conflict tool, file history and blame.
- **Diffs.** Unified and split diffs with syntax highlighting, image diffs and word wrap.
- **Remotes.** Fetch, pull and push with your existing credential helpers and SSH agent; the app asks for anything they can't answer.
- **Many repos.** Repos open in tabs, and local project groups get a dashboard with batch fetch and pull.
- **LFS and submodules.** Track LFS patterns, see LFS files, lock and unlock them, and initialize and update submodules.
- **Comfortable to live in.** Light and dark themes, a timeline graph with a compact mode, a command palette (<kbd>Ctrl</kbd>+<kbd>K</kbd>) and keyboard shortcuts for common actions (<kbd>Ctrl</kbd>+<kbd>/</kbd> lists them; <kbd>⌘</kbd> on macOS).

## Privacy

There are no accounts, no cloud sync and no telemetry. Git Revenge uses your own git config, credential helpers and SSH agent. The one request it makes on its own is a check of GitHub Releases for a newer version at startup, which you can turn off from the command palette.

## Install

Download the Windows installer or the macOS disk image from the [latest release](https://github.com/risadams/GitRevenge/releases/latest).

- **macOS:** from 0.6.0 the disk image is signed and notarized. Earlier releases aren't.
- **Windows:** the installer isn't signed yet, so SmartScreen warns before it runs.

## Requirements

- `git` 2.40 or newer.
- `git-lfs`, for the LFS features.
- On a very large repository, a commit-graph file makes the graph's first screen much faster: run `git commit-graph write --reachable`, or set `fetch.writeCommitGraph=true`.

## Development

You need Go 1.27.3 or newer, Node 24 or newer, and the [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation) at the version pinned in `go.mod`:

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

Performance checks run against a large repository you point them at; see [Performance targets](docs/PLAN.md#performance-targets).

## Docs

- [Plan](docs/PLAN.md): scope, architecture, performance targets, milestones and workflow.
- [Architecture decision records](docs/adr/).
- [Releasing](docs/RELEASING.md).
- [Accessibility](ACCESSIBILITY.md), [contributing](CONTRIBUTING.md), [code of conduct](CODE_OF_CONDUCT.md) and [security policy](SECURITY.md).
- Progress is tracked in [issues](https://github.com/risadams/GitRevenge/issues), grouped by [milestone](https://github.com/risadams/GitRevenge/milestones).

## License

[Apache-2.0](LICENSE)
