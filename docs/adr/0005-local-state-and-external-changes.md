# 0005 — Local state and coexisting with external git usage

- Status: Accepted
- Date: 2026-10-28

## Context

The app is local-only, with no accounts and no cloud. Users will keep using the CLI and their IDEs on the same repos while the app is open.

## Decision

- **App state is stored locally.** Settings, project groups, recents and open tabs live in a JSON file under the OS config directory: `%APPDATA%\GitRevenge` on Windows and `~/Library/Application Support/GitRevenge` on macOS. Writes are atomic: write a temp file, then rename it over the original.
- **Repos are watched for external changes.** `fsnotify` watches `HEAD`, refs, `packed-refs`, the index and the work tree, with a debounce of about 150ms. When `core.fsmonitor` is enabled, the app relies on it instead of watching every file.
- **Mutating operations are serialized per repo**, while reads run concurrently. If another process holds `index.lock`, the app retries briefly, then reports that another git process is running.
- **In-progress operations are detected**, even if they started outside the app. This covers merge, rebase, cherry-pick, revert and bisect. The app shows a banner with Continue, Abort and Skip.

## Consequences

- The UI stays consistent with changes made in other tools without manual refreshes.
- Watching huge work trees is costly, so the app depends on fsmonitor for the largest repos.
