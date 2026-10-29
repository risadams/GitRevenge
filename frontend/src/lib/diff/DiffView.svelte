<script lang="ts" module>
  /** What to compare. Opened from the commit panel or the working tree panel. */
  export interface DiffTarget {
    /** 'conflict' opens the conflict tool instead of a diff. */
    mode: 'commit' | 'staged' | 'unstaged' | 'conflict' | 'history' | 'blame'
    file: string
    oldPath?: string
    commit?: string
  }
</script>

<script lang="ts">
  import { FileContents, SetStagedContent, SetWorkTreeContent } from '../../../wailsjs/go/app/App.js'
  import type { git } from '../../../wailsjs/go/models'
  import { session } from '../session.svelte'
  import ImageDiff from './ImageDiff.svelte'
  import SplitDiff from './SplitDiff.svelte'
  import UnifiedDiff, { type Actions } from './UnifiedDiff.svelte'

  let {
    path,
    target,
    onClose,
    onOpen,
  }: {
    path: string
    target: DiffTarget
    onClose: () => void
    /** Opens another view of the same file (history or blame). */
    onOpen?: (t: DiffTarget) => void
  } = $props()

  const LAYOUT_KEY = 'gitrevenge.diffLayout'
  function savedLayout(): 'unified' | 'split' {
    try {
      return localStorage.getItem(LAYOUT_KEY) === 'split' ? 'split' : 'unified'
    } catch {
      return 'unified'
    }
  }
  let layout = $state<'unified' | 'split'>(savedLayout())
  function setLayout(l: 'unified' | 'split') {
    layout = l
    try {
      localStorage.setItem(LAYOUT_KEY, l)
    } catch {
      // Storage unavailable; the choice lasts for this session only.
    }
  }

  let contents = $state<git.FileContents | null>(null)
  let force = $state(false)

  async function load() {
    const t = target
    const res = await session.run(() =>
      FileContents(path, t.mode, t.commit ?? '', t.file, t.oldPath ?? '', force),
    )
    if (res && t === target) contents = res
  }

  // Reload when the target changes, and when the repo changes on disk (for work tree diffs).
  $effect(() => {
    void target
    void force
    if (target.mode !== 'commit') void session.revision[path]
    void load()
  })

  // Reset per-file state when switching files.
  $effect(() => {
    void target
    force = false
  })

  let actions = $derived.by((): Actions => {
    const file = target.file
    const shownOld = contents?.old.text ?? ''
    const shownNew = contents?.new.text ?? ''
    // After any write, reload so the view reflects the new state immediately.
    const after = async (p: Promise<void>) => {
      await session.run(() => p)
      await load()
    }
    if (target.mode === 'unstaged') {
      return {
        apply: { label: 'Stage', run: (c) => after(SetStagedContent(path, file, shownOld, c)) },
        revert: { label: 'Discard', danger: true, run: (c) => after(SetWorkTreeContent(path, file, shownNew, c)) },
      }
    }
    if (target.mode === 'staged') {
      // Unstaging rebuilds the index from HEAD plus the changes being kept.
      return { revert: { label: 'Unstage', run: (c) => after(SetStagedContent(path, file, shownNew, c)) } }
    }
    return {}
  })

  const modeLabel = { commit: '', staged: 'Staged', unstaged: 'Unstaged', conflict: 'Conflict', history: '', blame: '' }
</script>

<div class="diffview">
  <header>
    <button class="back" onclick={onClose} title="Back to the graph">←</button>
    <span class="file" title={target.file}
      >{target.oldPath && target.oldPath !== target.file ? `${target.oldPath} → ` : ''}{target.file}</span
    >
    {#if modeLabel[target.mode]}<span class="badge">{modeLabel[target.mode]}</span>{/if}
    {#if onOpen}
      <button onclick={() => onOpen({ mode: 'history', file: target.file })}>History</button>
      <button
        onclick={() =>
          onOpen({ mode: 'blame', file: target.file, commit: target.mode === 'commit' ? target.commit : undefined })}
        >Blame</button
      >
    {/if}
    {#if contents?.kind === 'text'}
      <span class="layout">
        <button class:on={layout === 'unified'} onclick={() => setLayout('unified')}>Unified</button>
        <button class:on={layout === 'split'} onclick={() => setLayout('split')}>Split</button>
      </span>
    {/if}
  </header>

  <div class="body">
    {#if !contents}
      <p class="muted">Loading…</p>
    {:else if contents.kind === 'large'}
      <div class="notice">
        <p>
          This file is large ({(Math.max(contents.old.size, contents.new.size) / (1 << 20)).toFixed(1)} MB). Showing it may be
          slow.
        </p>
        <button onclick={() => (force = true)}>Load anyway</button>
      </div>
    {:else if contents.kind === 'binary'}
      <p class="muted">Binary file. Sizes: {contents.old.exists ? contents.old.size : '—'} → {contents.new.exists ? contents.new.size : '—'} bytes.</p>
    {:else if contents.kind === 'image'}
      <ImageDiff {contents} />
    {:else if layout === 'split'}
      <SplitDiff path={target.file} oldText={contents.old.text} newText={contents.new.text} />
      {#if target.mode !== 'commit'}
        <p class="muted hint">Switch to Unified to stage or discard individual hunks and lines.</p>
      {/if}
    {:else}
      <UnifiedDiff path={target.file} oldText={contents.old.text} newText={contents.new.text} {actions} />
    {/if}
  </div>
</div>

<style>
  .diffview {
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
    border-bottom: 1px solid var(--paper-edge);
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
    border: 1px solid var(--paper-edge);
    color: var(--ink-mute);
  }

  .layout {
    margin-left: auto;
    display: flex;
    gap: 4px;
  }

  button {
    background: none;
    border: 1px solid var(--paper-edge);
    color: var(--ink);
    cursor: pointer;
    padding: 2px 8px;
  }

  button.on {
    background: var(--selected);
  }

  .back {
    border: none;
    font-size: 16px;
  }

  .body {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  .notice {
    padding: 16px;
  }

  .muted {
    color: var(--ink-mute);
    padding: 8px 12px;
  }

  .hint {
    font-size: 12px;
  }
</style>
