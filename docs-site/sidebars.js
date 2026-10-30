// @ts-check

// The docs in the order a newcomer reads them: install, then the guide from opening a
// repository through to settings. The project's own papers (plan, releasing, decisions) come
// last, collapsed.
//
// Every page is listed by hand, under a short label, so the order is the one written here
// and never a file name's. src/plugins/sidebar-complete.js fails the build if a page under
// docs/ is missing from this file.

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  mainSidebar: [
    {
      type: 'category',
      label: 'User guide',
      link: { type: 'doc', id: 'guide/README' },
      items: [
        { type: 'doc', id: 'guide/install', label: 'Installing' },
        { type: 'doc', id: 'guide/repositories', label: 'Repositories and tabs' },
        { type: 'doc', id: 'guide/graph', label: 'The commit graph' },
        { type: 'doc', id: 'guide/changes', label: 'Changes and commits' },
        { type: 'doc', id: 'guide/branches', label: 'Branches, merges and rebases' },
        { type: 'doc', id: 'guide/history', label: 'Diffs and history' },
        { type: 'doc', id: 'guide/remotes', label: 'Remotes, LFS and submodules' },
        { type: 'doc', id: 'guide/settings', label: 'Settings and shortcuts' },
      ],
    },
    {
      type: 'category',
      label: 'The project',
      collapsed: true,
      items: [
        { type: 'doc', id: 'PLAN', label: 'Plan' },
        { type: 'doc', id: 'RELEASING', label: 'Releasing' },
        {
          type: 'category',
          label: 'Design decisions',
          collapsed: true,
          items: [
            { type: 'doc', id: 'adr/wails-desktop-shell', label: 'ADR-0001 Wails desktop shell' },
            { type: 'doc', id: 'adr/system-git-cli', label: 'ADR-0002 The system git' },
            { type: 'doc', id: 'adr/svelte-canvas-graph', label: 'ADR-0003 A canvas graph' },
            { type: 'doc', id: 'adr/askpass-bridge', label: 'ADR-0004 The askpass bridge' },
            { type: 'doc', id: 'adr/local-state-and-external-changes', label: 'ADR-0005 Local state' },
            { type: 'doc', id: 'adr/docs-site', label: 'ADR-0006 The docs site' },
          ],
        },
      ],
    },
  ],
};

module.exports = sidebars;
