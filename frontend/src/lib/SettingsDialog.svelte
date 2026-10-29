<script lang="ts">
  // The settings panel: appearance, git, repositories and updates, saved to the local
  // settings file (ADR 0005).
  import { ChooseFile, ChooseFolder, Preferences, SetPreferences } from '../../wailsjs/go/app/App.js'
  import { commands } from './commands.svelte'
  import type { app } from '../../wailsjs/go/models'
  import { errorText } from './errors'
  import { prefs } from './prefs.svelte'
  import { theme } from './theme.svelte'
  import Icon from './ui/Icon.svelte'
  import InkButton from './ui/InkButton.svelte'
  import Modal from './ui/Modal.svelte'
  import { update } from './update.svelte'

  let { onClose }: { onClose: () => void } = $props()

  let p = $state<app.Preferences | null>(null)
  let initialGit = ''
  let error = $state('')
  let saving = $state(false)
  let restartNote = $state(false)

  $effect(() => {
    void Preferences()
      .then((res) => {
        p = res
        initialGit = res.gitPath
      })
      .catch((e) => (error = errorText(e)))
  })

  async function chooseGit() {
    const f = await ChooseFile('Choose the git executable')
    if (f && p) p.gitPath = f
  }

  async function chooseCloneDir() {
    const d = await ChooseFolder('Folder to clone repositories into')
    if (d && p) p.cloneDir = d
  }

  async function save() {
    if (!p) return
    error = ''
    saving = true
    try {
      await SetPreferences(p)
      // Apply what can change now; the git executable waits for a restart.
      theme.mode = p.theme as typeof theme.mode
      theme.density = p.graphDensity === 'compact' ? 'compact' : 'comfortable'
      update.checksOn = p.updateChecks
      prefs.zoom = p.zoom
      prefs.cloneDir = p.cloneDir
      prefs.apply()
      if (p.gitPath !== initialGit) {
        restartNote = true
        initialGit = p.gitPath
      } else onClose()
    } catch (e) {
      error = errorText(e)
    } finally {
      saving = false
    }
  }
</script>

<Modal title="Settings" eyebrow="make it yours" width={560} {onClose}>
  {#if !p}
    <p class="muted">{error || 'Loading…'}</p>
  {:else}
    <form
      onsubmit={(e) => {
        e.preventDefault()
        void save()
      }}
    >
      <fieldset>
        <legend class="meta">Appearance</legend>
        <label class="form-field">
          <span>Theme</span>
          <select bind:value={p.theme}>
            <option value="">Follow the system</option>
            <option value="light">Light (Ink Paper)</option>
            <option value="dark">Dark (Ink Parchment)</option>
          </select>
        </label>
        <label class="form-field">
          <span>Commit graph</span>
          <select bind:value={p.graphDensity}>
            <option value="">Timeline (two lines per commit)</option>
            <option value="compact">Compact rows</option>
          </select>
        </label>
        <label class="form-field">
          <span>Zoom: {p.zoom}%</span>
          <input type="range" min="80" max="150" step="5" bind:value={p.zoom} aria-valuetext="{p.zoom} percent" />
        </label>
      </fieldset>

      <fieldset>
        <legend class="meta">Git</legend>
        <label class="form-field">
          <span>Git executable (empty to find it automatically)</span>
          <span class="row">
            <input bind:value={p.gitPath} placeholder={p.gitInUse} />
            <button type="button" class="tool" onclick={chooseGit}><Icon name="folder" size={14} />Choose…</button>
          </span>
        </label>
        <p class="hint">In use now: <code>{p.gitInUse || 'none found'}</code>. A change takes effect after restarting Git Revenge.</p>
      </fieldset>

      <fieldset>
        <legend class="meta">Repositories</legend>
        <label class="form-field">
          <span>Clone into (empty for the folder of the last repository opened)</span>
          <span class="row">
            <input bind:value={p.cloneDir} />
            <button type="button" class="tool" onclick={chooseCloneDir}><Icon name="folder" size={14} />Choose…</button>
          </span>
        </label>
      </fieldset>

      <fieldset>
        <legend class="meta">Keyboard</legend>
        <span>
          <button
            type="button"
            class="tool"
            onclick={() => {
              onClose()
              commands.keybindingsOpen = true
            }}><Icon name="keyboard" size={14} />Customize keyboard shortcuts…</button
          >
        </span>
      </fieldset>

      <fieldset>
        <legend class="meta">Updates</legend>
        <label class="check"><input type="checkbox" bind:checked={p.updateChecks} /> Check GitHub for a newer version at startup</label>
      </fieldset>

      {#if error}<p class="error" role="alert">{error}</p>{/if}
      {#if restartNote}<p class="note" role="status">Saved. Restart Git Revenge to use the new git.</p>{/if}
      <span class="buttons">
        <InkButton type="submit" disabled={saving}>Save</InkButton>
        <InkButton variant="ghost" onclick={onClose}>{restartNote ? 'Close' : 'Cancel'}</InkButton>
      </span>
    </form>
  {/if}
</Modal>

<style>
  form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  fieldset {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin: 0;
    padding: 0;
    border: none;
  }

  legend {
    margin-bottom: 6px;
  }

  .row {
    display: flex;
    gap: 6px;
  }

  .row input {
    flex: 1;
    min-width: 0;
  }

  .hint {
    margin: -4px 0 0;
    font-size: 13px;
    color: var(--ink-mute);
  }

  .hint code {
    font-family: var(--font-mono);
    font-size: 12px;
    overflow-wrap: anywhere;
  }

  .check {
    font-size: 14px;
  }

  input[type='range'] {
    accent-color: var(--accent);
  }

  .buttons {
    display: flex;
    gap: 8px;
  }

  .muted {
    color: var(--ink-mute);
    font-style: italic;
  }

  .error {
    color: var(--danger);
  }

  .note {
    color: var(--success);
  }
</style>
