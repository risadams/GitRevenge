<script lang="ts">
  // One config value, edited with the control its type calls for. Text commits on Enter or
  // when focus leaves; Escape puts the value back.
  import { ChooseFile } from '../../../wailsjs/go/app/App.js'
  import Icon from '../ui/Icon.svelte'
  import { parseBool, type KeyInfo } from './catalog'

  let {
    label,
    value,
    implicit = false,
    info,
    disabled = false,
    onCommit,
  }: {
    /** Names the control for assistive technology, usually the key. */
    label: string
    value: string
    implicit?: boolean
    info?: KeyInfo
    disabled?: boolean
    onCommit: (value: string) => void
  } = $props()

  /** The text being typed; it follows value until edited. */
  let draft = $derived(value)

  let bool = $derived(info?.type === 'bool' ? parseBool(value, implicit) : undefined)
  let options = $derived(info?.values && !info.values.includes(value) ? [value, ...info.values] : (info?.values ?? []))
  /** A few choices read best side by side; more go in a list. */
  let pills = $derived(options.length > 0 && options.length <= 4)

  function commit() {
    if (draft !== value) onCommit(draft)
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      commit()
    } else if (e.key === 'Escape' && draft !== value) {
      e.stopPropagation()
      draft = value
    }
  }

  async function browse() {
    const p = await ChooseFile(`Choose ${label}`).catch(() => '')
    if (p) onCommit(p)
  }
</script>

{#if info?.type === 'bool' && bool !== undefined}
  <label class="switch">
    <input type="checkbox" checked={bool} {disabled} aria-label={label} onchange={(e) => onCommit(e.currentTarget.checked ? 'true' : 'false')} />
    <span class="track"></span>
    <span class="spelling">{implicit ? 'true' : value}</span>
  </label>
{:else if info?.type === 'enum' && pills}
  <span class="segmented pills" role="group" aria-label={label}>
    {#each options as o (o)}
      <button aria-pressed={o === value} {disabled} onclick={() => o !== value && onCommit(o)}>{o}</button>
    {/each}
  </span>
{:else if info?.type === 'enum'}
  <select class="field" {value} {disabled} aria-label={label} onchange={(e) => onCommit(e.currentTarget.value)}>
    {#each options as o (o)}<option value={o}>{o}</option>{/each}
  </select>
{:else}
  <span class="text">
    <input
      class="field"
      type={info?.type === 'int' ? 'number' : 'text'}
      bind:value={draft}
      {disabled}
      aria-label={label}
      spellcheck="false"
      autocomplete="off"
      onkeydown={onKey}
      onblur={commit}
    />
    {#if info?.type === 'path'}
      <button class="round" {disabled} onclick={browse} title="Choose a file" aria-label="Choose a file for {label}"><Icon name="folder" size={14} /></button>
    {/if}
  </span>
{/if}

<style>
  .spelling {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--ink-soft);
  }

  .segmented {
    align-self: flex-start;
  }

  .text {
    display: flex;
    gap: 6px;
    width: 100%;
  }

  .text .field {
    flex: 1;
    min-width: 0;
  }

  select.field {
    align-self: flex-start;
  }

  .round {
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    padding: 0;
    background: none;
    border: 1px solid var(--paper-edge);
    border-radius: 50%;
    color: var(--ink-soft);
  }

  .round:hover:not(:disabled) {
    background: var(--wash-hover);
    color: var(--ink);
  }
</style>
