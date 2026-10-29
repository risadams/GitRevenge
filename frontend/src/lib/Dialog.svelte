<script lang="ts" module>
  export interface Field {
    key: string
    label: string
    type?: 'text' | 'textarea' | 'checkbox' | 'select'
    value?: string | boolean
    placeholder?: string
    options?: { value: string; label: string }[]
  }

  export interface DialogSpec {
    title: string
    message?: string
    fields?: Field[]
    submit: string
    danger?: boolean
    /** Runs on submit; return an error message to keep the dialog open, or nothing to close. */
    run: (values: Record<string, string | boolean>) => Promise<string | void>
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'

  let { spec, onClose }: { spec: DialogSpec; onClose: () => void } = $props()

  // svelte-ignore state_referenced_locally
  let values = $state<Record<string, string | boolean>>(
    Object.fromEntries((spec.fields ?? []).map((f) => [f.key, f.value ?? (f.type === 'checkbox' ? false : '')])),
  )
  let error = $state('')
  let busy = $state(false)
  let form: HTMLFormElement

  onMount(() => {
    form.querySelector<HTMLElement>('input, textarea, select')?.focus()
  })

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    busy = true
    error = ''
    const res = await spec.run(values)
    busy = false
    if (res) error = res
    else onClose()
  }
</script>

<svelte:window onkeydown={(e) => e.key === 'Escape' && onClose()} />

<div class="backdrop" role="presentation" onmousedown={(e) => e.target === e.currentTarget && onClose()}>
  <div class="dialog" role="dialog" aria-modal="true" aria-label={spec.title}>
    <form bind:this={form} onsubmit={submit}>
      <h3>{spec.title}</h3>
      {#if spec.message}<p>{spec.message}</p>{/if}
      {#each spec.fields ?? [] as f (f.key)}
        {#if f.type === 'checkbox'}
          <label class="check"><input type="checkbox" bind:checked={values[f.key] as boolean} /> {f.label}</label>
        {:else}
          <label>
            <span>{f.label}</span>
            {#if f.type === 'textarea'}
              <textarea bind:value={values[f.key]} placeholder={f.placeholder} rows="4"></textarea>
            {:else if f.type === 'select'}
              <select bind:value={values[f.key]}>
                {#each f.options ?? [] as o (o.value)}<option value={o.value}>{o.label}</option>{/each}
              </select>
            {:else}
              <input bind:value={values[f.key]} placeholder={f.placeholder} />
            {/if}
          </label>
        {/if}
      {/each}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <div class="buttons">
        <button type="button" onclick={onClose}>Cancel</button>
        <button type="submit" class="primary" class:danger={spec.danger} disabled={busy}>{spec.submit}</button>
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
    background: var(--scrim);
  }

  .dialog {
    width: min(440px, calc(100vw - 32px));
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 16px;
    background: var(--paper-warm);
    border: 1px solid var(--paper-edge);
    box-shadow: var(--shadow);
  }

  h3 {
    margin: 0;
    font-size: 14px;
  }

  p {
    margin: 0;
    color: var(--ink-mute);
    white-space: pre-wrap;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  label.check {
    flex-direction: row;
    align-items: center;
    gap: 6px;
  }

  input:not([type='checkbox']),
  textarea,
  select {
    padding: 4px 6px;
    background: var(--paper);
    color: var(--ink);
    border: 1px solid var(--paper-edge);
    font: inherit;
  }

  .error {
    color: var(--danger);
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

  .primary.danger {
    background: var(--danger);
    color: var(--on-accent);
    border-color: var(--danger);
  }
</style>
