<script lang="ts">
  import { FileHistory } from '../../../wailsjs/go/app/App.js'
  import type { git } from '../../../wailsjs/go/models'
  import { formatTime } from '../graph/draw'
  import { session } from '../session.svelte'
  import DiffView, { type DiffTarget } from './DiffView.svelte'

  let {
    path,
    file,
    onClose,
    onOpen,
  }: { path: string; file: string; onClose: () => void; onOpen: (t: DiffTarget) => void } = $props()

  let commits = $state<git.FileCommit[] | null>(null)
  let selected = $state<git.FileCommit | null>(null)

  $effect(() => {
    const f = file
    commits = null
    selected = null
    void session.run(() => FileHistory(path, f)).then((c) => {
      if (f !== file) return
      commits = c ?? []
      selected = commits[0] ?? null
    })
  })
</script>

<div class="history">
  <header>
    <button class="back" onclick={onClose} title="Back">←</button>
    <span class="file">History of {file}</span>
    {#if commits}<span class="muted">{commits.length} {commits.length === 1 ? 'commit' : 'commits'}</span>{/if}
    <span class="spacer"></span>
    <button onclick={() => onOpen({ mode: 'blame', file })}>Blame</button>
  </header>
  <ul class="list">
    {#if !commits}
      <li class="muted pad">Loading…</li>
    {:else}
      {#each commits as c (c.id)}
        <li>
          <button class:active={selected?.id === c.id} onclick={() => (selected = c)}>
            <span class="sha">{c.id.slice(0, 7)}</span>
            <span class="subject">{c.subject}</span>
            {#if c.oldPath}<span class="badge" title="Renamed from {c.oldPath}">renamed</span>{/if}
            <span class="muted">{c.author}</span>
            <span class="muted when">{formatTime(c.time)}</span>
          </button>
        </li>
      {:else}
        <li class="muted pad">No commits touch this file.</li>
      {/each}
    {/if}
  </ul>
  {#if selected}
    {#key selected.id}
      <div class="diff">
        <DiffView
          {path}
          target={{ mode: 'commit', commit: selected.id, file: selected.path, oldPath: selected.oldPath }}
          onClose={() => (selected = null)}
        />
      </div>
    {/key}
  {/if}
</div>

<style>
  .history {
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

  .spacer {
    flex: 1;
  }

  header button {
    padding: 2px 8px;
    background: none;
    border: 1px solid var(--border);
    color: var(--fg);
    cursor: pointer;
  }

  .back {
    border: none !important;
    font-size: 16px;
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    max-height: 35%;
    overflow-y: auto;
    border-bottom: 1px solid var(--border);
  }

  .list button {
    display: flex;
    gap: 10px;
    align-items: baseline;
    width: 100%;
    padding: 4px 8px;
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .list button:hover {
    background: var(--bg-hover);
  }

  .list button.active {
    background: var(--bg-selected);
  }

  .sha {
    font-family: var(--font-mono);
    color: var(--fg-muted);
  }

  .subject {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .when {
    white-space: nowrap;
  }

  .badge {
    font-size: 10px;
    padding: 0 4px;
    border: 1px solid var(--border);
    color: #75beff;
  }

  .diff {
    flex: 1;
    min-height: 0;
  }

  .muted {
    color: var(--fg-muted);
  }

  .pad {
    padding: 8px;
  }
</style>
