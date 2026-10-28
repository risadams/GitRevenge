# Git Revenge

Git Revenge is a fast, local-only desktop git client for Windows and macOS. It has a Go backend and is inspired by GitKraken, minus the AI, Launchpad and cloud features.

> **Status:** early development. The project is working toward `0.1.0`, and nothing is usable yet.

## Goals

- **Fast on very large repos.** The reference is the Linux kernel: the first screen of the graph appears in under 1s, and the graph scrolls at 60fps.
- **Powerful.** Commit graph, staging at file, hunk and line level, interactive rebase, a 3-pane conflict tool, stash, LFS, submodules, blame, and undo/redo.
- **Multi-repo.** Repos open in tabs, and you can create local project groups with a dashboard and batch fetch/pull.
- **Local only.** No accounts, cloud or telemetry. It uses your existing git config, credential helpers and SSH agent.

## Requirements

- `git` 2.40 or newer on your `PATH`.
- `git-lfs` if you want the LFS features.

## Docs

- [Plan](docs/PLAN.md): scope, architecture, performance targets, milestones and workflow.
- [Architecture decision records](docs/adr/).
- Progress is tracked in [issues](https://github.com/risadams/GitRevenge/issues), grouped by [milestone](https://github.com/risadams/GitRevenge/milestones).

## License

[Apache-2.0](LICENSE)
