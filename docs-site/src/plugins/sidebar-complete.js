// A site plugin that fails the build when a page of the docs is missing from the sidebar.
// sidebars.js lists every page by hand, so the order a reader sees is the one written there;
// a page added to docs/ and not to sidebars.js would otherwise be built but reachable from
// nowhere.
//
// It asks Docusaurus which sidebar each page landed in, rather than working out page ids
// itself, since Docusaurus drops number prefixes from them (adr/0001-wails-desktop-shell.md
// is adr/wails-desktop-shell).
module.exports = function sidebarComplete() {
  return {
    name: 'gitrevenge-sidebar-complete',
    allContentLoaded({ allContent }) {
      const docs = allContent['docusaurus-plugin-content-docs'].default;
      const current = docs.loadedVersions.find((v) => v.versionName === 'current');
      const missing = current.docs.filter((doc) => !doc.sidebar).map((doc) => doc.source.replace('@site/../', ''));
      if (missing.length > 0) {
        throw new Error(
          `gitrevenge: these pages are in no sidebar; add each to docs-site/sidebars.js where a reader would look for it:\n  ${missing.join('\n  ')}`,
        );
      }
    },
  };
};
