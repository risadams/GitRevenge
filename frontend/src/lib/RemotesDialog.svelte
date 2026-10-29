<script lang="ts">
  import { onMount } from 'svelte'
  import { RemoteAction, Remotes } from '../../wailsjs/go/app/App.js'
  import type { git } from '../../wailsjs/go/models'
  import { errorText } from './errors'

  let { path, onClose }: { path: string; onClose: () => void } = $props()

  let remotes = $state<git.Remote[]>([])
  let error = $state('')
  /** Row being edited ("" for a new remote), or null. */
  let editing = $state<string | null>(null)
  let name = $state('')
  let url = $state('')
  let confirmRemove = $state<string | null>(null)

  async function load() {
    try {
      remotes = (await Remotes(path)) ?? []
    } catch (e) {
      error = errorText(e)
    }
  }

  async function act(action: string, args: string[]) {
    error = ''
    try {
      await RemoteAction(path, action, args)
      return true
    } catch (e) {
      error = errorText(e)
      return false
    }
  }

  function startEdit(r?: git.Remote) {
    editing = r?.name ?? ''
    name = r?.name ?? (remotes.length ? '' : 'origin')
    url = r?.fetchUrl ?? ''
    error = ''
  }

  async function save(e: SubmitEvent) {
    e.preventDefault()
    let ok = true
    if (editing === '') ok = await act('add', [name, url])
    else if (editing !== null) {
      const old = remotes.find((r) => r.name === editing)
      if (old && name !== old.name) ok = await act('rename', [old.name, name])
      if (ok && old && url !== old.fetchUrl) ok = await act('seturl', [name, url])
    }
    if (ok) {
      editing = null
      await load()
    }
  }

  async function remove(n: string) {
    confirmRemove = null
    if (await act('remove', [n])) await load()
  }

  onMount(load)
</script>

<svelte:window onkeydown={(e) => e.key === 'Escape' && onClose()} />

<div class="backdrop" role="presentation" onmousedown={(e) => e.target === e.currentTarget && onClose()}>
  <div class="dialog" role="dialog" aria-modal="true" aria-label="Remotes">
    <h3>Remotes</h3>
    {#if remotes.length === 0 && editing === null}<p class="muted">No remotes yet.</p>{/if}
    <ul>
      {#each remotes as r (r.name)}
        <li>
          <div class="info">
            <strong>{r.name}</strong>
            <span class="url" title={r.fetchUrl}>{r.fetchUrl}</span>
            {#if r.pushUrl && r.pushUrl !== r.fetchUrl}<span class="url" title={r.pushUrl}>push: {r.pushUrl}</span>{/if}
          </div>
          {#if confirmRemove === r.name}
            <button class="danger" onclick={() => remove(r.name)}>Remove {r.name}?</button>
            <button onclick={() => (confirmRemove = null)}>Keep</button>
          {:else}
            <button onclick={() => startEdit(r)}>Edit</button>
            <button onclick={() => (confirmRemove = r.name)}>Remove</button>
          {/if}
        </li>
      {/each}
    </ul>

    {#if editing !== null}
      <form onsubmit={save}>
        <label><span>Name</span><input bind:value={name} placeholder="origin" /></label>
        <label><span>URL</span><input bind:value={url} placeholder="git@github.com:you/repo.git" /></label>
        <div class="buttons">
          <button type="button" onclick={() => (editing = null)}>Cancel</button>
          <button type="submit" class="primary">{editing === '' ? 'Add remote' : 'Save'}</button>
        </div>
      </form>
    {:else}
      <div class="buttons">
        <button onclick={() => startEdit()}>Add remote…</button>
        <button class="primary" onclick={onClose}>Done</button>
      </div>
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
  </div>
</div>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 200;
    display: grid;
    place-items: center;
    background: rgba(0, 0, 0, 0.45);
  }

  .dialog {
    width: min(560px, calc(100vw - 32px));
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 16px;
    background: var(--bg-raised);
    border: 1px solid var(--border);
  }

  h3 {
    margin: 0;
    font-size: 14px;
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  li {
    display: flex;
    gap: 6px;
    align-items: center;
    padding: 6px 0;
    border-bottom: 1px solid var(--border);
  }

  .info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .url {
    color: var(--fg-muted);
    font-family: var(--font-mono);
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  input {
    padding: 4px 6px;
    background: var(--bg);
    color: var(--fg);
    border: 1px solid var(--border);
    font: inherit;
  }

  .buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  button {
    padding: 3px 10px;
    background: none;
    border: 1px solid var(--border);
    color: var(--fg);
    cursor: pointer;
  }

  .primary {
    background: var(--bg-selected);
    border-color: var(--bg-selected);
  }

  .danger {
    border-color: #a33;
    color: #f48771;
  }

  .muted {
    color: var(--fg-muted);
    margin: 0;
  }

  .error {
    color: #f48771;
    margin: 0;
    white-space: pre-wrap;
  }
</style>
