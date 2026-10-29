<script lang="ts">
  import { Refs } from '../../wailsjs/go/app/App.js'
  import { EventsOn } from '../../wailsjs/runtime/runtime.js'
  import { onMount } from 'svelte'
  import type { DialogSpec } from './Dialog.svelte'
  import type { Ref } from './graph/rowCache'
  import RepoExtras from './RepoExtras.svelte'
  import { session } from './session.svelte'

  let {
    path,
    onPick,
    onActivate,
    onNewBranch,
    onManageRemotes,
    onContext,
    openDialog,
  }: {
    path: string
    onPick: (ref: Ref) => void
    /** Double-click: check the ref out. */
    onActivate?: (ref: Ref) => void
    onNewBranch?: () => void
    onManageRemotes?: () => void
    onContext?: (ref: Ref, x: number, y: number) => void
    openDialog: (spec: DialogSpec) => void
  } = $props()

  /** Groups render at most this many items; the filter narrows the rest. Keeps the DOM
   * small in repositories with tens of thousands of refs. */
  const MAX_ITEMS = 500

  let refs = $state<Ref[]>([])
  let filter = $state('')
  let open = $state<Record<string, boolean>>({ local: true, remote: false, tag: false, stash: true })

  const groups: { kind: Ref['kind']; title: string }[] = [
    { kind: 'local', title: 'Branches' },
    { kind: 'remote', title: 'Remotes' },
    { kind: 'tag', title: 'Tags' },
    { kind: 'stash', title: 'Stashes' },
  ]

  let byKind = $derived.by(() => {
    const q = filter.trim().toLowerCase()
    const out: Record<string, Ref[]> = { local: [], remote: [], tag: [], stash: [] }
    for (const r of refs) {
      if (!q || r.short.toLowerCase().includes(q)) out[r.kind].push(r)
    }
    out.local.sort((a, b) => Number(b.head) - Number(a.head) || a.short.localeCompare(b.short, undefined, { numeric: true }))
    out.remote.sort((a, b) => a.short.localeCompare(b.short, undefined, { numeric: true }))
    // Newest-looking tags first (v2.0 before v1.9).
    out.tag.sort((a, b) => b.short.localeCompare(a.short, undefined, { numeric: true }))
    return out
  })

  async function load() {
    refs = ((await session.run(() => Refs(path))) as Ref[] | undefined) ?? []
  }

  onMount(() => {
    void load()
    // Refs are re-read with every graph reload.
    return EventsOn('graph:progress', (p: { path: string; done: boolean }) => {
      if (p.path === path && p.done) void load()
    })
  })
</script>

<aside class="sidebar">
  <input class="filter" type="search" placeholder="Filter branches and tags" bind:value={filter} />
  <div class="groups">
    {#each groups as g (g.kind)}
      {@const items = byKind[g.kind]}
      <section>
        <button class="group" onclick={() => (open[g.kind] = !open[g.kind])} aria-expanded={open[g.kind] || !!filter}>
          <span class="caret">{open[g.kind] || filter ? '▾' : '▸'}</span>
          {g.title}
          <span class="count">{items.length.toLocaleString()}</span>
        </button>
        {#if g.kind === 'local' && onNewBranch}
          <button class="add" onclick={onNewBranch} title="New branch from HEAD" aria-label="New branch">+</button>
        {:else if g.kind === 'remote' && onManageRemotes}
          <button class="add" onclick={onManageRemotes} title="Add, edit or remove remotes" aria-label="Manage remotes">⚙</button>
        {/if}
        {#if open[g.kind] || filter}
          <ul>
            {#each items.slice(0, MAX_ITEMS) as r (r.name)}
              <li>
                <button
                  class="ref"
                  class:head={r.head}
                  onclick={() => onPick(r)}
                  ondblclick={() => onActivate?.(r)}
                  oncontextmenu={(e) => {
                    if (!onContext) return
                    e.preventDefault()
                    onContext(r, e.clientX, e.clientY)
                  }}
                  title={r.subject ?? r.name}
                >
                  <span class="name">{r.kind === 'stash' ? r.subject : r.short}</span>
                  {#if r.gone}
                    <span class="track" title="Upstream branch was deleted">gone</span>
                  {:else if r.ahead || r.behind}
                    <span class="track" title="Ahead / behind {r.upstream}">
                      {#if r.ahead}↑{r.ahead}{/if}{#if r.behind}↓{r.behind}{/if}
                    </span>
                  {/if}
                </button>
              </li>
            {/each}
            {#if items.length > MAX_ITEMS}
              <li class="more">{(items.length - MAX_ITEMS).toLocaleString()} more. Type to filter.</li>
            {/if}
          </ul>
        {/if}
      </section>
    {/each}
    <RepoExtras {path} {openDialog} />
  </div>
</aside>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    width: 240px;
    min-width: 180px;
    border-right: 1px solid var(--paper-edge);
    background: var(--paper-warm);
  }

  .filter {
    margin: 8px;
    padding: 4px 6px;
    background: var(--paper);
    color: var(--ink);
    border: 1px solid var(--paper-edge);
  }

  .groups {
    flex: 1;
    overflow-y: auto;
  }

  button {
    display: flex;
    gap: 6px;
    align-items: center;
    width: 100%;
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
    padding: 3px 10px;
  }

  button:hover {
    background: var(--wash-hover);
  }

  section {
    position: relative;
  }

  .group {
    font-weight: 600;
    padding: 6px 8px;
  }

  .add {
    position: absolute;
    right: 44px;
    top: 3px;
    width: auto;
    padding: 2px 6px;
    color: var(--ink-mute);
  }

  .caret {
    width: 10px;
    color: var(--ink-mute);
  }

  .count,
  .track {
    margin-left: auto;
    color: var(--ink-mute);
    font-weight: normal;
    font-size: 11px;
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0 0 4px;
  }

  .ref {
    padding-left: 26px;
  }

  .ref.head {
    font-weight: 600;
  }

  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .more {
    padding: 3px 26px;
    color: var(--ink-mute);
    font-size: 12px;
  }
</style>
