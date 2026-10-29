<script lang="ts">
  import { CommitDetail } from '../../wailsjs/go/app/App.js'
  import type { app, git } from '../../wailsjs/go/models'
  import { formatTime } from './graph/draw'
  import type { DiffTarget } from './diff/DiffView.svelte'
  import { session } from './session.svelte'

  let {
    path,
    commitId,
    openFile,
    current,
  }: {
    path: string
    commitId: string
    openFile: (t: DiffTarget) => void
    /** The diff currently open, to highlight its file. */
    current?: DiffTarget | null
  } = $props()

  let info = $state<app.CommitInfo | null>(null)

  const statusNames: Record<string, string> = {
    A: 'Added',
    M: 'Modified',
    D: 'Deleted',
    R: 'Renamed',
    C: 'Copied',
    T: 'Type changed',
  }

  // Load whenever the selected commit changes; drop answers for a commit no longer selected.
  $effect(() => {
    const id = commitId
    info = null
    void session.run(() => CommitDetail(path, id)).then((res) => {
      if (res && id === commitId) info = res
    })
  })

  function showFile(f: git.FileChange) {
    openFile({ mode: 'commit', commit: commitId, file: f.path, oldPath: f.oldPath })
  }
</script>

<aside class="panel">
  {#if !info}
    <p class="muted pad">Loading…</p>
  {:else}
    {@const c = info.commit}
    <header class="pad">
      <h3>{c.subject}</h3>
      {#if c.body}<pre class="body">{c.body}</pre>{/if}
      <dl>
        <dt>Author</dt>
        <dd>{c.author.name} &lt;{c.author.email}&gt;<br /><span class="muted">{formatTime(c.author.time)}</span></dd>
        {#if c.committer.name !== c.author.name || c.committer.time !== c.author.time}
          <dt>Committer</dt>
          <dd>{c.committer.name}<br /><span class="muted">{formatTime(c.committer.time)}</span></dd>
        {/if}
        <dt>Commit</dt>
        <dd class="mono">{c.id}</dd>
        {#if c.parents.length}
          <dt>{c.parents.length > 1 ? 'Parents' : 'Parent'}</dt>
          <dd class="mono">{c.parents.map((p) => p.slice(0, 10)).join(' ')}</dd>
        {/if}
      </dl>
      {#if c.parents.length > 1}
        <p class="muted note">Changes are shown against the first parent.</p>
      {/if}
    </header>

    <h4 class="pad">{info.files.length} {info.files.length === 1 ? 'file' : 'files'} changed</h4>
    <ul class="files">
      {#each info.files as f (f.path)}
        <li>
          <button class:active={current?.mode === 'commit' && current.commit === commitId && current.file === f.path} onclick={() => showFile(f)} title={statusNames[f.status] ?? f.status}>
            <span class="status s-{f.status}">{f.status}</span>
            <span class="file">{f.oldPath ? `${f.oldPath} → ${f.path}` : f.path}</span>
          </button>
        </li>
      {/each}
    </ul>

  {/if}
</aside>

<style>
  .panel {
    display: flex;
    flex-direction: column;
    width: 420px;
    min-width: 300px;
    border-left: 1px solid var(--paper-edge);
    background: var(--paper-warm);
    overflow-y: auto;
  }

  .pad {
    padding: 0 12px;
  }

  h3 {
    margin: 12px 0 6px;
    font-size: 14px;
  }

  h4 {
    margin: 12px 0 4px;
    font-size: 12px;
    color: var(--ink-mute);
  }

  .body {
    white-space: pre-wrap;
    font-family: inherit;
    margin: 0 0 8px;
  }

  dl {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 4px 10px;
    margin: 8px 0;
  }

  dt {
    color: var(--ink-mute);
  }

  dd {
    margin: 0;
    overflow-wrap: anywhere;
  }

  .mono,
  pre {
    font-family: var(--font-mono);
    font-size: 12px;
  }

  .note {
    font-size: 12px;
  }

  .files {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .files button {
    display: flex;
    gap: 8px;
    width: 100%;
    padding: 3px 12px;
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .files button:hover {
    background: var(--wash-hover);
  }

  .files button.active {
    background: var(--selected);
  }

  .status {
    width: 12px;
    font-family: var(--font-mono);
    font-weight: 600;
  }

  .s-A {
    color: var(--success);
  }

  .s-D {
    color: var(--danger);
  }

  .s-M,
  .s-T {
    color: var(--accent-c);
  }

  .s-R,
  .s-C {
    color: var(--accent-a);
  }

  .file {
    overflow-wrap: anywhere;
  }

  .muted {
    color: var(--ink-mute);
  }
</style>
