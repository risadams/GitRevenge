# Git config

The **Git Config** tab edits git's own settings files: your global config, a repository's local config, and any file they include. It shows what each setting does, where an inherited value comes from, and every alias. Every save backs up the file first.

## Opening it

- From the command palette: **Edit global git config**, or **Edit this repository's git config** when a repository is open.
- Right-click a repository's tab and choose **Edit git config…**.
- In **Settings**, under Git, choose **Edit git config…**.
- In the commit composer, click **change** next to the name you commit as, or click the **Signed** mark. The tab opens at the file that sets that value, with the row highlighted.

The tab sits after your repository tabs. A dot on it means a file has unsaved changes. Closing it with unsaved changes asks first.

## Files

The side panel lists the files you can edit:

| File | What it holds |
|---|---|
| **Global** | Settings for every repository on this computer, usually `~/.gitconfig`. If git also reads `~/.config/git/config`, a note says so; changes go to the file shown. |
| **Local** | One repository's own settings, in `.git/config`. Each open repository has one. |
| **Included files** | Files that another one reads with `[include]` or `[includeIf]`. Click **edit that file** on a value that comes from one to add it to the list. |

System config (from the git installation) applies too, but it's read-only: changing it usually needs administrator rights.

## Settings

![The Settings view of a repository's config](../img/gitconfig-settings-light.webp#gh-light-mode-only)
![The Settings view of a repository's config](../img/gitconfig-settings-dark.webp#gh-dark-mode-only)

Settings are grouped by section, such as `[core]` or `[remote "origin"]`. Common settings show a line on what they do and a control that fits:

- **On/off settings** are switches. If the file says `yes` or `on`, it keeps saying so until you change it.
- **Settings with a few values**, such as `core.autocrlf`, are a row of choices. Longer lists are a menu.
- **Paths** have a button to choose a file. Anything else is a text field: press <kbd>Enter</kbd> or click away to apply it, <kbd>Escape</kbd> to put it back.
- **Settings with several values**, such as `remote.origin.fetch`, list each one, with **+ add a value**.

Changed settings get a rose dot and say what they were (**changed from false**), or **new**.

### Inherited values

With **everything git reads** selected, the list also shows values that come from other files, greyed, with a tag saying where from: **global**, **system**, **local** or **included**.

- **Override here** copies an inherited value into this file, where you can change it.
- A value this file sets over another says so: **overrides global: yes**. **Use global** removes it from this file, so the inherited value applies again.

Choose **only this file** to hide inherited values. The filter box searches keys and values.

### Adding a setting

**+ Add a setting** opens a form. The key field suggests every setting your git knows, with a description for common ones, and offers the allowed values where there's a fixed list.

### Secrets

Values that hold credentials (`http.extraHeader`, passwords and tokens) and the password part of URLs show as hidden. **Show** reveals one row until you switch files or leave the tab.

## Saving

Edits stay in the app until you **Save** (<kbd>Ctrl</kbd>+<kbd>S</kbd>). **Discard** throws them away.

- Git checks the file before it's saved. If it can't read it, nothing is written, and the header says which line to fix.
- The file as it was is backed up first. See [History](#history).
- If something else changes the file while you have unsaved edits, a banner offers **Reload theirs** or **Keep mine**. Saving after **Keep mine** shows the file on disk and asks before replacing it.

A file with no unsaved edits reloads on its own when it changes, and every open repository picks up changes to global config straight away.

## Aliases

![The Aliases view of the global config](../img/gitconfig-aliases-light.webp#gh-light-mode-only)
![The Aliases view of the global config](../img/gitconfig-aliases-dark.webp#gh-dark-mode-only)

The **Aliases** view lists the file's aliases, then the ones it inherits. **New alias** and **Edit** open a form with a name and what the alias runs: a git command (`log --oneline --graph`) or a shell command, shown on a dark ground.

Warnings point out aliases that won't do what you expect:

- one named after a built-in command, such as `log`, which git ignores;
- one that calls itself;
- a local alias that replaces a global one.

**Copy to** puts an alias in another file (save it there to keep it), and **Copy here** brings an inherited one into this file. Git Revenge shows aliases but never runs them.

## Raw text

![A config file in the raw text editor, with an error on line 1](../img/gitconfig-raw-light.webp#gh-light-mode-only)
![A config file in the raw text editor, with an error on line 1](../img/gitconfig-raw-dark.webp#gh-dark-mode-only)

**Raw text** is the file as text, for anything the other views don't cover: comments, ordering, or a lot of changes at once. Edits show up in Settings and Aliases too. If git can't read a line, it's marked, with a note beside it, and the file can't be saved until it's fixed.

Nothing is hidden in raw text, including tokens and passwords.

## History

![A backup compared with the file as it is now](../img/gitconfig-history-light.webp#gh-light-mode-only)
![A backup compared with the file as it is now](../img/gitconfig-history-dark.webp#gh-dark-mode-only)

Every save keeps the file as it was before. **History** lists those backups by day, and compares the one you pick with the file as it is now, unified or side by side.

**Restore** puts a backup back. The file as it is now is backed up first, so a restore can be undone the same way.

The last 20 backups of each file are kept, in a `config-backups` folder next to [your settings](install.md#where-your-data-lives).
