<script lang="ts">
  import { CommitRow, WorkingTree } from '../../wailsjs/go/app/App.js'
  import CommitPanel from './CommitPanel.svelte'
  import DiffView, { type DiffTarget } from './diff/DiffView.svelte'
  import GraphView from './graph/GraphView.svelte'
  import type { Ref, Row } from './graph/rowCache'
  import { session } from './session.svelte'
  import Sidebar from './Sidebar.svelte'
  import WorkingTreePanel from './WorkingTreePanel.svelte'

  let { path }: { path: string } = $props()

  let graph: GraphView
  let selected = $state<Row | undefined>()
  /** Which panel is on the right: uncommitted changes, or the selected commit. */
  let view = $state<'wip' | 'commit'>('commit')
  let changes = $state(0)
  /** A file diff shown in place of the graph, or null for the graph. */
  let diff = $state<DiffTarget | null>(null)
  let checkedInitial = false

  async function refreshCount() {
    const wt = await session.run(() => WorkingTree(path))
    if (!wt) return
    changes = wt.status.files.length
    // Like GitKraken: open on the changes when there are any.
    if (!checkedInitial) {
      checkedInitial = true
      if (changes > 0) view = 'wip'
    }
  }

  $effect(() => {
    void session.revision[path]
    void refreshCount()
  })

  function onSelect(row: Row | undefined, auto: boolean) {
    selected = row
    if (!auto) diff = null
    // The automatic HEAD selection on load must not pull focus away from the changes.
    if (!auto) view = 'commit'
  }

  async function pick(ref: Ref) {
    const row = await session.run(() => CommitRow(path, ref.commit))
    if (row === undefined) return
    if (row < 0) {
      session.error = `${ref.short} points at a commit that isn't in the graph yet.`
      return
    }
    graph.reveal(row)
  }
</script>

<div class="repo">
  <Sidebar {path} onPick={pick} />
  <div class="main">
    <button
      class="wip"
      class:active={view === 'wip'}
      onclick={() => {
        view = 'wip'
        if (diff?.mode === 'commit') diff = null
      }}
    >
      <span class="dot" class:dirty={changes > 0}></span>
      {#if changes > 0}
        Working tree · {changes} {changes === 1 ? 'change' : 'changes'}
      {:else}
        Working tree · clean
      {/if}
    </button>
    <!-- The graph stays mounted while a diff is open, so scroll position and selection
         survive returning to it. -->
    <div class="graph" class:hidden={diff}>
      <GraphView bind:this={graph} {path} {onSelect} />
    </div>
    {#if diff}
      <div class="graph">
        <DiffView {path} target={diff} onClose={() => (diff = null)} />
      </div>
    {/if}
  </div>
  {#if view === 'wip'}
    <WorkingTreePanel {path} current={diff} openFile={(t) => (diff = t)} />
  {:else if selected}
    <CommitPanel {path} commitId={selected.id} current={diff} openFile={(t) => (diff = t)} />
  {/if}
</div>

<style>
  .repo {
    display: flex;
    height: 100%;
    min-height: 0;
  }

  .main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .graph {
    flex: 1;
    min-height: 0;
  }

  .hidden {
    display: none;
  }

  .wip {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    background: none;
    border: none;
    border-bottom: 1px solid var(--border);
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .wip:hover {
    background: var(--bg-hover);
  }

  .wip.active {
    background: var(--bg-selected);
  }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 2px dashed var(--fg-muted);
  }

  .dot.dirty {
    border-color: #e2c08d;
  }
</style>
