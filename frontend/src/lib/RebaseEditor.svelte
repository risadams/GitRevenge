<script lang="ts">
  import { InteractiveRebase, RebasePlan } from '../../wailsjs/go/app/App.js'
  import type { git } from '../../wailsjs/go/models'
  import { errorText } from './errors'

  let { path, base, onClose }: { path: string; base: string; onClose: () => void } = $props()

  type Action = 'pick' | 'reword' | 'squash' | 'fixup' | 'drop'
  interface Item {
    commit: string
    subject: string
    original: string
    message: string
    action: Action
  }

  // Shown newest first, like the graph; git wants oldest first, so it's reversed on start.
  let items = $state<Item[]>([])
  let error = $state('')
  let busy = $state(false)

  async function load() {
    try {
      const plan = (await RebasePlan(path, base)) ?? []
      items = plan
        .map((p) => ({ commit: p.commit, subject: p.subject, original: p.message, message: p.message, action: 'pick' as Action }))
        .reverse()
    } catch (e) {
      error = errorText(e)
    }
  }

  $effect(() => {
    void base
    void load()
  })

  function move(i: number, d: -1 | 1) {
    const j = i + d
    if (j < 0 || j >= items.length) return
    ;[items[i], items[j]] = [items[j], items[i]]
  }

  /** The commit a squash/fixup at display index i folds into (the next older kept one). */
  function target(i: number): Item | undefined {
    for (let j = i + 1; j < items.length; j++) if (items[j].action !== 'drop') return items[j]
  }

  function setAction(i: number, a: Action) {
    const it = items[i]
    it.action = a
    if (a === 'squash') {
      // Start from both messages, as git's squash does.
      const into = target(i)
      it.message = into ? `${into.original}\n\n${it.original}` : it.original
    } else if (a === 'reword') {
      it.message = it.original
    }
  }

  let problem = $derived.by(() => {
    if (items.every((it) => it.action === 'drop')) return 'Every commit is dropped. Use Reset instead.'
    for (let i = 0; i < items.length; i++) {
      const it = items[i]
      if ((it.action === 'squash' || it.action === 'fixup') && !target(i)) return 'The oldest kept commit can’t be squashed or fixed up: there’s nothing older to fold it into.'
      if ((it.action === 'reword' || it.action === 'squash') && !it.message.trim()) return `${it.subject} needs a message.`
    }
    return ''
  })

  async function start() {
    busy = true
    error = ''
    try {
      const plan = [...items].reverse().map((it) => ({ commit: it.commit, subject: it.subject, message: it.message, action: it.action }))
      await InteractiveRebase(path, base, plan as git.RebaseItem[])
      onClose()
    } catch (e) {
      error = errorText(e)
    } finally {
      busy = false
    }
  }

  const actions: Action[] = ['pick', 'reword', 'squash', 'fixup', 'drop']
</script>

<div class="rebase">
  <header>
    <button class="back" onclick={onClose} title="Cancel">←</button>
    <strong>Interactive rebase</strong>
    <span class="muted">onto {base.slice(0, 7)} · {items.length} {items.length === 1 ? 'commit' : 'commits'}, newest first</span>
    <span class="spacer"></span>
    <button onclick={onClose}>Cancel</button>
    <button class="primary" disabled={busy || !!problem || items.length === 0} onclick={start} title={problem}>Start rebase</button>
  </header>
  {#if error}<p class="error" role="alert">{error}</p>{:else if problem}<p class="warn">{problem}</p>{/if}

  <ol>
    {#each items as it, i (it.commit)}
      <li class={it.action}>
        <span class="move">
          <button disabled={i === 0} onclick={() => move(i, -1)} aria-label="Move up">↑</button>
          <button disabled={i === items.length - 1} onclick={() => move(i, 1)} aria-label="Move down">↓</button>
        </span>
        <select value={it.action} onchange={(e) => setAction(i, (e.currentTarget as HTMLSelectElement).value as Action)} aria-label="Action for {it.subject}">
          {#each actions as a (a)}<option value={a}>{a}</option>{/each}
        </select>
        <span class="sha">{it.commit.slice(0, 7)}</span>
        <div class="text">
          <span class="subject">{it.subject}</span>
          {#if it.action === 'squash' && target(i)}
            <span class="muted note">folds into {target(i)?.subject}</span>
          {:else if it.action === 'fixup' && target(i)}
            <span class="muted note">folds into {target(i)?.subject}, discarding this message</span>
          {/if}
          {#if it.action === 'reword' || it.action === 'squash'}
            <textarea bind:value={it.message} rows="3" aria-label="New message for {it.subject}"></textarea>
          {/if}
        </div>
      </li>
    {/each}
  </ol>
</div>

<style>
  .rebase {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 8px;
    border-bottom: 1px solid var(--paper-edge);
  }

  .spacer {
    flex: 1;
  }

  button,
  select {
    padding: 2px 8px;
    background: none;
    border: 1px solid var(--paper-edge);
    color: var(--ink);
    font: inherit;
    cursor: pointer;
  }

  select {
    background: var(--paper);
  }

  button:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .primary {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
  }

  .back {
    border: none;
    font-size: 16px;
  }

  ol {
    list-style: none;
    margin: 0;
    padding: 0;
    overflow-y: auto;
  }

  li {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 6px 8px;
    border-bottom: 1px solid var(--paper-edge);
  }

  li.drop .subject {
    text-decoration: line-through;
    color: var(--ink-mute);
  }

  .move {
    display: flex;
    gap: 2px;
  }

  .move button {
    padding: 0 5px;
  }

  .sha {
    font-family: var(--font-mono);
    color: var(--ink-mute);
    padding-top: 2px;
  }

  .text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  textarea {
    background: var(--paper);
    color: var(--ink);
    border: 1px solid var(--paper-edge);
    font: inherit;
    padding: 4px 6px;
    resize: vertical;
  }

  .note {
    font-size: 12px;
  }

  .muted {
    color: var(--ink-mute);
  }

  .warn {
    margin: 0;
    padding: 4px 8px;
    color: var(--warning);
  }

  .error {
    margin: 0;
    padding: 4px 8px;
    color: var(--danger);
    white-space: pre-wrap;
  }
</style>
