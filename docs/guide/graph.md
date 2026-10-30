# The commit graph

The graph is the main view of a repository: every commit, the branches and tags that point at it, and the lines connecting them.

It doesn't load the whole history when you open a repository. It loads the commits just ahead of where you're looking, and more as you scroll, search or jump. That's what keeps it quick on repositories with a million commits or more.

## Moving around

- Scroll, or use the arrow keys to move the selection one commit at a time.
- <kbd>Page Up</kbd> and <kbd>Page Down</kbd> move a page at a time.
- <kbd>Home</kbd> and <kbd>End</kbd> go to the newest and oldest commits.

Selecting a commit shows its details on the right: author, committer, message, [notes](history.md#notes), [signature](history.md#signatures) and the files it changed.

## Searching

Press <kbd>Ctrl</kbd>+<kbd>F</kbd> to search. You can search by commit message, author or SHA. You can also search for a trailer, such as `Co-authored-by: Sam` or `Closes: #12`.

<kbd>Enter</kbd> goes to the next match and <kbd>Shift</kbd>+<kbd>Enter</kbd> to the previous one. Matches further back than what's loaded are found too; the graph loads up to them.

## Branches and tags

The sidebar lists local branches, remotes, tags and stashes. Type in **Filter branches and tags** to narrow the list. Click a branch or tag to jump to its commit.

## Two layouts

The default **timeline** layout gives each commit two lines: the subject, and the author and date. **Compact rows** fits one commit per line, so you see more history at once. Switch between them in [Settings](settings.md).

## Right-click menus

Right-click a commit, branch or tag for what you can do with it: check it out, create a branch or tag there, merge, rebase, cherry-pick, revert, reset, or see its reflog. [Branches, merges and rebases](branches.md) covers these.
