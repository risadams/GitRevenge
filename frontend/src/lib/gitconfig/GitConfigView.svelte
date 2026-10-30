<script lang="ts">
  // The Git Config tab (#195): pick a config file, edit its working copy, save or discard.
  import Icon from '../ui/Icon.svelte'
  import InkButton from '../ui/InkButton.svelte'
  import Modal from '../ui/Modal.svelte'
  import { gitConfig, problemText } from './config.svelte'

  let scope = $derived(gitConfig.scope)
  let doc = $derived(gitConfig.doc)
  let outside = $derived(scope ? gitConfig.outside[scope.path] : false)

  function onKey(e: KeyboardEvent) {
    // Mod+S saves, as in any editor.
    if ((e.ctrlKey || e.metaKey) && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 's') {
      e.preventDefault()
      if (doc?.dirty) void gitConfig.save()
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<section class="config" aria-label="Git config">
  <header class="view-head">
    <label class="scope">
      <span class="visually-hidden">Config file</span>
      <select value={scope?.path ?? ''} onchange={(e) => gitConfig.select(e.currentTarget.value)}>
        {#each gitConfig.scopes as s (s.path)}
          <option value={s.path}>{gitConfig.docs[s.path]?.dirty ? '● ' : ''}{gitConfig.label(s)}</option>
        {/each}
      </select>
    </label>
    {#if scope}
      <span class="view-file" title={scope.path}>{scope.path}</span>
      {#if doc && !doc.exists}<span class="view-badge">new file</span>{/if}
    {/if}
    <span class="spacer"></span>
    {#if doc?.dirty}
      <button class="tool" disabled={gitConfig.busy} onclick={() => gitConfig.reload()}><Icon name="undo" size={14} />Discard changes</button>
    {/if}
    <InkButton disabled={!doc?.dirty || gitConfig.busy} onclick={() => gitConfig.save()}>Save</InkButton>
  </header>

  {#if scope?.kind === 'global' && gitConfig.globalAlt}
    <p class="note" role="note">
      <Icon name="info" size={14} />Git also reads <code>{gitConfig.globalAlt}</code>. Changes are saved to the file above.
    </p>
  {/if}
  {#if outside}
    <div class="banner" role="alert">
      <Icon name="warning" size={14} />
      <span class="message">This file changed outside Git Revenge while you were editing it.</span>
      <button class="tool" onclick={() => gitConfig.reload()}>Reload (discard mine)</button>
      <button class="tool" onclick={() => gitConfig.keepMine()}>Keep mine</button>
    </div>
  {/if}
  {#if gitConfig.error}
    <div class="banner" role="alert">
      <Icon name="warning" size={14} />
      <span class="message">{gitConfig.error}</span>
      <button class="tool" onclick={() => (gitConfig.error = '')}>Dismiss</button>
    </div>
  {:else if doc?.problem}
    <div class="banner" role="alert">
      <Icon name="warning" size={14} />
      <span class="message">Git can't read this file ({problemText(doc.problem)}). Fix it in the Raw view before saving.</span>
    </div>
  {/if}

  <div class="body">
    {#if doc}
      <ul class="entries">
        {#each doc.entries as e, i (i)}
          <li><code class="key">{e.key}</code> <span class="value">{e.implicit ? '(true)' : e.value}</span></li>
        {:else}
          <li class="meta">This file has no settings yet.</li>
        {/each}
      </ul>
    {/if}
  </div>
</section>

{#if gitConfig.conflict}
  <Modal title="Overwrite the changed file?" onClose={() => (gitConfig.conflict = null)} width={640}>
    <p>
      <code>{gitConfig.conflict.path}</code> changed on disk since you opened it. Saving replaces it with your version. The current file is backed up first.
    </p>
    <pre class="disk" aria-label="The file on disk now">{gitConfig.conflict.disk}</pre>
    <div class="form-actions">
      <button class="tool" onclick={() => gitConfig.reload()}>Discard mine and reload</button>
      <InkButton onclick={() => gitConfig.save(true)}>Overwrite</InkButton>
    </div>
  </Modal>
{/if}

<style>
  .config {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .scope select {
    min-width: 200px;
  }

  .spacer {
    flex: 1;
  }

  .note,
  .banner {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    padding: 6px 14px;
    border-bottom: 1px solid var(--paper-edge);
  }

  .note {
    color: var(--ink-mute);
  }

  .banner {
    background: var(--danger-wash);
  }

  .banner > :global(svg) {
    color: var(--danger);
  }

  .message {
    flex: 1;
    min-width: 0;
  }

  .body {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 12px 14px;
  }

  .entries {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .entries li {
    padding: 3px 0;
  }

  code,
  .disk {
    font-family: var(--font-mono);
    font-size: 12px;
  }

  .key {
    color: var(--accent-b);
  }

  .disk {
    max-height: 40vh;
    overflow: auto;
    padding: 10px;
    background: var(--paper-warm);
    border: 1px solid var(--paper-edge);
    border-radius: 6px;
    white-space: pre;
  }
</style>
