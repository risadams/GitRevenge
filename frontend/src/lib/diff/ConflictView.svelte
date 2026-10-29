<script lang="ts">
  import { syntaxHighlighting } from '@codemirror/language'
  import { EditorState } from '@codemirror/state'
  import { EditorView, lineNumbers } from '@codemirror/view'
  import { classHighlighter } from '@lezer/highlight'
  import { ConflictFile, RepoState, ResolveConflict, TakeConflictSide } from '../../../wailsjs/go/app/App.js'
  import type { git } from '../../../wailsjs/go/models'
  import { session } from '../session.svelte'
  import { languageFor } from './highlight'
  import { hasMarkers, parseConflicts, resolution, type Choice, type ConflictRegion } from './markers'

  let { path, file, onClose }: { path: string; file: string; onClose: () => void } = $props()

  let cf = $state<git.ConflictFile | null>(null)
  let op = $state('')
  let regions = $state<ConflictRegion[]>([])
  let result = $state('')
  let busy = $state(false)

  let oursHost = $state<HTMLDivElement>()
  let theirsHost = $state<HTMLDivElement>()
  let resultHost = $state<HTMLDivElement>()
  let resultView: EditorView | undefined

  // In a rebase git swaps the meaning: "ours" is the branch being rebased onto.
  let oursLabel = $derived(op === 'rebase' ? 'Ours (upstream being rebased onto)' : 'Ours (current branch)')
  let theirsLabel = $derived(op === 'rebase' ? 'Theirs (your commit being replayed)' : 'Theirs (incoming)')

  async function load() {
    const [c, st] = await Promise.all([session.run(() => ConflictFile(path, file)), session.run(() => RepoState(path))])
    if (!c) return
    cf = c
    op = st?.operation ?? ''
    result = c.merged.text
    regions = parseConflicts(result)
  }

  $effect(() => {
    void file
    void load()
  })

  function readOnly(text: string, lang: ReturnType<typeof languageFor> extends Promise<infer L> ? L : never) {
    return EditorState.create({
      doc: text,
      extensions: [
        lineNumbers(),
        EditorView.editable.of(false),
        EditorState.readOnly.of(true),
        syntaxHighlighting(classHighlighter),
        ...(lang ? [lang] : []),
      ],
    })
  }

  // Build the three editors once the file is loaded.
  $effect(() => {
    if (!cf || cf.binary || !oursHost || !theirsHost || !resultHost) return
    const c = cf
    let views: EditorView[] = []
    let cancelled = false
    void languageFor(file).then((lang) => {
      if (cancelled) return
      const ours = new EditorView({ state: readOnly(c.ours.text, lang), parent: oursHost! })
      const theirs = new EditorView({ state: readOnly(c.theirs.text, lang), parent: theirsHost! })
      resultView = new EditorView({
        state: EditorState.create({
          doc: result,
          extensions: [
            lineNumbers(),
            syntaxHighlighting(classHighlighter),
            ...(lang ? [lang] : []),
            // Keep the conflict list in step with manual edits.
            EditorView.updateListener.of((u) => {
              if (u.docChanged) {
                result = u.state.doc.toString()
                regions = parseConflicts(result)
              }
            }),
          ],
        }),
        parent: resultHost!,
      })
      views = [ours, theirs, resultView]
    })
    return () => {
      cancelled = true
      views.forEach((v) => v.destroy())
      resultView = undefined
    }
  })

  function choose(region: ConflictRegion, choice: Choice) {
    if (!resultView) return
    resultView.dispatch({ changes: { from: region.from, to: region.to, insert: resolution(region, choice) } })
  }

  function jump(region: ConflictRegion) {
    resultView?.dispatch({ selection: { anchor: region.from }, scrollIntoView: true })
    resultView?.focus()
  }

  async function markResolved() {
    busy = true
    const ok = await session.run(async () => {
      await ResolveConflict(path, file, result)
      return true
    })
    busy = false
    if (ok) onClose()
  }

  async function take(side: 'ours' | 'theirs') {
    busy = true
    const ok = await session.run(async () => {
      await TakeConflictSide(path, file, side)
      return true
    })
    busy = false
    if (ok) onClose()
  }
</script>

