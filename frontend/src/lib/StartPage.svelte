<script lang="ts">
  import { onMount } from 'svelte'
  import { ChooseFolder, GroupAction, Groups } from '../../wailsjs/go/app/App.js'
  import type { settings } from '../../wailsjs/go/models'
  import Dialog, { type DialogSpec } from './Dialog.svelte'
  import { errorText } from './errors'
  import GroupDashboard from './GroupDashboard.svelte'
  import { session } from './session.svelte'

  let groups = $state<settings.Group[]>([])
  let open = $state<Record<string, boolean>>({})
  let dialog = $state<DialogSpec | null>(null)
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
      run: async (v) => {
        const err = await act('create', String(v.name))
        if (!err) open[String(v.name).trim()] = true
        return err
      },
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

  async function openAll(g: settings.Group) {
    for (const p of g.paths) await session.open(p)
  }

  onMount(load)
</script>

{#if dashboard}
  <GroupDashboard name={dashboard} onClose={() => (dashboard = '')} />
{:else}
  <section class="start">
    <h2>Open a repository</h2>
    <button onclick={() => session.choose()}>Open repository…</button>

    <div class="heading">
      <h3>Groups</h3>
      <button class="small" onclick={newGroup}>New group…</button>
    </div>
    {#if groups.length === 0}
      <p class="muted">Group related repositories to open them together, see their status side by side, and fetch or pull them in one go.</p>
    {/if}
    {#each groups as g (g.name)}
      <div class="group">
        <div class="group-head">
          <button class="caret" onclick={() => (open[g.name] = !open[g.name])} aria-expanded={!!open[g.name]}>
            {open[g.name] ? '▾' : '▸'} <strong>{g.name}</strong> <span class="muted">{g.paths.length}</span>
          </button>
          <span class="actions">
            <button class="small" disabled={!g.paths.length} onclick={() => openAll(g)}>Open all</button>
            <button class="small" onclick={() => (dashboard = g.name)}>Dashboard</button>
            <button class="small" onclick={() => addRepo(g)}>Add…</button>
            <button class="small" onclick={() => browseInto(g)} title="Add a repository by browsing">Browse…</button>
            <button class="small" onclick={() => renameGroup(g)}>Rename</button>
            <button class="small" onclick={() => deleteGroup(g)}>Delete</button>
          </span>
        </div>
        {#if open[g.name]}
          <ul>
            {#each g.paths as p (p)}
              <li>
                <button class="link" onclick={() => session.open(p)} title={p}>
                  <strong>{base(p)}</strong>
                  <span class="muted">{p}</span>
                </button>
                <button class="forget" aria-label="Remove {base(p)} from {g.name}" onclick={() => act('remove', g.name, p)}>×</button>
              </li>
            {:else}
              <li class="muted empty">No repositories yet.</li>
            {/each}
          </ul>
        {/if}
      </div>
    {/each}

    {#if session.recent.length > 0}
      <h3>Recent</h3>
      <ul>
        {#each session.recent as r (r.path)}
          <li class:missing={!r.exists}>
            <button class="link" disabled={!r.exists} onclick={() => session.open(r.path)} title={r.path}>
              <strong>{r.name}</strong>
              <span class="muted">{r.path}</span>
            </button>
            <button class="forget" aria-label="Remove {r.name} from recents" onclick={() => session.forget(r.path)}>×</button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>
{/if}

{#if dialog}
  <Dialog spec={dialog} onClose={() => (dialog = null)} />
{/if}

<style>
  .start {
    max-width: 720px;
    margin: 48px auto;
    padding: 0 16px;
  }

  .heading {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 24px;
  }

  .heading h3 {
    margin: 0;
  }

  .group {
    border-bottom: 1px solid var(--border);
    padding: 4px 0;
  }

  .group-head {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .caret {
    flex: 1;
    text-align: left;
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    cursor: pointer;
    padding: 4px 0;
  }

  .actions {
    display: flex;
    gap: 4px;
  }

  .small {
    padding: 1px 8px;
    background: none;
    border: 1px solid var(--border);
    color: var(--fg);
    cursor: pointer;
    font-size: 12px;
  }

  .small:disabled {
    opacity: 0.5;
  }

  ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .group ul {
    padding-left: 16px;
  }

  li {
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--border);
  }

  .group li:last-child {
    border-bottom: none;
  }

  li.missing {
    opacity: 0.5;
  }

  li.empty {
    padding: 6px 0;
  }

  .link {
    flex: 1;
    display: flex;
    gap: 12px;
    align-items: baseline;
    text-align: left;
    background: none;
    border: none;
    color: inherit;
    padding: 8px 0;
    cursor: pointer;
    min-width: 0;
  }

  .link .muted {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .forget {
    background: none;
    border: none;
    color: var(--fg-muted);
    cursor: pointer;
  }

  .muted {
    color: var(--fg-muted);
  }
</style>
