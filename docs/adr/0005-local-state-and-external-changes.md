# 0005 — Local state and coexisting with external git usage

- Status: Accepted
- Date: 2026-10-28

## Context

The app is local-only, with no accounts and no cloud. Users will keep using the CLI and their IDEs on the same repos while the app is open.

## Decision

- **App state is stored locally.** Settings, project groups, recents and open tabs live in a JSON file under the OS config directory: `%APPDATA%\GitRevenge` on Windows and `~/Library/Application Support/GitRevenge` on macOS. Writes are atomic: write a temp file, then rename it over the original.
- **Repos are watched for external changes.** Each repo gets one native recursive subscription (`ReadDirectoryChangesW` on Windows, FSEvents on macOS, via `github.com/syncthing/notify`) covering the work tree and the git directory. Events are classified (HEAD, refs, index, in-progress state, config, work tree) and debounced: about 150ms of quiet, at most 1s of delay. Work-tree paths are passed through one batched `git check-ignore`, so ignored files never trigger refreshes.
- **Mutating operations are serialized per repo**, while reads run concurrently. If another process holds `index.lock`, the app retries briefly, then reports that another git process is running.
- **In-progress operations are detected**, even if they started outside the app. This covers merge, rebase, cherry-pick, revert and bisect. The app shows a banner with Continue, Abort and Skip.

## Consequences

- The UI stays consistent with changes made in other tools without manual refreshes.
- Cost does not grow with the number of directories, unlike per-directory watchers such as `fsnotify`, which uses kqueue with one file descriptor per file on macOS.
- `core.fsmonitor`, where enabled, speeds up the `git status` that follows a change. The watcher itself does not depend on it.
- macOS FSEvents needs cgo, which Wails' macOS builds already require.
- *Amended 2026-10-28 (#15): originally specified `fsnotify` plus reliance on fsmonitor for large trees.*
