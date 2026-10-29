<script lang="ts">
  import { onMount } from 'svelte'
  import { GroupBatch, GroupDashboard } from '../../wailsjs/go/app/App.js'
  import type { app } from '../../wailsjs/go/models'
  import { EventsOn } from '../../wailsjs/runtime/runtime.js'
  import { formatTime } from './graph/draw'
  import { session } from './session.svelte'

  let { name, onClose }: { name: string; onClose: () => void } = $props()

  let rows = $state<app.GroupRepo[]>([])
  let loading = $state(true)
  let running = $state<'' | 'fetch' | 'pull'>('')
  /** Result of the last batch, by repo path. */
  let results = $state<Record<string, app.BatchResult>>({})

  async function refresh() {
    loading = true
    rows = (await session.run(() => GroupDashboard(name))) ?? []
    loading = false
  }

  async function batch(op: 'fetch' | 'pull') {
    running = op
    results = {}
    const res = await session.run(() => GroupBatch(name, op))
    for (const r of res ?? []) results[r.path] = r
    running = ''
    await refresh()
  }

  async function openAll() {
    for (const r of rows) if (r.exists) await session.open(r.path)
  }

  onMount(() => {
    void refresh()
    // Show each repository's result as soon as it finishes.
    return EventsOn('group:progress', (p: app.BatchResult & { group: string }) => {
      if (p.group === name) results[p.path] = p
    })
  })
</script>

<section class="dashboard">
  <header>
    <button class="back" onclick={onClose} title="Back">←</button>
    <h2>{name}</h2>
    <span class="spacer"></span>
    <button disabled={!!running} onclick={refresh}>Refresh</button>
    <button disabled={!!running || rows.length === 0} onclick={() => batch('fetch')}>{running === 'fetch' ? 'Fetching…' : 'Fetch all'}</button>
    <button disabled={!!running || rows.length === 0} onclick={() => batch('pull')} title="Fast-forward only: diverged repositories are reported, never merged"
      >{running === 'pull' ? 'Pulling…' : 'Pull all'}</button
    >
    <button disabled={rows.length === 0} onclick={openAll}>Open all as tabs</button>
  </header>

  {#if loading && rows.length === 0}
    <p class="muted">Loading…</p>
  {:else if rows.length === 0}
    <p class="muted">This group has no repositories yet. Add some from the start page.</p>
  {:else}
    <table>
      <thead>
        <tr><th>Repository</th><th>Branch</th><th>Sync</th><th>Changes</th><th>Last fetch</th><th></th></tr>
      </thead>
      <tbody>
        {#each rows as r (r.path)}
          {@const res = results[r.path]}
          <tr class:missing={!r.exists}>
            <td title={r.path}><strong>{r.name}</strong></td>
            <td>{r.branch || '—'}</td>
            <td>
              {#if !r.upstream && r.exists}<span class="muted">no upstream</span>
              {:else if r.ahead || r.behind}{#if r.ahead}↑{r.ahead} {/if}{#if r.behind}↓{r.behind}{/if}
              {:else if r.exists}<span class="ok">up to date</span>{/if}
            </td>
            <td>
              {#if r.conflicts}<span class="bad">{r.conflicts} conflicted</span>
              {:else if r.changes}{r.changes}
              {:else if r.exists}<span class="muted">clean</span>{/if}
            </td>
            <td class="muted">{r.lastFetch ? formatTime(r.lastFetch) : r.exists ? 'never' : ''}</td>
            <td class="action">
              {#if res?.skipped}
                <span class="muted" title={res.message}>– {res.message}</span>
              {:else if res}
                <span class={res.ok ? 'ok' : 'bad'} title={res.message}>{res.ok ? '✓' : '✕'} {res.ok ? '' : res.message}</span>
              {:else if r.error}
                <span class="bad">{r.error}</span>
              {:else}
                <button onclick={() => session.open(r.path)}>Open</button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</section>

<style>
  .dashboard {
    max-width: 1000px;
    margin: 24px auto;
    padding: 0 16px;
  }

  header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  h2 {
    margin: 0;
  }

  .spacer {
    flex: 1;
  }

  button {
    padding: 3px 10px;
    background: none;
    border: 1px solid var(--paper-edge);
    color: var(--ink);
    cursor: pointer;
  }

  button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .back {
    border: none;
    font-size: 16px;
  }

  table {
    width: 100%;
    margin-top: 16px;
    border-collapse: collapse;
  }

  th {
    text-align: left;
    color: var(--ink-mute);
    font-weight: normal;
    font-size: 12px;
    padding: 4px 8px;
    border-bottom: 1px solid var(--paper-edge);
  }

  td {
    padding: 6px 8px;
    border-bottom: 1px solid var(--paper-edge);
    white-space: nowrap;
  }

  td.action {
    white-space: normal;
  }

  tr.missing {
    opacity: 0.6;
  }

  .ok {
    color: var(--success);
  }

  .bad {
    color: var(--danger);
  }

  .muted {
    color: var(--ink-mute);
  }
</style>
