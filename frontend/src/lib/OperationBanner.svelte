<script lang="ts">
  import { RepoState, RunOperationAction } from '../../wailsjs/go/app/App.js'
  import type { app } from '../../wailsjs/go/models'
  import { session } from './session.svelte'
  import Icon from './ui/Icon.svelte'

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
    <Icon name="warning" />
    <strong>{labels[opState.operation] ?? opState.operation} in progress</strong>
    {#if opState.total > 0}<span class="meta">Step {opState.step} of {opState.total}</span>{/if}
    {#if opState.conflicts.length > 0}
      <span class="conflicts" title={opState.conflicts.join('\n')}>
        {opState.conflicts.length} conflicted {opState.conflicts.length === 1 ? 'file' : 'files'}
      </span>
    {/if}
    <span class="actions">
      {#each opState.actions as action (action)}
        <button class="tool" class:primary={action === 'continue'} disabled={busy} onclick={() => act(action)}>
          {action === 'abort' && opState.operation === 'bisect' ? 'Stop bisect' : action[0].toUpperCase() + action.slice(1)}
        </button>
      {/each}
    </span>
  </div>
{/if}

<style>
  .op {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 12px;
    align-items: center;
    padding: 6px 8px 6px 14px;
    background: var(--warning-wash);
    border-bottom: 1px solid var(--paper-edge);
  }

  .op > :global(svg) {
    color: var(--warning);
  }

  strong {
    font-weight: 600;
  }

  .conflicts {
    color: var(--danger);
  }

  .actions {
    margin-left: auto;
    display: flex;
    gap: 4px;
  }

  .primary {
    background: var(--accent);
    color: var(--on-accent);
  }

  .primary:hover:not(:disabled) {
    background: var(--accent);
  }
</style>
