# Accessibility

Git Revenge aims to be usable with a keyboard alone, with a screen reader, and by people with low vision or colour-vision differences. It is early software and doesn't meet that goal everywhere yet. This page says what we aim for and what we know is missing.

## What we aim for

- **Keyboard.** Every action is reachable without a mouse, through shortcuts, the command palette and standard focus order. Focus is always visible, and dialogs trap and restore focus.
- **Screen readers.** Controls have accessible names and roles. Status changes, errors and progress are announced.
- **Colour and contrast.** Text and controls meet WCAG 2.2 AA contrast in both light and dark themes. Colour is never the only carrier of meaning. Diff additions and removals also use markers, and graph lanes also differ by position.
- **Motion and size.** Animation respects `prefers-reduced-motion`, and the UI follows system text scaling.

## Known gaps

- **Commit graph.** It is drawn on a canvas ([ADR 0003](docs/adr/0003-svelte-canvas-graph.md)), so it has no built-in accessibility tree. Keyboard navigation and a screen-reader-friendly list of the same rows are needed.
- **Diff, blame and conflict views.** These use CodeMirror 6, whose screen-reader support is good but has not been tested with this app.
- **Testing.** There has been no audit with NVDA, JAWS or VoiceOver yet.

## Reporting a problem

Open an issue with the `accessibility` label and tell us the assistive technology and operating system you use. Accessibility bugs are treated as bugs, not enhancements.

## For contributors

- Use native elements before ARIA, and add the keyboard behaviour with the mouse behaviour.
- Check new UI with the keyboard only, in both themes, and at 200% zoom.
- Add accessible names to icon-only buttons.
