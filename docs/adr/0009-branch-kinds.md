# 0009 — Branch kinds are built in, matched by local patterns

- Status: Accepted
- Date: 2026-10-31

## Context

Some branches aren't like the others. `main` is the line everything returns to; `release/1.4` is a release in progress, closer to a tag than to a feature branch; `gh-pages` holds a built site and shares no history with the code; `dependabot/*` branches are opened and deleted by a bot; `production` may be what a deploy pipeline watches. Today the graph and sidebar treat them all alike, and the cleanup, force-push and Insights rules can't tell them apart. The design came from a grilling session on 2026-10-31 (epic #276).

## Decision

- **Five built-in kinds:** Trunk, Release, Environment, Docs and Bot. Kinds carry behaviour, not just decoration, so they are a fixed set rather than labels people define. Custom labels can come later if a real need appears.
- **Matched by glob against the branch name**, without the remote prefix, so `origin/release/1.4` is a Release branch. The first matching kind wins, in the order above.
- **Built-in defaults** work with no setup: Trunk `main`, `master`, `develop`, `trunk`; Release `release/*`, `releases/*`, `hotfix/*`; Docs `gh-pages`, `gh-docs`, `docs-site`; Bot `dependabot/*`, `renovate/*`; Environment none, since `staging` and `production` are too often ordinary branches.
- **Patterns are local.** Overrides live in Git Revenge's settings, keyed by repository path, beside the per-repo hide patterns. Nothing is written into the repository, so the app stays local-only and one person's kinds don't change anyone else's view.
- **A commit's kind is its lane's kind.** A lane takes the kind of the branch tip that starts it and keeps it along that first-parent line. Commits in a merged-in side line with no branch of their own have no kind.
- **What a kind changes:**
  - Look: a chip colour and icon in the graph, a sidebar section, and in the timeline the kind's icon in place of author initials (all but Trunk). Every mark has the kind's name in its tooltip and accessible name.
  - Layout: Trunk lanes are pinned, the first in column 0, the next in column 1.
  - Visibility: Docs lanes draw muted; Docs and Bot each have a visibility toggle, on by default.
  - Safety: cleanup on pull never deletes Trunk, Release, Environment or Docs branches; force push to Trunk, Release or Environment and deleting Trunk or Environment ask first, naming the kind.
  - Insights: Docs commits are left out, Bot commits are grouped as dependency updates, and Release branches can bound the releases window.
- **Kinds are an Insights input.** ADR 0008 makes Insights a pure function of repository state and explicit inputs. The kind patterns in use affect results, so they join the inputs and are printed in every export.

## Consequences

- Most repositories get sensible kinds without configuration, and the Docs default fixes the common `gh-pages` case.
- Two people with different patterns can see different Insights numbers for the same repository; the exported inputs show why.
- Kinds don't travel with a clone. A team that wants shared kinds would need a committed file, which this decision defers.
- Pinned Trunk lanes change the incremental layout and have to stay within the kernel benchmark's targets.
