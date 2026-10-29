<script lang="ts" module>
  export interface MenuItem {
    label: string
    /** Destructive items ask for a second click before running. */
    danger?: boolean
    /** Shown beside the label, in Mod+Shift+K form. */
    shortcut?: string
    run: () => void | Promise<void>
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { formatShortcut } from './commands.svelte'

  let { x, y, items, onClose }: { x: number; y: number; items: MenuItem[]; onClose: () => void } = $props()

  let menu: HTMLDivElement
  let armed = $state(-1)
  // Keep the menu on screen near the window edges.
  let left = $state(0)
  let top = $state(0)

  const buttons = () => [...menu.querySelectorAll<HTMLButtonElement>('button')]

  onMount(() => {
    const before = document.activeElement as HTMLElement | null
    const r = menu.getBoundingClientRect()
    left = Math.max(4, Math.min(x, window.innerWidth - r.width - 4))
    top = Math.max(4, Math.min(y, window.innerHeight - r.height - 4))
    buttons()[0]?.focus()
    return () => before?.focus?.()
  })

  async function pick(i: number) {
    const item = items[i]
    if (item.danger && armed !== i) {
      armed = i
      return
    }
    onClose()
    await item.run()
  }

  /** Arrow keys, Home and End move between items; Tab closes, as in native menus. */
  function onKey(e: KeyboardEvent) {
    const all = buttons()
    const i = all.indexOf(document.activeElement as HTMLButtonElement)
    const go = (j: number) => {
      e.preventDefault()
      all[(j + all.length) % all.length]?.focus()
    }
    if (e.key === 'ArrowDown') go(i + 1)
    else if (e.key === 'ArrowUp') go(i < 0 ? all.length - 1 : i - 1)
    else if (e.key === 'Home') go(0)
    else if (e.key === 'End') go(all.length - 1)
    else if (e.key === 'Tab') {
      e.preventDefault()
      onClose()
    }
  }
</script>

<svelte:window
  onmousedown={(e) => {
    if (!menu.contains(e.target as Node)) onClose()
  }}
  onkeydown={(e) => e.key === 'Escape' && onClose()}
  onblur={onClose}
/>

<div class="menu" bind:this={menu} role="menu" tabindex="-1" style:left="{left}px" style:top="{top}px" onkeydown={onKey}>
  {#each items as item, i (item.label)}
    <button role="menuitem" class:danger={item.danger} class:armed={armed === i} onclick={() => pick(i)} onmouseenter={(e) => e.currentTarget.focus()}>
      <span class="label">{armed === i ? `Click again to ${item.label.toLowerCase()}` : item.label}</span>
      {#if item.shortcut && armed !== i}<kbd>{formatShortcut(item.shortcut)}</kbd>{/if}
    </button>
  {/each}
</div>

<style>
  .menu {
    position: fixed;
    z-index: 100;
    min-width: 200px;
    max-width: 360px;
    padding: 5px;
    background: var(--paper);
    border: 1px solid var(--paper-edge);
    border-radius: 8px;
    box-shadow: var(--shadow);
    outline: none;
    animation: open 0.16s var(--ease-ink);
  }

  button {
    display: flex;
    align-items: center;
    gap: 16px;
    width: 100%;
    padding: 6px 12px;
    background: none;
    border: none;
    border-radius: 5px;
    font-size: 14px;
    text-align: left;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Hover moves focus, so one highlight serves mouse and keyboard. */
  button:hover:not(:disabled) {
    background: none;
  }

  button:focus {
    outline: none;
    background: var(--selected);
  }

  button:focus-visible {
    outline: 2px solid var(--link);
    outline-offset: -2px;
  }

  .label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  kbd {
    flex: none;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--ink-mute);
  }

  .danger {
    color: var(--danger);
  }

  .armed,
  .armed:focus {
    background: var(--danger-wash);
    font-style: italic;
  }

  @keyframes open {
    from {
      opacity: 0;
      transform: translateY(-3px);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .menu {
      animation: none;
    }
  }
</style>
