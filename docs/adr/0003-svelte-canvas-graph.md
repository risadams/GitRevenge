# 0003 — Svelte 5 frontend with a canvas commit graph

- Status: Accepted
- Date: 2026-10-28

## Context

The UI has to scroll a graph of 1M+ commits at 60fps and show large diffs. The maintainer had no framework preference.

## Decision

- **Svelte 5 + TypeScript + Vite** for the UI.
- **The commit graph is drawn on a `<canvas>`.** Go computes the lane layout incrementally (columns, colors and edges for each row) and sends compact rows. The UI draws only the visible rows.
- **CodeMirror 6** for diff, conflict (merge view) and blame views.
- **Plain CSS with design tokens.** No component library. The visual design is deferred to the 0.6 polish milestone.

## Consequences

- Svelte has a small runtime and fine-grained reactivity, which suits streaming data.
- The canvas avoids the scaling limits a DOM or SVG graph hits well before 1M rows. In exchange, we have to implement hit-testing, selection and accessibility ourselves.
- CodeMirror 6 handles large documents with virtualization and is lighter than Monaco.