<div class="conflict">
  <header>
    <button class="back" onclick={onClose} title="Back">←</button>
    <span class="file">{file}</span>
    <span class="badge">Conflict</span>
    <span class="spacer"></span>
    <button disabled={busy} onclick={() => take('ours')} title="Resolve with the whole ours version">Use ours</button>
    <button disabled={busy} onclick={() => take('theirs')} title="Resolve with the whole theirs version">Use theirs</button>
    {#if cf && !cf.binary}
      <button
        class="primary"
        disabled={busy || hasMarkers(result)}
        onclick={markResolved}
        title={hasMarkers(result) ? 'Resolve every conflict first' : 'Save the result and mark the file resolved'}
        >Mark resolved</button
      >
    {/if}
  </header>

  {#if !cf}
    <p class="muted">Loading…</p>
  {:else if cf.binary}
    <p class="muted">This file is binary, so it can't be merged line by line. Choose a whole version above.</p>
  {:else}
    <div class="sides">
      <section>
        <h4>{oursLabel}{cf.ours.exists ? '' : ' (deleted)'}</h4>
        <div class="editor" bind:this={oursHost}></div>
      </section>
      <section>
        <h4>{theirsLabel}{cf.theirs.exists ? '' : ' (deleted)'}</h4>
        <div class="editor" bind:this={theirsHost}></div>
      </section>
    </div>
    <div class="strip">
      {#if regions.length === 0}
        <span class="done">No conflicts left. Review the result, then Mark resolved.</span>
      {:else}
        <span>{regions.length} {regions.length === 1 ? 'conflict' : 'conflicts'}:</span>
        {#each regions as r, i (r.from)}
          <span class="region">
            <button class="link" onclick={() => jump(r)}>#{i + 1}</button>
            <button onclick={() => choose(r, 'ours')}>Ours</button>
            <button onclick={() => choose(r, 'theirs')}>Theirs</button>
            <button onclick={() => choose(r, 'both')}>Both</button>
            {#if r.base !== null}<button onclick={() => choose(r, 'base')}>Base</button>{/if}
          </span>
        {/each}
      {/if}
    </div>
    <section class="result">
      <h4>Result (editable)</h4>
      <div class="editor" bind:this={resultHost}></div>
    </section>
  {/if}
</div>

<style>
  .conflict {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 8px;
    border-bottom: 1px solid var(--border);
  }

  .file {
    font-family: var(--font-mono);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .badge {
    font-size: 11px;
    padding: 0 6px;
    border: 1px solid #a33;
    color: #f48771;
  }

  .spacer {
    flex: 1;
  }

  button {
    padding: 2px 8px;
    background: none;
    border: 1px solid var(--border);
    color: var(--fg);
    cursor: pointer;
  }

  button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .primary {
    background: var(--bg-selected);
    border-color: var(--bg-selected);
  }

  .back {
    border: none;
    font-size: 16px;
  }

  .link {
    border: none;
    color: #75beff;
    padding: 0 4px;
  }

  .sides {
    display: flex;
    flex: 1;
    min-height: 0;
    border-bottom: 1px solid var(--border);
  }

  .sides section {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .sides section + section {
    border-left: 1px solid var(--border);
  }

  h4 {
    margin: 0;
    padding: 4px 8px;
    font-size: 12px;
    color: var(--fg-muted);
  }

  .editor {
    flex: 1;
    min-height: 0;
    overflow: auto;
    font-size: 12px;
    tab-size: 4;
  }

  .editor :global(.cm-editor) {
    height: 100%;
    background: var(--bg);
    color: var(--fg);
  }

  .editor :global(.cm-gutters) {
    background: var(--bg-raised);
    color: var(--fg-muted);
    border-right: 1px solid var(--border);
  }

  .editor :global(.cm-content) {
    font-family: var(--font-mono);
    caret-color: var(--fg);
  }

  .strip {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;
    padding: 6px 8px;
    border-bottom: 1px solid var(--border);
    font-size: 12px;
  }

  .region {
    display: inline-flex;
    gap: 3px;
    align-items: center;
  }

  .done {
    color: #89d185;
  }

  .result {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .muted {
    color: var(--fg-muted);
    padding: 8px 12px;
  }
</style>
