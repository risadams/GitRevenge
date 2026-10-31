<script lang="ts">
  import { onMount } from 'svelte'
  import { ChooseFolder, GroupAction, Groups, Init } from '../../wailsjs/go/app/App.js'
  import type { settings } from '../../wailsjs/go/models'
  import { commands } from './commands.svelte'
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte'
  import Dialog, { type DialogSpec } from './Dialog.svelte'
  import { errorText } from './errors'
  import GroupDashboard from './GroupDashboard.svelte'
  import { session } from './session.svelte'
  import Icon from './ui/Icon.svelte'
  import InkButton from './ui/InkButton.svelte'

  let groups = $state<settings.Group[]>([])
  let dialog = $state<DialogSpec | null>(null)
  let menu = $state<{ x: number; y: number; items: MenuItem[] } | null>(null)
  /** The group whose dashboard is showing, or ''. */
  let dashboard = $state('')

  const base = (p: string) => p.split(/[\\/]/).pop() ?? p

  async function load() {
    groups = (await session.run(() => Groups())) ?? []
  }

  /** Runs a group action from a dialog; an error keeps the dialog open. */
  async function act(action: string, ...args: string[]): Promise<string | void> {
    try {
      await GroupAction(action, args)
      await load()
    } catch (e) {
      return errorText(e)
    }
  }

  function newGroup() {
    dialog = {
      title: 'New group',
      fields: [{ key: 'name', label: 'Name', placeholder: 'e.g. Client project' }],
      submit: 'Create',
      run: (v) => act('create', String(v.name)),
    }
  }

  function renameGroup(g: settings.Group) {
    dialog = {
      title: `Rename ${g.name}`,
      fields: [{ key: 'name', label: 'Name', value: g.name }],
      submit: 'Rename',
      run: (v) => act('rename', g.name, String(v.name)),
    }
  }

  function deleteGroup(g: settings.Group) {
    dialog = {
      title: `Delete group ${g.name}?`,
      message: 'Only the group is removed. The repositories and their files are untouched.',
      submit: 'Delete group',
      danger: true,
      run: () => act('delete', g.name),
    }
  }

  function addRepo(g: settings.Group) {
    const candidates = session.recent.filter((r) => r.exists && !g.paths.includes(r.path))
    dialog = {
      title: `Add a repository to ${g.name}`,
      message: candidates.length ? undefined : 'No recent repositories to pick from. Browse for one instead.',
      fields: candidates.length
        ? [
            {
              key: 'path',
              label: 'Recent repository',
              type: 'select',
              value: candidates[0].path,
              options: candidates.map((r) => ({ value: r.path, label: `${r.name} (${r.path})` })),
            },
          ]
        : [],
      submit: candidates.length ? 'Add' : 'Browse…',
      run: async (v) => {
        let path = String(v.path ?? '')
        if (!path) {
          path = await ChooseFolder(`Add a repository to ${g.name}`)
          if (!path) return 'No folder chosen.'
        }
        return act('add', g.name, path)
      },
    }
  }

  async function browseInto(g: settings.Group) {
    const path = await session.run(() => ChooseFolder(`Add a repository to ${g.name}`))
    if (path) {
      const err = await act('add', g.name, path)
      if (err) session.error = err
    }
  }

  async function newRepo() {
    const dir = await session.run(() => ChooseFolder('Choose or create a folder for the new repository'))
    if (!dir) return
    dialog = {
      title: 'New repository',
      message: `Create an empty repository in ${dir}`,
      fields: [{ key: 'branch', label: 'Initial branch', placeholder: 'Your git default (init.defaultBranch)' }],
      submit: 'Create',
      run: async (v) => {
        try {
          await session.open(await Init(dir, String(v.branch)))
        } catch (e) {
          return errorText(e)
        }
      },
    }
  }

  async function openAll(g: settings.Group) {
    for (const p of g.paths) await session.open(p)
  }

  function more(e: MouseEvent, g: settings.Group) {
    const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
    menu = {
      x: r.left,
      y: r.bottom + 2,
      items: [
        { label: 'Browse for a repository…', icon: 'folder-open', run: () => browseInto(g) },
        { label: 'Rename…', icon: 'edit', run: () => renameGroup(g) },
        { label: 'Delete group', icon: 'trash', danger: true, run: () => deleteGroup(g) },
      ],
    }
  }

  const DAY = 24 * 60 * 60 * 1000
  const dayFmt = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' })

  /** "Today", "Yesterday", or a short date. */
  function when(t: string): string {
    const d = new Date(t)
    if (isNaN(d.getTime())) return ''
    const now = new Date()
    const midnight = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()
    if (d.getTime() >= midnight) return 'Today'
    if (d.getTime() >= midnight - DAY) return 'Yesterday'
    return dayFmt.format(d)
  }

  /** Group cards cycle through the inks for their splash. */
  const TONES = ['var(--ink-rose)', 'var(--ink-teal-soft)', 'var(--ink-plum)', 'var(--ink-sepia)']

  onMount(load)
