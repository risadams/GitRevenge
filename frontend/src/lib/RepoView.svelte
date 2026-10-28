<script lang="ts">
  import { CommitRow } from '../../wailsjs/go/app/App.js'
  import CommitPanel from './CommitPanel.svelte'
  import GraphView from './graph/GraphView.svelte'
  import type { Ref, Row } from './graph/rowCache'
  import { session } from './session.svelte'
  import Sidebar from './Sidebar.svelte'

  let { path }: { path: string } = $props()

  let graph: GraphView
  let selected = $state<Row | undefined>()

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
    <GraphView bind:this={graph} {path} onSelect={(row) => (selected = row)} />
  </div>
  {#if selected}
    <CommitPanel {path} commitId={selected.id} />
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
  }
</style>
