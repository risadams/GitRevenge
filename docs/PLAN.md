# Git Revenge — Plan

Git Revenge is a fast, local-only desktop git client inspired by GitKraken. It has a Go backend and is built to handle multiple repos, including very large ones.

This document records the decisions from the initial requirements session (2026-10-28). Architectural decisions with lasting consequences are recorded in [`docs/adr/`](adr/). Progress is tracked in GitHub issues, grouped by milestone.

## Goals

- An easy to use, powerful, fast git client.
- Works well on very large repos (see [Performance targets](#performance-targets)).
- Works across many repos, with user-defined project groups.
- Fully local: no accounts, no cloud, no telemetry.

## Non-goals

- AI features.
- GitKraken Launchpad, Workspaces, Cloud Patches, accounts or any cloud sync.
- Hosting-provider or issue-tracker integrations (GitHub/GitLab/Bitbucket/Jira APIs, pull requests).

## Platforms

- **Windows and macOS are first-class.** Linux is a nice-to-have.
- Requires a system `git` **>= 2.40**. `git-lfs` is required for LFS features.

## Architecture

| Concern | Decision | ADR |
|---|---|---|
| App shell | Wails v2 desktop app (WebView2 / WKWebView) | [0001](adr/0001-wails-desktop-shell.md) |
| Git engine | System `git` CLI behind a thin Go layer | [0002](adr/0002-system-git-cli.md) |
| Frontend | Svelte 5 + TypeScript + Vite; canvas commit graph; CodeMirror 6 for diffs | [0003](adr/0003-svelte-canvas-graph.md) |
| Credentials | Existing helpers first; askpass bridge into the UI | [0004](adr/0004-askpass-bridge.md) |
| State | Local JSON settings in the OS config dir | [0005](adr/0005-local-state-and-external-changes.md) |

### Backend

- **Git calls.** Every call runs with `GIT_TERMINAL_PROMPT=0`. Calls are context-cancellable and stream their output.
- **Persistent helper.** Each open repo keeps a long-lived `git cat-file --batch` process for fast object reads.
- **Per-repo operation queue.** Mutating operations are serialized, and reads run concurrently. If `index.lock` is held by another process, the app retries briefly and then reports it clearly.
- **Commit graph.** The lane layout is computed incrementally in Go while commits stream from `git log`. Only compact row data is sent to the UI.
- **Watching for external changes.** `fsnotify` watches `HEAD`, refs, `packed-refs`, the index and the work tree, debounced to about 150ms. The app uses `core.fsmonitor` where it is enabled.
- **In-progress operations.** A merge, rebase, cherry-pick, revert or bisect that is in progress is detected even if it started outside the app.

### Frontend

- **Commit graph.** Drawn on a `<canvas>` with virtualized rows, so only the visible rows are drawn.
- **Diff, conflict and blame views.** Built with CodeMirror 6.
- **Visual design.** Until the 0.6 polish milestone, the UI is plain and functional.

## Multi-repo

- **Tabs.** Each open repo gets a tab and keeps warm backend state. Open tabs are restored on launch.
- **Project groups.** A group is a named, local collection of repo paths, and a repo can be in several groups. A group offers a dashboard (branch, ahead/behind, dirty state, last fetch), batch fetch, batch pull (fast-forward only), and an action to open every repo in the group as tabs.
- **Not in v1:** cross-repo branch operations and cross-repo commits.

## Feature scope

**v1 (milestones 0.1 to 0.6):**
- Commit graph with search and filter, and a commit detail panel.
- Staging at file, hunk and line level, a commit composer, amend, and signing (respecting git config).
- Diffs (split and unified, plus image diffs).
- Branches, tags and remotes, including fetch, pull, push and force-with-lease.
- Merge, cherry-pick, and drag-and-drop in the graph.
- A 3-pane conflict tool.
- Stash.
- Interactive rebase, reset, revert, and reflog-based undo/redo.
- LFS and submodules.
- File history and blame.
- Clone and init.
- Command palette, keyboard shortcuts, and light/dark themes.

**Later (1.1+):**
- Worktrees.
- Bisect UI.
- Hooks UI.
- Gitflow.
- An integrated terminal.
- A setup UI for partial clone and sparse checkout. The app still works on repos that already use them.
- Signature verification badges.

## Performance targets

The reference repo is the Linux kernel (about 1.3M commits and 90k files).

| Metric | Target |
|---|---|
| First screen of the graph | < 1s |
| Graph scrolling | 60fps across the full history |
| `status` refresh (warm cache) | < 500ms |
| Switching repo tabs | Instant |
| Memory with the kernel repo open | < 500MB |

The benchmark suite is opt-in: set `GITREVENGE_BENCH_REPO=/path/to/linux` and run `go test -bench`. It does not run in CI. Run it before tagging each milestone.

## Milestones

1. **0.1.0 Foundation:** scaffold, git CLI layer, repo open and tabs, read-only graph, commit detail, CI and release.
2. **0.2.0 Working tree:** status, staging, commit composer, diff viewer, stash.
3. **0.3.0 Branching & remotes:** branches, tags, remotes, credential prompts, merge, cherry-pick, drag-and-drop.
4. **0.4.0 History rewriting:** interactive rebase, reset, revert, undo/redo, conflict tool.
5. **0.5.0 Multi-repo & extras:** project groups and dashboard, LFS, submodules, file history and blame, clone and init.
6. **0.6.0 Polish:** visual design, themes, command palette, shortcuts, perf pass, update check, signed releases.

Each feature area has an epic issue. Its tasks are tracked as sub-issues, each sized for one PR.

## Workflow

- **Issues to PRs.** Each issue gets its own branch, named `<issue#>-short-slug`, and one PR that says `Closes #N` and is squash-merged into `main`.
- **Commits** follow [Conventional Commits](https://www.conventionalcommits.org/). Release notes are generated from them.
- **Merging.** Chore PRs may be merged once CI is green. Feature PRs are reviewed by the maintainer.
- **Versioning.** Semver, starting at `0.1.0`.
- **Testing.**
  - Go tests run against fixture repos that are built on the fly with real git.
  - Vitest covers frontend logic.
  - CI runs on Windows and macOS.
- **Linting.** `golangci-lint`, `svelte-check` and `eslint`.

## Distribution

- **Builds.** A GitHub Actions workflow is triggered by `v*` tags.
  - Windows: an NSIS installer, unsigned for now.
  - macOS: a universal `.dmg`, signed and notarized with the maintainer's Apple Developer account.
- **Update check.** The app checks GitHub Releases for a newer semver version. There is no auto-update.
- **Audience.** First the maintainer, then an internal team, then open source under Apache-2.0.
