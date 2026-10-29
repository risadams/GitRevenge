<script lang="ts">
  import { onMount } from 'svelte'
  import { CancelNetwork, Fetch, Pull, PushCurrent, Redo, Undo, UndoState } from '../../wailsjs/go/app/App.js'
  import { EventsOn } from '../../wailsjs/runtime/runtime.js'
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte'
  import { session } from './session.svelte'

  let { path }: { path: string } = $props()

  /** The running network operation, if any. */
  let running = $state<string | null>(null)
  let progress = $state('')
  let menu = $state<{ x: number; y: number; items: MenuItem[] } | null>(null)

  const labels: Record<string, string> = { fetch: 'Fetching', pull: 'Pulling', push: 'Pushing' }

  /** Labels of the operations Undo and Redo would reverse; "" when unavailable. */
  let undoLabel = $state('')
  let redoLabel = $state('')

  async function refreshUndo() {
    const info = await UndoState(path).catch(() => null)
    undoLabel = info?.undo ?? ''
    redoLabel = info?.redo ?? ''
  }

  // Re-read after anything changes on disk (including our own operations).
  $effect(() => {
    void session.revision[path]
    void refreshUndo()
  })

  async function undo() {
    if (!undoLabel || running) return
    await session.run(() => Undo(path))
    await refreshUndo()
  }

  async function redo() {
    if (!redoLabel || running) return
    await session.run(() => Redo(path))
    await refreshUndo()
  }

  function onKey(e: KeyboardEvent) {
    // Leave text editing shortcuts to the field being typed in.
    const t = e.target as HTMLElement
    if (t.closest('input, textarea, select, [contenteditable]')) return
    if (!(e.ctrlKey || e.metaKey)) return
    const k = e.key.toLowerCase()
    if (k === 'z' && !e.shiftKey) {
      e.preventDefault()
      void undo()
    } else if ((k === 'z' && e.shiftKey) || k === 'y') {
      e.preventDefault()
      void redo()
    }
  }

  async function run(op: string, fn: () => Promise<void>) {
    running = op
    progress = ''
    await session.run(fn)
    running = null
    progress = ''
    await refreshUndo()
  }

  const fetchAll = () => run('fetch', () => Fetch(path, ''))
  const pull = (mode: string) => run('pull', () => Pull(path, mode))
  const push = (force: boolean) => run('push', () => PushCurrent(path, force))

  function dropdown(e: MouseEvent, items: MenuItem[]) {
    const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
    menu = { x: r.left, y: r.bottom + 2, items }
  }

  onMount(() =>
    EventsOn('net:progress', (p: { path: string; phase: string; percent: number; line: string }) => {
      if (p.path !== path) return
      progress = p.phase ? `${p.phase} ${p.percent}%` : p.line
    }),
  )
</script>

<svelte:window onkeydown={onKey} />

<div class="toolbar">
  <button disabled={!undoLabel || !!running} onclick={undo} title={undoLabel ? `Undo: ${undoLabel} (Ctrl+Z)` : 'Nothing to undo'}
    >Undo</button
  >
  <button disabled={!redoLabel || !!running} onclick={redo} title={redoLabel ? `Redo: ${redoLabel} (Ctrl+Shift+Z)` : 'Nothing to redo'}
    >Redo</button
  >
  <span class="sep"></span>
  <button disabled={!!running} onclick={fetchAll} title="Fetch all remotes">Fetch</button>
  <span class="split">
    <button disabled={!!running} onclick={() => pull('')} title="Pull using your git config (pull.rebase)">Pull</button>
    <button
      class="caret"
      disabled={!!running}
      aria-label="Pull options"
      onclick={(e) =>
        dropdown(e, [
          { label: 'Pull (fast-forward only)', run: () => pull('ff-only') },
          { label: 'Pull (merge)', run: () => pull('merge') },
          { label: 'Pull (rebase)', run: () => pull('rebase') },
        ])}>▾</button
    >
  </span>
  <span class="split">
    <button disabled={!!running} onclick={() => push(false)} title="Push the current branch">Push</button>
    <button
      class="caret"
      disabled={!!running}
      aria-label="Push options"
      onclick={(e) =>
        dropdown(e, [{ label: 'Force push (with lease)', danger: true, run: () => push(true) }])}>▾</button
    >
  </span>

  {#if running}
    <span class="progress" role="status">
      <span class="spinner"></span>
      <span class="text">{labels[running]}{progress ? `: ${progress}` : '…'}</span>
      <button onclick={() => CancelNetwork(path)}>Cancel</button>
    </span>
  {/if}
</div>

{#if menu}
  <ContextMenu x={menu.x} y={menu.y} items={menu.items} onClose={() => (menu = null)} />
{/if}

<style>
  .toolbar {
    display: flex;
    white-space: nowrap;
    overflow: hidden;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-bottom: 1px solid var(--border);
  }

  button {
    padding: 2px 10px;
    background: none;
    border: 1px solid var(--border);
    color: var(--fg);
    cursor: pointer;
  }

  button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .split {
    display: inline-flex;
  }

  .sep {
    width: 1px;
    height: 18px;
    background: var(--border);
    margin: 0 4px;
  }

  .split button:first-child {
    border-right: none;
  }

  .caret {
    padding: 2px 5px;
  }

  .progress {
    margin-left: auto;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--fg-muted);
    font-size: 12px;
    white-space: nowrap;
  }

  /* Only the text shrinks, so Cancel always stays clickable. */
  .progress .text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .progress button {
    flex-shrink: 0;
  }

  .spinner {
    width: 10px;
    height: 10px;
    border: 2px solid var(--fg-muted);
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
