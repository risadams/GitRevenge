<script lang="ts">
  import { session } from './session.svelte'

  function onAuxClick(e: MouseEvent, path: string) {
    // Middle-click closes a tab, as in browsers.
    if (e.button === 1) void session.close(path)
  }
</script>

<nav class="tabs" aria-label="Open repositories">
  <button class="tab home" class:active={session.active === ''} onclick={() => session.select('')} title="Start page">
    ⌂
  </button>
  {#each session.tabs as tab (tab.path)}
    <div class="tab" class:active={session.active === tab.path} title={tab.path}>
      <button class="label" onclick={() => session.select(tab.path)} onauxclick={(e) => onAuxClick(e, tab.path)}>
        {tab.name}
        {#if tab.head}<span class="head">{tab.head}</span>{/if}
      </button>
      <button class="close" aria-label="Close {tab.name}" onclick={() => session.close(tab.path)}>×</button>
    </div>
  {/each}
  <button class="tab add" onclick={() => session.choose()} title="Open repository">+</button>
</nav>

<style>
  .tabs {
    display: flex;
    overflow-x: auto;
    background: var(--bg-raised);
    border-bottom: 1px solid var(--border);
  }

  button {
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    cursor: pointer;
    padding: 6px 10px;
  }

  .tab {
    display: flex;
    align-items: center;
    border-right: 1px solid var(--border);
    color: var(--fg-muted);
    white-space: nowrap;
  }

  .tab.active {
    background: var(--bg);
    color: var(--fg);
  }

  .head {
    margin-left: 6px;
    color: var(--fg-muted);
    font-size: 11px;
  }

  .close {
    padding: 6px 8px 6px 0;
  }
</style>
