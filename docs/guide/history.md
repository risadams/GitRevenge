# Diffs and history

## Diffs

Click a file in a commit, or in your uncommitted changes, to see its diff. Pick a layout:

- **Unified:** one column, with removed and added lines interleaved. This is the layout for [staging hunks and lines](changes.md#staging).
- **Split:** the old and new versions side by side.
- **File:** the whole file as of that version.

Code is syntax highlighted. Images (PNG, JPEG, GIF, WebP, AVIF, BMP, ICO and SVG) are shown as pictures and compared side by side, fitted to the view or at actual size. <kbd>Alt</kbd>+<kbd>Z</kbd> turns word wrap on and off. Press <kbd>Escape</kbd> to go back to the graph.

## Rendered previews

Some files read better rendered than as text. In the **File** layout, these get a **Rendered** option, and the app remembers whether you left it on:

| Files | Shown as |
|---|---|
| Markdown (`.md`, `.markdown`) | formatted text, including LaTeX math and Mermaid diagrams |
| CSV and TSV (`.csv`, `.tsv`, `.tab`) | a table you can sort by any column |
| JSON, YAML and TOML | a tree you can expand and search |
| LaTeX (`.tex`, `.latex`) | typeset math |
| Mermaid (`.mmd`, `.mermaid`) | the diagram |
| PDF | the document |

## File history and blame

From a file's diff, **History** lists every commit that changed it, and **Blame** shows who last changed each line and in which commit.

## The reflog

The reflog is git's record of where HEAD and each branch have been. It's how you find a commit you thought you'd lost to a reset or a rebase. Choose **Show the reflog…** in the command palette for HEAD's, or right-click a branch for its own.

## Notes

Git notes attach extra text to a commit without changing it. The details panel shows a selected commit's notes, and lets you add, edit or remove one.

## Signatures

A signed commit shows a badge in its details:

| Badge | Meaning |
|---|---|
| Verified | The signature is good, from a key you trust. |
| Can't verify | The signature may be fine, but your machine has no trusted key to check it against. |
| Bad signature | The commit doesn't match its signature. |
| Expired | The signature or its key has expired. |
| Revoked | The key has been revoked. |

### Why a signature says "Can't verify"

Git checks signatures with your own keys and settings, so this usually means something is missing locally rather than wrong with the commit.

- **SSH-signed commits** need an allowed-signers file listing whose key is whose. Point git at one with `gpg.ssh.allowedSignersFile`, and add a line for each signer:

  ```sh
  git config --global gpg.ssh.allowedSignersFile ~/.ssh/allowed_signers
  echo "you@example.com $(cat ~/.ssh/id_ed25519.pub)" >> ~/.ssh/allowed_signers
  ```

- **Commits merged on GitHub** are signed by GitHub's own GPG key. Import it, then tell GPG you trust it:

  ```sh
  curl -s https://github.com/web-flow.gpg | gpg --import
  gpg --edit-key noreply@github.com trust
  ```

The commit's details show the result as soon as git can verify it.
