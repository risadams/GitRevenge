<script lang="ts">
  import { onMount } from 'svelte'
  import { CancelClone, ChooseFolder, Clone, CloneName } from '../../wailsjs/go/app/App.js'
  import { EventsOn } from '../../wailsjs/runtime/runtime.js'
  import { errorText } from './errors'
  import { session } from './session.svelte'

  let { onClose }: { onClose: () => void } = $props()

  const dirOf = (p: string) => p.replace(/[\\/][^\\/]*$/, '')

  let url = $state('')
  // Default to the folder holding the most recently opened repository.
  let parent = $state(session.recent[0] ? dirOf(session.recent[0].path) : '')
  let name = $state('')
  /** True once the user types a name, so it stops following the URL. */
  let nameEdited = $state(false)
  let running = $state(false)
  let progress = $state('')
  let error = $state('')
  let input: HTMLInputElement

  async function onUrl() {
    if (nameEdited) return
    const u = url
    const n = await CloneName(u)
    if (u === url && !nameEdited) name = n
  }

  async function browse() {
    const dir = await session.run(() => ChooseFolder('Clone into folder'))
    if (dir) parent = dir
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    running = true
    progress = ''
    error = ''
    try {
      const dir = await Clone(url, parent, name)
      await session.open(dir)
      onClose()
    } catch (err) {
      error = errorText(err)
    } finally {
      running = false
    }
  }

  /** Cancel stops a running clone (its error then shows here); otherwise it closes. */
  function cancel() {
    if (running) CancelClone()
    else onClose()
  }

  onMount(() => {
    input.focus()
    return EventsOn('net:progress', (p: { op: string; phase: string; percent: number; line: string }) => {
      if (p.op !== 'clone') return
      progress = p.phase ? `${p.phase} ${p.percent}%` : p.line
    })
  })
</script>

<svelte:window onkeydown={(e) => e.key === 'Escape' && cancel()} />

<div class="backdrop" role="presentation" onmousedown={(e) => e.target === e.currentTarget && !running && onClose()}>
  <div class="dialog" role="dialog" aria-modal="true" aria-label="Clone a repository">
    <form onsubmit={submit}>
      <h3>Clone a repository</h3>
      <label>
        <span>Repository URL</span>
        <input
          bind:this={input}
          bind:value={url}
          oninput={onUrl}
          disabled={running}
          placeholder="https://… or git@host:owner/repo.git"
        />
      </label>
      <label>
        <span>Into folder</span>
        <span class="row">
          <input bind:value={parent} disabled={running} placeholder="Choose a folder" />
          <button type="button" disabled={running} onclick={browse}>Browse…</button>
        </span>
      </label>
      <label>
        <span>Folder name</span>
        <input bind:value={name} oninput={() => (nameEdited = true)} disabled={running} />
      </label>
      {#if running}
        <p class="status" role="status"><span class="spinner"></span> {progress || 'Connecting…'}</p>
      {/if}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <div class="buttons">
        <button type="button" onclick={cancel}>Cancel</button>
        <button type="submit" class="primary" disabled={running || !url.trim() || !parent.trim() || !name.trim()}
          >Clone</button
        >
      </div>
    </form>
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
    width: min(520px, calc(100vw - 32px));
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 16px;
    background: var(--bg-raised);
    border: 1px solid var(--border);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  }

  h3 {
    margin: 0;
    font-size: 14px;
  }

  p {
    margin: 0;
    color: var(--fg-muted);
    white-space: pre-wrap;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .row {
    display: flex;
    gap: 6px;
  }

  .row input {
    flex: 1;
    min-width: 0;
  }

  input {
    padding: 4px 6px;
    background: var(--bg);
    color: var(--fg);
    border: 1px solid var(--border);
    font: inherit;
  }

  .status {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .spinner {
    flex: none;
    width: 10px;
    height: 10px;
    border: 2px solid var(--fg-muted);
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error {
    color: #f48771;
  }

  .buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  button {
    padding: 4px 12px;
    background: none;
    border: 1px solid var(--border);
    color: var(--fg);
    cursor: pointer;
  }

  .primary {
    background: var(--bg-selected);
    border-color: var(--bg-selected);
  }
</style>
