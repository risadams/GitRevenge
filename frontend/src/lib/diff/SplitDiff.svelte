<script lang="ts">
  import { MergeView } from '@codemirror/merge'
  import { EditorState } from '@codemirror/state'
  import { EditorView, lineNumbers } from '@codemirror/view'
  import { syntaxHighlighting } from '@codemirror/language'
  import { classHighlighter } from '@lezer/highlight'
  import { languageFor } from './highlight'

  let { path, oldText, newText }: { path: string; oldText: string; newText: string } = $props()

  let host: HTMLDivElement

  // Rebuild the read-only merge view whenever the inputs change.
  $effect(() => {
    const [p, a, b] = [path, oldText, newText]
    let view: MergeView | undefined
    let cancelled = false
    void languageFor(p).then((lang) => {
      if (cancelled) return
      const ext = [
        EditorView.editable.of(false),
        EditorState.readOnly.of(true),
        lineNumbers(),
        syntaxHighlighting(classHighlighter),
        EditorView.theme({}, { dark: true }),
        ...(lang ? [lang] : []),
      ]
      view = new MergeView({
        a: { doc: a, extensions: ext },
        b: { doc: b, extensions: ext },
        parent: host,
        collapseUnchanged: { margin: 3, minSize: 6 },
        gutter: true,
      })
    })
    return () => {
      cancelled = true
      view?.destroy()
    }
  })
</script>

<div class="split" bind:this={host}></div>

<style>
  .split {
    font-size: 12px;
    tab-size: 4;
  }

  .split :global(.cm-mergeView) {
    height: 100%;
  }

  .split :global(.cm-editor) {
    background: var(--bg);
    color: var(--fg);
  }

  .split :global(.cm-gutters) {
    background: var(--bg-raised);
    color: var(--fg-muted);
    border-right: 1px solid var(--border);
  }

  .split :global(.cm-content) {
    font-family: var(--font-mono);
  }
</style>
