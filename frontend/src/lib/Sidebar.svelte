<script lang="ts">
  import { Refs } from '../../wailsjs/go/app/App.js'
  import { EventsOn } from '../../wailsjs/runtime/runtime.js'
  import { onMount } from 'svelte'
  import type { DialogSpec } from './Dialog.svelte'
  import type { Ref } from './graph/rowCache'
  import { provideCommands, type Command } from './commands.svelte'
  import RepoExtras from './RepoExtras.svelte'
  import { session } from './session.svelte'
  import Icon from './ui/Icon.svelte'

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

  provideCommands(() => {
    const list: Command[] = []
    for (const r of refs) {
      if (r.kind === 'stash') continue
      if (r.kind === 'local' && !r.head && onActivate)
        list.push({ id: `checkout.${r.name}`, group: 'Branch', title: `Check out ${r.short}`, keywords: 'switch branch', run: () => onActivate(r) })
      const what = r.kind === 'tag' ? 'tag' : r.kind === 'remote' ? 'remote branch' : 'branch'
      list.push({ id: `goto.${r.name}`, group: 'Go to', title: `Go to ${r.short}`, keywords: `${what} show reveal graph`, run: () => onPick(r) })
    }
    return list
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
  <header>
    <div class="repo-name">{session.activeTab?.name ?? ''}</div>
    <label class="filter">
      <Icon name="search" size={14} />
      <input type="search" placeholder="Filter branches and tags" aria-label="Filter branches and tags" bind:value={filter} />
    </label>
  </header>
  <div class="groups">
    {#each groups as g (g.kind)}
      {@const items = byKind[g.kind]}
      <section>
        <button class="group" onclick={() => (open[g.kind] = !open[g.kind])} aria-expanded={open[g.kind] || !!filter}>
          <span class="caret" class:open={open[g.kind] || filter}><Icon name="right" size={12} /></span>
          <span class="meta title">{g.title}</span>
          <span class="meta count">{items.length.toLocaleString()}</span>
        </button>
        {#if g.kind === 'local' && onNewBranch}
          <button class="add" onclick={onNewBranch} title="New branch from HEAD" aria-label="New branch"><Icon name="plus" size={14} /></button>
        {:else if g.kind === 'remote' && onManageRemotes}
          <button class="add" onclick={onManageRemotes} title="Add, edit or remove remotes" aria-label="Manage remotes"><Icon name="gear" size={14} /></button>
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
                  {#if r.head}<span class="here" aria-label="checked out"></span>{/if}
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
    width: 248px;
    min-width: 180px;
    border-right: 1px solid var(--paper-edge);
    background: var(--paper-warm);
  }

  header {
    padding: 14px 12px 8px;
  }

  .repo-name {
    padding: 0 4px 8px;
    font-size: 19px;
    font-weight: 500;
    letter-spacing: -0.015em;
    line-height: 1.2;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-variation-settings:
      'SOFT' 50,
      'WONK' 1;
  }

  .filter {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 8px;
    background: var(--paper);
    border: 1px solid var(--paper-edge);
    border-radius: 4px;
    color: var(--ink-mute);
  }

  .filter:focus-within {
    outline: 2px solid var(--link);
    outline-offset: 1px;
  }

  .filter input {
    flex: 1;
    min-width: 0;
    height: 28px;
    padding: 0;
    background: none;
    border: none;
    outline: none;
    font-size: 13px;
  }

  .groups {
    flex: 1;
    overflow-y: auto;
    padding-bottom: 12px;
  }

  button {
    display: flex;
    gap: 8px;
    align-items: center;
    width: 100%;
    background: none;
    border: none;
    border-radius: 0;
    text-align: left;
    padding: 4px 12px;
  }

  button:hover:not(:disabled) {
    background: var(--wash-hover);
  }

  section {
    position: relative;
  }

  .group {
    padding: 10px 12px 4px;
  }

  .title {
    color: var(--ink-soft);
  }

  .add {
    position: absolute;
    right: 46px;
    top: 6px;
    width: auto;
    padding: 3px 5px;
    border-radius: 4px;
    color: var(--ink-mute);
  }

  .add:hover:not(:disabled) {
    color: var(--ink);
  }

  .caret {
    display: grid;
    place-items: center;
    width: 12px;
    color: var(--ink-mute);
    transition: transform 0.2s var(--ease-ink);
  }

  .caret.open {
    transform: rotate(90deg);
  }

  .count,
  .track {
    margin-left: auto;
  }

  .track {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--link);
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0 0 4px;
  }

  .ref {
    min-height: 28px;
    padding-left: 32px;
    font-size: 14px;
    color: var(--ink-soft);
    transition:
      background-color 0.2s,
      transform 0.3s var(--ease-ink);
  }

  .ref:hover:not(:disabled) {
    color: var(--ink);
    transform: translateX(3px);
  }

  .ref.head {
    color: var(--ink);
    font-weight: 600;
  }

  .here {
    flex: none;
    width: 7px;
    height: 7px;
    margin-left: -15px;
    border-radius: 50%;
    background: var(--link);
  }

  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .more {
    padding: 4px 32px;
    color: var(--ink-mute);
    font-size: 12px;
    font-style: italic;
  }

  @media (prefers-reduced-motion: reduce) {
    .caret,
    .ref {
      transition: none;
    }

    .ref:hover:not(:disabled) {
      transform: none;
    }
  }
</style>
