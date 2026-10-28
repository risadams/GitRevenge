<script lang="ts" module>
  /** Unsent commit messages per repository, kept while switching tabs. A plain cache that is
   * never rendered, so it needn't be reactive. */
  // eslint-disable-next-line svelte/prefer-svelte-reactivity
  const drafts = new Map<string, { summary: string; body: string }>()
</script>

<script lang="ts">
  import {
    CommitChanges,
    DiscardFiles,
    LastCommitMessage,
    StageFiles,
    UnstageFiles,
    WorkingTree,
  } from '../../wailsjs/go/app/App.js'
  import type { app, git } from '../../wailsjs/go/models'
  import type { DiffTarget } from './diff/DiffView.svelte'
  import { session } from './session.svelte'

  let {
    path,
    openFile,
    current,
    onCommitted,
  }: {
    path: string
    openFile: (t: DiffTarget | null) => void
    /** The diff currently open, to highlight its file. */
    current?: DiffTarget | null
    onCommitted?: (id: string) => void
  } = $props()

  let wt = $state<app.WorkingTree | null>(null)
  let confirmDiscard = $state<string[] | null>(null)
  let busy = $state(false)

  // Read once: the repo view is keyed by path, so this panel never switches repositories.
  // svelte-ignore state_referenced_locally
  const draft = drafts.get(path) ?? { summary: '', body: '' }
  let summary = $state(draft.summary)
  let body = $state(draft.body)
  let amend = $state(false)
  /** The draft set aside while amending, restored if amend is unticked. */
  let stash: { summary: string; body: string } | null = null

  $effect(() => {
    if (!amend) drafts.set(path, { summary, body })
  })

  let files = $derived(wt?.status.files ?? [])
  let unstaged = $derived(files.filter((f) => f.untracked || f.conflicted || f.workTree !== '.'))
  let staged = $derived(files.filter((f) => !f.untracked && !f.conflicted && f.index !== '.'))
  let canCommit = $derived(summary.trim() !== '' && (staged.length > 0 || amend) && !busy)

  async function refresh() {
    const res = await session.run(() => WorkingTree(path))
    if (!res) return
    wt = res
  }

  // Refresh on open and whenever the backend reports a change on disk.
  $effect(() => {
    void session.revision[path]
    void refresh()
  })

  function show(f: git.FileStatus, inStaged: boolean) {
    openFile({ mode: inStaged ? 'staged' : 'unstaged', file: f.path, oldPath: f.origPath })
  }

  const isOpen = (f: git.FileStatus, inStaged: boolean) =>
    current?.file === f.path && current.mode === (inStaged ? 'staged' : 'unstaged')

  async function act(fn: () => Promise<void>) {
    busy = true
    await session.run(fn)
    busy = false
    await refresh()
  }

  const stage = (paths: string[]) => act(() => StageFiles(path, paths))
  const unstage = (paths: string[]) => act(() => UnstageFiles(path, paths))

  async function discard(paths: string[]) {
    confirmDiscard = null
    if (current && paths.includes(current.file)) openFile(null)
    await act(() => DiscardFiles(path, paths))
  }

  async function toggleAmend() {
    if (amend) {
      stash = { summary, body }
      const msg = (await session.run(() => LastCommitMessage(path))) ?? ''
      const nl = msg.indexOf('\n')
      summary = nl < 0 ? msg : msg.slice(0, nl)
      body = nl < 0 ? '' : msg.slice(nl + 1).trim()
    } else if (stash) {
      ;({ summary, body } = stash)
      stash = null
    }
  }

  async function commit() {
    if (!canCommit) return
    busy = true
    const message = body.trim() ? `${summary.trim()}\n\n${body.trim()}` : summary.trim()
    const id = await session.run(() => CommitChanges(path, message, amend))
    busy = false
    if (id) {
      summary = ''
      body = ''
      amend = false
      stash = null
      drafts.delete(path)
      if (current?.mode !== 'commit') openFile(null)
      onCommitted?.(id)
    }
    await refresh()
  }

  function onKey(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault()
      void commit()
    }
  }

  function code(f: git.FileStatus, inStaged: boolean): string {
    if (f.untracked) return 'A'
    if (f.conflicted) return '!'
    return inStaged ? f.index : f.workTree
  }
</script>

