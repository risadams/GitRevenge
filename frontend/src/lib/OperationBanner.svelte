<script lang="ts">
  import { RepoState, RunOperationAction } from '../../wailsjs/go/app/App.js'
  import type { app } from '../../wailsjs/go/models'
  import { session } from './session.svelte'

  let { path }: { path: string } = $props()

  let opState = $state<app.RepoState | null>(null)
  let busy = $state(false)

  const labels: Record<string, string> = {
    merge: 'Merge',
    rebase: 'Rebase',
    'cherry-pick': 'Cherry-pick',
    revert: 'Revert',
    bisect: 'Bisect',
    am: 'Applying patches',
  }

  async function refresh() {
    opState = (await session.run(() => RepoState(path))) ?? null
  }

  async function act(action: string) {
    busy = true
    await session.run(() => RunOperationAction(path, action))
    busy = false
    await refresh()
  }

  // Re-fetch when the tab changes or the backend reports a change on disk.
  $effect(() => {
    void session.revision[path]
    void refresh()
  })
</script>

{#if opState?.operation}
  <div class="op" role="status">
    <strong>{labels[opState.operation] ?? opState.operation} in progress</strong>
    {#if opState.total > 0}<span>step {opState.step} of {opState.total}</span>{/if}
    {#if opState.conflicts.length > 0}
      <span class="conflicts" title={opState.conflicts.join('\n')}>
        {opState.conflicts.length} conflicted {opState.conflicts.length === 1 ? 'file' : 'files'}
      </span>
    {/if}
    <span class="actions">
      {#each opState.actions as action (action)}
        <button disabled={busy} onclick={() => act(action)}>
          {action === 'abort' && opState.operation === 'bisect' ? 'Stop bisect' : action[0].toUpperCase() + action.slice(1)}
        </button>
      {/each}
    </span>
  </div>
{/if}

<style>
  .op {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 6px 12px;
    background: #4a3b12;
    border-bottom: 1px solid var(--border);
  }

  .conflicts {
    color: #f0a0a0;
  }

  .actions {
    margin-left: auto;
    display: flex;
    gap: 6px;
  }
</style>
