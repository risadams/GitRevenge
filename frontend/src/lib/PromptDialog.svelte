<script lang="ts">
  import { onMount } from 'svelte'
  import { AnswerPrompt } from '../../wailsjs/go/app/App.js'
  import { EventsOn } from '../../wailsjs/runtime/runtime.js'

  /** A credential or host-key prompt relayed from git/ssh (app.Prompt in Go). */
  interface Prompt {
    id: string
    text: string
    kind: 'text' | 'secret' | 'hostkey'
  }

  // Prompts can arrive back to back (username, then password); show them in order.
  let queue = $state<Prompt[]>([])
  let current = $derived(queue[0])
  let value = $state('')
  let input = $state<HTMLInputElement | null>(null)

  function answer(ok: boolean) {
    if (!current) return
    const p = current
    AnswerPrompt(p.id, p.kind === 'hostkey' ? 'yes' : value, ok)
    value = ''
    queue = queue.slice(1)
  }

  $effect(() => {
    if (current && input) input.focus()
  })

  onMount(() => {
    const offPrompt = EventsOn('askpass:prompt', (p: Prompt) => (queue = [...queue, p]))
    const offExpired = EventsOn('askpass:expired', (id: string) => (queue = queue.filter((p) => p.id !== id)))
    return () => {
      offPrompt()
      offExpired()
    }
  })
</script>

{#if current}
  <div class="backdrop">
    <div class="dialog" role="dialog" aria-modal="true" aria-label="Credentials requested">
      <form
        onsubmit={(e) => {
          e.preventDefault()
          answer(true)
        }}
      >
        {#if current.kind === 'hostkey'}
          <h3>Trust this SSH host?</h3>
          <!-- ssh's own wording, including the key fingerprint to compare. -->
          <pre>{current.text}</pre>
          <p class="muted">
            Only continue if the fingerprint matches the one your Git host publishes. The key is then saved to
            known_hosts.
          </p>
          <div class="buttons">
            <button type="button" onclick={() => answer(false)}>Cancel</button>
            <button type="submit" class="primary">Trust and connect</button>
          </div>
        {:else}
          <h3>Git needs your credentials</h3>
          <label>
            <span>{current.text}</span>
            {#if current.kind === 'secret'}
              <input type="password" bind:this={input} bind:value autocomplete="off" />
            {:else}
              <input bind:this={input} bind:value autocomplete="off" />
            {/if}
          </label>
          <p class="muted">Git Revenge passes this to git and doesn't store it.</p>
          <div class="buttons">
            <button type="button" onclick={() => answer(false)}>Cancel</button>
            <button type="submit" class="primary">Continue</button>
          </div>
        {/if}
      </form>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 300;
    display: grid;
    place-items: center;
    background: var(--scrim);
  }

  .dialog {
    width: min(480px, calc(100vw - 32px));
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 16px;
    background: var(--paper-warm);
    border: 1px solid var(--paper-edge);
  }

  h3 {
    margin: 0;
    font-size: 14px;
  }

  pre {
    margin: 0;
    padding: 8px;
    background: var(--paper);
    white-space: pre-wrap;
    font-family: var(--font-mono);
    font-size: 12px;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  input {
    padding: 4px 6px;
    background: var(--paper);
    color: var(--ink);
    border: 1px solid var(--paper-edge);
    font: inherit;
  }

  .muted {
    margin: 0;
    color: var(--ink-mute);
    font-size: 12px;
  }

  .buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  button {
    padding: 4px 12px;
    background: none;
    border: 1px solid var(--paper-edge);
    color: var(--ink);
    cursor: pointer;
  }

  .primary {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--on-accent);
  }
</style>
