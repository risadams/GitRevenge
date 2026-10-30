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
  <label class="check">
    <input type="checkbox" checked={bool} {disabled} aria-label={label} onchange={(e) => onCommit(e.currentTarget.checked ? 'true' : 'false')} />
    <span class="spelling">{implicit ? 'true' : value}</span>
  </label>
{:else if info?.type === 'enum'}
  <select {value} {disabled} aria-label={label} onchange={(e) => onCommit(e.currentTarget.value)}>
    {#each options as o (o)}<option value={o}>{o}</option>{/each}
  </select>
{:else}
  <span class="text">
    <input
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
      <button class="tool" {disabled} onclick={browse} title="Choose a file" aria-label="Choose a file for {label}"><Icon name="folder" size={14} /></button>
    {/if}
  </span>
{/if}

<style>
  .check {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .spelling {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--ink-mute);
  }

  .text {
    display: flex;
    gap: 4px;
    width: 100%;
  }

  input[type='text'],
  input[type='number'],
  select {
    width: 100%;
    min-width: 0;
    font-family: var(--font-mono);
    font-size: 12px;
  }

  select {
    width: auto;
  }
</style>