{#snippet fileRow(f: git.FileStatus, inStaged: boolean)}
  <li class:active={isOpen(f, inStaged)}>
    <button class="name" onclick={() => show(f, inStaged)} title={f.origPath ? `${f.origPath} → ${f.path}` : f.path}>
      <span class="code c-{code(f, inStaged)}">{code(f, inStaged)}</span>
      <span class="path">{f.origPath ? `${f.origPath} → ${f.path}` : f.path}</span>
    </button>
    {#if confirmDiscard?.length === 1 && confirmDiscard[0] === f.path}
      <span class="confirm">
        Discard?
        <button class="danger" onclick={() => discard([f.path])}>Yes</button>
        <button onclick={() => (confirmDiscard = null)}>No</button>
      </span>
    {:else if inStaged}
      <button class="icon" disabled={busy} onclick={() => unstage([f.path])} title="Unstage">−</button>
    {:else}
      {#if !f.conflicted}
        <button class="icon" disabled={busy} onclick={() => (confirmDiscard = [f.path])} title="Discard changes">↺</button>
      {/if}
      <button class="icon" disabled={busy} onclick={() => stage([f.path])} title="Stage">+</button>
    {/if}
  </li>
{/snippet}

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<aside class="panel" onkeydown={onKey} role="region" aria-label="Uncommitted changes">
  <div class="lists">
    <section>
      <header>
        <span>Unstaged ({unstaged.length})</span>
        {#if unstaged.length}
          {#if confirmDiscard && confirmDiscard.length > 1}
            <span class="confirm">
              Discard all unstaged changes?
              <button class="danger" onclick={() => discard(confirmDiscard ?? [])}>Yes</button>
              <button onclick={() => (confirmDiscard = null)}>No</button>
            </span>
          {:else}
            <button
              disabled={busy}
              onclick={() => (confirmDiscard = unstaged.filter((f) => !f.conflicted).map((f) => f.path))}
              title="Discard all unstaged changes">Discard all</button
            >
            <button disabled={busy} onclick={() => stage([])}>Stage all</button>
          {/if}
        {/if}
      </header>
      <ul>
        {#each unstaged as f (f.path)}{@render fileRow(f, false)}{/each}
      </ul>
    </section>
    <section>
      <header>
        <span>Staged ({staged.length})</span>
        {#if staged.length}<button disabled={busy} onclick={() => unstage([])}>Unstage all</button>{/if}
      </header>
      <ul>
        {#each staged as f (f.path)}{@render fileRow(f, true)}{/each}
      </ul>
    </section>
  </div>

  <form
    class="composer"
    onsubmit={(e) => {
      e.preventDefault()
      void commit()
    }}
  >
    <div class="summary">
      <input bind:value={summary} placeholder="Summary (required)" aria-label="Commit summary" />
      <span class="count" class:long={summary.length > 72} title="Keep summaries under about 72 characters"
        >{summary.length}</span
      >
    </div>
    <textarea bind:value={body} placeholder="Description" rows="4" aria-label="Commit description"></textarea>
    <div class="row">
      <label><input type="checkbox" bind:checked={amend} onchange={toggleAmend} /> Amend last commit</label>
      {#if wt?.signing.enabled}
        <span class="signed" title="commit.gpgsign is set; key: {wt.signing.key || 'default'}"
          >Signed ({wt.signing.format})</span
        >
      {/if}
    </div>
    <button type="submit" class="primary" disabled={!canCommit}>
      {amend ? 'Amend commit' : `Commit ${staged.length} ${staged.length === 1 ? 'file' : 'files'}`}
    </button>
  </form>
</aside>

<style>
  .panel {
    display: flex;
    flex-direction: column;
    width: 420px;
    min-width: 300px;
    border-left: 1px solid var(--border);
    background: var(--bg-raised);
  }

  .lists {
    flex: 1;
    overflow-y: auto;
  }

  header {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 12px 4px;
    font-weight: 600;
  }

  header > span:first-child {
    flex: 1;
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  li {
    display: flex;
    align-items: center;
    padding-right: 8px;
  }

  li:hover {
    background: var(--bg-hover);
  }

  li.active {
    background: var(--bg-selected);
  }

  button {
    background: none;
    border: 1px solid var(--border);
    color: inherit;
    font: inherit;
    cursor: pointer;
    padding: 1px 6px;
  }

  button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .name {
    flex: 1;
    display: flex;
    gap: 8px;
    min-width: 0;
    border: none;
    text-align: left;
    padding: 3px 12px;
  }

  .path {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .icon {
    border: none;
    width: 22px;
  }

  .code {
    width: 12px;
    font-family: var(--font-mono);
    font-weight: 600;
  }

  .c-A {
    color: #89d185;
  }

  .c-D {
    color: #f48771;
  }

  .c-M,
  .c-T {
    color: #e2c08d;
  }

  .c-R,
  .c-C {
    color: #75beff;
  }

  .c-\! {
    color: #ff5370;
  }

  .confirm {
    display: flex;
    gap: 4px;
    align-items: center;
    font-weight: normal;
    font-size: 12px;
  }

  .danger {
    border-color: #a33;
    color: #f48771;
  }

  .composer {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 10px 12px;
    border-top: 1px solid var(--border);
  }

  .summary {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  input:not([type='checkbox']),
  textarea {
    flex: 1;
    padding: 4px 6px;
    background: var(--bg);
    color: var(--fg);
    border: 1px solid var(--border);
    font: inherit;
    resize: vertical;
  }

  .count {
    color: var(--fg-muted);
    font-size: 11px;
    min-width: 20px;
    text-align: right;
  }

  .count.long {
    color: #e2c08d;
  }

  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
  }

  .signed {
    color: #89d185;
  }

  .primary {
    padding: 6px;
    background: var(--bg-selected);
    border-color: var(--bg-selected);
  }
</style>