</script>

{#if dashboard}
  <GroupDashboard name={dashboard} onClose={() => (dashboard = '')} />
{:else}
  <div class="page">
    <svg class="grain" aria-hidden="true" focusable="false"><rect width="100%" height="100%" filter="url(#paper-grain)" /></svg>
    <span class="blob rose" aria-hidden="true"></span>
    <span class="blob teal" aria-hidden="true"></span>

    <div class="start">
      <header class="hero">
        <span class="hand kicker">{session.recent.length ? 'welcome back' : 'hello there'}</span>
        <h1>
          {#if session.recent.length}Pick up where you <em>left off.</em>{:else}Bring a repository <em>along.</em>{/if}
        </h1>
        <p class="lede">Open a repository, clone one, or start fresh.{session.recent.length ? ' Your recent work is below.' : ''}</p>
        <div class="actions">
          <InkButton size="lg" arrow onclick={() => session.choose()}>Open repository</InkButton>
          <InkButton variant="secondary" size="lg" onclick={() => (commands.cloneOpen = true)}>Clone a repository</InkButton>
          <InkButton variant="secondary" size="lg" onclick={newRepo}>New repository</InkButton>
        </div>
      </header>

      <div class="columns">
        <section class="recent" aria-labelledby="recent-h">
          <h2 id="recent-h">Recent</h2>
          {#if session.recent.length === 0}
            <p class="empty"><Icon name="folder-open" size={28} wash="teal" />Repositories you open show up here.</p>
          {:else}
            <ul>
              {#each session.recent as r (r.path)}
                <li class:missing={!r.exists}>
                  <button class="row" disabled={!r.exists} onclick={() => session.open(r.path)} title={r.exists ? r.path : `${r.path} (not found)`}>
                    <span class="meta">{r.exists ? when(r.lastOpened) : 'Missing'}</span>
                    <span class="name">{r.name}</span>
                    <span class="path">{r.path}</span>
                  </button>
                  <button class="forget" aria-label="Remove {r.name} from recents" title="Remove from recents" onclick={() => session.forget(r.path)}>
                    <Icon name="close" size={13} />
                  </button>
                </li>
              {/each}
            </ul>
          {/if}
        </section>

        <section class="groups" aria-labelledby="groups-h">
          <div class="groups-head">
            <h2 id="groups-h">Groups</h2>
            <button class="tool" onclick={newGroup}><Icon name="plus" />New group</button>
          </div>
          {#if groups.length === 0}
            <p class="empty">
              <Icon name="worktree" size={28} wash="rose" />Group related repositories to open them together, see their status side by side, and fetch or pull them in one go.
            </p>
          {/if}
          <div class="cards">
            {#each groups as g, i (g.name)}
              <article class="card" style:--splash={TONES[i % TONES.length]}>
                <span class="splash" aria-hidden="true"></span>
                <div class="card-body">
                  <span class="num" aria-hidden="true">{String(i + 1).padStart(2, '0')}</span>
                  <h3>{g.name}</h3>
                  <ul class="repos">
                    {#each g.paths as p (p)}
                      <li>
                        <button class="repo" onclick={() => session.open(p)} title={p}>{base(p)}</button>
                        <button class="forget" aria-label="Remove {base(p)} from {g.name}" title="Remove from group" onclick={() => act('remove', g.name, p)}>
                          <Icon name="close" size={12} />
                        </button>
                      </li>
                    {:else}
                      <li class="none">No repositories yet.</li>
                    {/each}
                  </ul>
                  <div class="card-actions">
                    <button class="link" onclick={() => (dashboard = g.name)}>Dashboard</button>
                    <button class="link" disabled={!g.paths.length} onclick={() => openAll(g)}>Open all</button>
                    <button class="link" onclick={() => addRepo(g)}>Add…</button>
                    <button class="tool more" aria-label="More actions for {g.name}" title="More" onclick={(e) => more(e, g)}><Icon name="more" /></button>
                  </div>
                </div>
              </article>
            {/each}
          </div>
        </section>
      </div>
    </div>
  </div>
{/if}

{#if menu}
  <ContextMenu x={menu.x} y={menu.y} items={menu.items} onClose={() => (menu = null)} />
{/if}

{#if dialog}
  <Dialog spec={dialog} onClose={() => (dialog = null)} />
{/if}

<style>
  .page {
    position: relative;
    min-height: 100%;
    overflow: hidden;
  }

  /* Paper grain and watercolour blobs: decoration only. */
  .grain {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    opacity: 0.06;
    mix-blend-mode: multiply;
  }

  :global([data-theme='dark']) .grain {
    opacity: 0.09;
    mix-blend-mode: screen;
  }

  .blob {
    position: absolute;
    pointer-events: none;
    filter: blur(70px);
    border-radius: 45% 55% 50% 50% / 55% 45%;
    opacity: 0.2;
  }

  .blob.rose {
    top: 30px;
    right: 6%;
    width: 420px;
    height: 300px;
    background: var(--ink-rose);
  }

  .blob.teal {
    top: 200px;
    right: 24%;
    width: 260px;
    height: 220px;
    background: var(--ink-teal-soft);
  }

  :global([data-theme='dark']) .blob {
    opacity: 0.14;
  }

  .start {
    position: relative;
    max-width: 1180px;
    margin: 0 auto;
    padding: 56px 40px 48px;
  }

  .hero {
    max-width: 760px;
  }

  .kicker {
    font-size: 30px;
    color: var(--link);
  }

  h1 {
    margin: 8px 0 0;
    font-size: clamp(40px, 5.5vw, 68px);
    font-weight: 400;
    line-height: 1;
    letter-spacing: -0.03em;
    font-variation-settings:
      'SOFT' 50,
      'WONK' 1;
  }

  h1 em {
    color: var(--link);
    font-weight: 500;
  }

  .lede {
    margin: 16px 0 0;
    max-width: 56ch;
    font-size: 18px;
    line-height: 1.55;
    color: var(--ink-soft);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 26px;
  }

  .columns {
    display: flex;
    flex-wrap: wrap;
    gap: 40px 64px;
    margin-top: 56px;
  }

  .recent {
    flex: 1 1 420px;
    min-width: 0;
  }

  .groups {
    flex: 1 1 460px;
    min-width: 0;
  }

  h2 {
    margin: 0;
    font-size: 30px;
    font-weight: 400;
    letter-spacing: -0.02em;
  }

  .empty {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    color: var(--ink-mute);
    font-style: italic;
    line-height: 1.6;
  }

  .empty :global(svg) {
    color: var(--ink-soft);
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .recent ul {
    margin-top: 10px;
  }

  .recent li {
    display: flex;
    align-items: center;
    border-bottom: 1px dashed var(--hairline);
  }

  .row {
    flex: 1;
    min-width: 0;
    display: grid;
    grid-template-columns: 84px minmax(0, auto) minmax(0, 1fr);
    align-items: baseline;
    gap: 16px;
    padding: 14px 0;
    background: none;
    border: none;
    border-radius: 0;
    text-align: left;
    transition: transform 0.3s var(--ease-ink);
  }

  .row:hover:not(:disabled) {
    background: none;
    transform: translateX(6px);
  }

  .row:hover:not(:disabled) .name {
    color: var(--link);
  }

  .row:disabled {
    opacity: 1;
  }

  .name {
    font-size: 20px;
    font-weight: 500;
    letter-spacing: -0.01em;
    white-space: nowrap;
    transition: color 0.3s;
  }

  .path {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--ink-faint);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  li.missing .name {
    text-decoration: line-through;
    color: var(--ink-mute);
  }

  .forget {
    display: grid;
    place-items: center;
    width: 26px;
    height: 26px;
    padding: 0;
    background: none;
    border: none;
    border-radius: 50%;
    color: var(--ink-mute);
    opacity: 0;
    transition: opacity 0.2s;
  }

  li:hover > .forget,
  .forget:focus-visible {
    opacity: 1;
  }

  .forget:hover:not(:disabled) {
    background: var(--wash-hover);
    color: var(--ink);
  }

  .groups-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 8px 20px;
    margin-top: 6px;
  }

  .card {
    position: relative;
    padding: 24px 20px 20px;
  }

  .splash {
    position: absolute;
    top: 12px;
    left: 8px;
    width: 58px;
    height: 50px;
    background: var(--splash);
    opacity: 0.45;
    filter: url(#wc-edge);
    mix-blend-mode: multiply;
    border-radius: 62% 38% 46% 54% / 48% 56% 44% 52%;
    pointer-events: none;
  }

  :global([data-theme='dark']) .splash {
    mix-blend-mode: screen;
    opacity: 0.4;
  }

  .card-body {
    position: relative;
  }

  .num {
    font-family: var(--font-hand);
    font-size: 34px;
    line-height: 1;
    opacity: 0.85;
  }

  h3 {
    margin: 10px 0 8px;
    font-size: 22px;
    font-weight: 500;
    letter-spacing: -0.01em;
  }

  .repos li {
    display: flex;
    align-items: center;
  }

  .repo {
    position: relative;
    flex: 1;
    min-width: 0;
    padding: 3px 0 3px 18px;
    background: none;
    border: none;
    border-radius: 0;
    text-align: left;
    color: var(--ink-soft);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .repo::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    width: 9px;
    height: 6px;
    margin-top: -3px;
    background: var(--splash);
    opacity: 0.6;
    border-radius: 50%;
    filter: url(#wc-edge);
  }

  .repo:hover:not(:disabled) {
    background: none;
    color: var(--link);
  }

  .none {
    color: var(--ink-mute);
    font-style: italic;
    padding: 3px 0;
  }

  .card-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 2px 14px;
    margin-top: 14px;
  }

  .link {
    padding: 0;
    border: none;
    border-radius: 0;
    font-style: italic;
    color: var(--link);
    background: linear-gradient(var(--link), var(--link)) 0 100% / 100% 1px no-repeat;
    transition:
      background-size 0.3s,
      color 0.3s;
  }

  .link:hover:not(:disabled) {
    color: var(--link-hover);
    background: linear-gradient(var(--selection), var(--selection)) 0 100% / 100% 50% no-repeat;
  }

  /* Pinned to the card's corner so the action row never wraps around it. */
  .more {
    position: absolute;
    top: 2px;
    right: -6px;
    height: 28px;
    padding: 0 6px;
  }

  @media (prefers-reduced-motion: reduce) {
    .row,
    .link {
      transition: none;
    }

    .row:hover:not(:disabled) {
      transform: none;
    }
  }
</style>
