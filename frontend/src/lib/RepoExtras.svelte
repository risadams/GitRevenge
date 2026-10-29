<script lang="ts">
  import { LFS, LFSAction, LFSLocks, Submodules, UpdateSubmodules } from '../../wailsjs/go/app/App.js'
  import type { app, git } from '../../wailsjs/go/models'
  import type { DialogSpec } from './Dialog.svelte'
  import { errorText } from './errors'
  import { session } from './session.svelte'

  let { path, openDialog }: { path: string; openDialog: (spec: DialogSpec) => void } = $props()

  let subs = $state<git.Submodule[]>([])
  let lfs = $state<app.LFSInfo | null>(null)
  /** Locks from the server, by file path; null until checked (it needs the network). */
  let locks = $state<Record<string, git.LFSLock> | null>(null)
  let open = $state<Record<string, boolean>>({ subs: true, lfs: false, files: false })
  let busy = $state(false)

  const stateLabel: Record<string, string> = {
    ok: '',
    uninitialized: 'not initialized',
    modified: 'different commit',
    conflict: 'conflict',
  }

  async function load() {
    const [s, l] = await Promise.all([session.run(() => Submodules(path)), session.run(() => LFS(path))])
    subs = s ?? []
    lfs = l ?? null
  }

  $effect(() => {
    void session.revision[path]
    void load()
  })

  async function run(fn: () => Promise<unknown>) {
    busy = true
    await session.run(fn)
    busy = false
    await load()
  }

  async function checkLocks() {
    busy = true
    const list = await session.run(() => LFSLocks(path))
    busy = false
    if (list) locks = Object.fromEntries(list.map((l) => [l.path, l]))
  }

  async function lockAction(file: string, action: 'lock' | 'unlock', force = false) {
    await run(() => LFSAction(path, action, [file, force ? 'true' : '']))
    await checkLocks()
  }

  function track() {
    openDialog({
      title: 'Track files with LFS',
      message: 'Matching files are stored in Git LFS from now on. This edits .gitattributes, which you then commit.',
      fields: [
        { key: 'pattern', label: 'Pattern', placeholder: '*.psd' },
        { key: 'lockable', label: 'Lockable (read-only until locked)', type: 'checkbox' },
      ],
      submit: 'Track',
      run: async (v) => {
        try {
          await LFSAction(path, 'track', [String(v.pattern).trim(), v.lockable ? 'true' : ''])
          await load()
        } catch (e) {
          return errorText(e)
        }
      },
    })
  }

  const join = (base: string, rel: string) => base.replace(/[\\/]+$/, '') + '/' + rel
</script>

{#if subs.length}
  <section>
    <div class="head">
      <button class="group" onclick={() => (open.subs = !open.subs)}>
        <span class="caret">{open.subs ? '▾' : '▸'}</span> Submodules <span class="count">{subs.length}</span>
      </button>
      <button class="tool" disabled={busy} onclick={() => run(() => UpdateSubmodules(path))} title="Initialize and update all submodules"
        >Update</button
      >
    </div>
    {#if open.subs}
      <ul>
        {#each subs as s (s.path)}
          <li>
            <button class="item" onclick={() => session.open(join(path, s.path))} title="Open {s.path} as a tab ({s.commit.slice(0, 7)})">
              <span class="name">{s.path}</span>
              {#if stateLabel[s.state]}<span class="badge s-{s.state}">{stateLabel[s.state]}</span>{/if}
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>
{/if}

{#if lfs?.available}
  <section>
    <div class="head">
      <button class="group" onclick={() => (open.lfs = !open.lfs)}>
        <span class="caret">{open.lfs ? '▾' : '▸'}</span> LFS <span class="count">{lfs.status.files.length}</span>
      </button>
      <button class="tool" onclick={track} title="Track a pattern with LFS">+</button>
    </div>
    {#if open.lfs}
      <ul>
        {#each lfs.status.patterns as p (p.pattern)}
          <li class="row">
            <span class="name mono" title="from {p.source}">{p.pattern}</span>
            {#if p.lockable}<span class="badge">lockable</span>{/if}
            <button class="x" disabled={busy} onclick={() => run(() => LFSAction(path, 'untrack', [p.pattern]))} aria-label="Stop tracking {p.pattern}"
              >×</button
            >
          </li>
        {:else}
          <li class="muted pad">No patterns tracked.</li>
        {/each}
      </ul>
      {#if lfs.status.files.length}
        <div class="head">
          <button class="group sub" onclick={() => (open.files = !open.files)}>
            <span class="caret">{open.files ? '▾' : '▸'}</span> Files <span class="count">{lfs.status.files.length}</span>
          </button>
          <button class="tool" disabled={busy} onclick={checkLocks} title="Ask the LFS server which files are locked">Check locks</button>
        </div>
        {#if open.files}
          <ul>
            {#each lfs.status.files as f (f.name)}
              {@const lock = locks?.[f.name]}
              <li class="row">
                <span class="name" title={f.name}>{f.name}</span>
                {#if !f.downloaded}<span class="badge" title="Only the pointer is here; content not downloaded">pointer</span>{/if}
                {#if lock}
                  <span class="badge" class:mine={lock.ours} title="Locked {lock.lockedAt}">🔒 {lock.ours ? 'you' : lock.owner}</span>
                  <button class="tool" disabled={busy} onclick={() => lockAction(f.name, 'unlock', !lock.ours)}
                    >{lock.ours ? 'Unlock' : 'Force unlock'}</button
                  >
                {:else if locks}
                  <button class="tool" disabled={busy} onclick={() => lockAction(f.name, 'lock')}>Lock</button>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      {/if}
    {/if}
  </section>
{/if}

<style>
  section {
    border-top: 1px solid var(--paper-edge);
  }

  .head {
    display: flex;
    align-items: center;
    padding-right: 8px;
  }

  button {
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    cursor: pointer;
  }

  .group {
    flex: 1;
    text-align: left;
    font-weight: 600;
    padding: 6px 8px;
  }

  .group.sub {
    font-weight: normal;
    padding-left: 22px;
  }

  .caret {
    color: var(--ink-mute);
  }

  .count {
    color: var(--ink-mute);
    font-weight: normal;
    font-size: 11px;
  }

  .tool {
    padding: 0 6px;
    border: 1px solid var(--paper-edge);
    font-size: 11px;
    color: var(--ink-mute);
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0 0 4px;
  }

  .item,
  .row {
    display: flex;
    gap: 6px;
    align-items: center;
    width: 100%;
    padding: 3px 10px 3px 26px;
    text-align: left;
  }

  .item:hover {
    background: var(--wash-hover);
  }

  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .mono {
    font-family: var(--font-mono);
  }

  .badge {
    font-size: 10px;
    padding: 0 4px;
    border: 1px solid var(--paper-edge);
    color: var(--ink-mute);
    white-space: nowrap;
  }

  .badge.mine {
    color: var(--success);
  }

  .s-uninitialized,
  .s-modified {
    color: var(--accent-c);
  }

  .s-conflict {
    color: var(--danger);
  }

  .x {
    color: var(--ink-mute);
  }

  .muted {
    color: var(--ink-mute);
  }

  .pad {
    padding: 3px 26px;
  }
</style>
