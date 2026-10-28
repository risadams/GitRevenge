<script lang="ts" module>
  export interface MenuItem {
    label: string
    /** Destructive items ask for a second click before running. */
    danger?: boolean
    run: () => void | Promise<void>
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'

  let { x, y, items, onClose }: { x: number; y: number; items: MenuItem[]; onClose: () => void } = $props()

  let menu: HTMLDivElement
  let armed = $state(-1)
  // Keep the menu on screen near the window edges.
  let left = $state(0)
  let top = $state(0)

  onMount(() => {
    const r = menu.getBoundingClientRect()
    left = Math.min(x, window.innerWidth - r.width - 4)
    top = Math.min(y, window.innerHeight - r.height - 4)
    menu.focus()
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
</script>

<svelte:window
  onmousedown={(e) => {
    if (!menu.contains(e.target as Node)) onClose()
  }}
  onkeydown={(e) => e.key === 'Escape' && onClose()}
  onblur={onClose}
/>

<div class="menu" bind:this={menu} role="menu" tabindex="-1" style:left="{left}px" style:top="{top}px">
  {#each items as item, i (item.label)}
    <button role="menuitem" class:danger={item.danger} onclick={() => pick(i)}>
      {armed === i ? `Click again to ${item.label.toLowerCase()}` : item.label}
    </button>
  {/each}
</div>

<style>
  .menu {
    position: fixed;
    z-index: 100;
    min-width: 160px;
    padding: 4px 0;
    background: var(--bg-raised);
    border: 1px solid var(--border);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
    outline: none;
  }

  button {
    display: block;
    width: 100%;
    padding: 5px 14px;
    background: none;
    border: none;
    color: var(--fg);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  button:hover {
    background: var(--bg-selected);
  }

  .danger {
    color: #f48771;
  }
</style>
