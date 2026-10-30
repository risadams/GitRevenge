<script lang="ts">
  import { commands, formatShortcut } from './commands.svelte'
  import ContextMenu, { type MenuItem } from './ContextMenu.svelte'
  import { gitConfig } from './gitconfig/config.svelte'
  import { session } from './session.svelte'
  import { theme } from './theme.svelte'
  import Icon from './ui/Icon.svelte'
  import Mark from './ui/Mark.svelte'

  function onAuxClick(e: MouseEvent, path: string) {
    // Middle-click closes a tab, as in browsers.
    if (e.button === 1) void session.close(path)
  }

  let menu = $state<{ x: number; y: number; items: MenuItem[] } | null>(null)

  function tabMenu(e: MouseEvent, path: string) {
    e.preventDefault()
    menu = {
      x: e.clientX,
      y: e.clientY,
      items: [
        { label: 'Edit git config…', icon: 'gear', run: () => gitConfig.show(path) },
        { label: 'Close tab', icon: 'close', separator: true, run: () => session.close(path) },
      ],
    }
  }

  let toggleLabel = $derived(theme.resolved === 'dark' ? 'Switch to the light theme' : 'Switch to the dark theme')
</script>

<nav class="tabs" aria-label="Open repositories">
  <div class="scroll">
    <button class="tab home" class:active={session.active === ''} onclick={() => session.select('')} title="Start page" aria-label="Start page">
      <Mark size={18} />
    </button>
    {#each session.tabs as tab (tab.path)}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="tab" class:active={session.active === tab.path} title={tab.path} oncontextmenu={(e) => tabMenu(e, tab.path)}>
        <button class="label" onclick={() => session.select(tab.path)} onauxclick={(e) => onAuxClick(e, tab.path)}>
          {tab.name}
          {#if tab.head}<span class="head">{tab.head}</span>{/if}
        </button>
        <button class="close" aria-label="Close {tab.name}" title="Close (middle-click)" onclick={() => session.close(tab.path)}>
          <Icon name="close" size={12} />
        </button>
      </div>
    {/each}
    {#if gitConfig.open}
      <div class="tab" class:active={session.active === session.configTabId} title="Git config">
        <button
          class="label"
          onclick={() => session.select(session.configTabId)}
          onauxclick={(e) => e.button === 1 && gitConfig.requestClose()}
        >
          <Icon name="gear" size={14} />
          Git Config
          {#if gitConfig.dirtyPaths.length > 0}<span class="dot" aria-label="unsaved changes" title="Unsaved changes">●</span>{/if}
        </button>
        <button class="close" aria-label="Close Git Config" title="Close (middle-click)" onclick={() => gitConfig.requestClose()}>
          <Icon name="close" size={12} />
        </button>
      </div>
    {/if}
    <button class="tab add" onclick={() => session.choose()} title="Open repository" aria-label="Open repository">
      <Icon name="plus" />
    </button>
  </div>
  <button class="palette" onclick={() => (commands.paletteOpen = true)} title="Command palette" aria-label="Open the command palette">
    <Icon name="search" size={14} />
    <span>Commands</span>
    <kbd>{formatShortcut('Mod+K')}</kbd>
  </button>
  <button class="theme" onclick={() => (commands.settingsOpen = true)} aria-label="Settings" title="Settings ({formatShortcut('Mod+,')})">
    <Icon name="gear" size={17} />
  </button>
  <button class="theme" onclick={() => session.run(() => theme.toggle())} aria-label={toggleLabel} title={toggleLabel}>
    <Icon name={theme.resolved === 'dark' ? 'sun' : 'moon'} size={17} />
  </button>
</nav>

{#if menu}
  <ContextMenu x={menu.x} y={menu.y} items={menu.items} onClose={() => (menu = null)} />
{/if}

<style>
  .tabs {
    display: flex;
    align-items: stretch;
    min-height: 40px;
    padding: 0 6px 0 8px;
    background: var(--paper-warm);
    border-bottom: 1px solid var(--paper-edge);
  }

  .scroll {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: flex-end;
    gap: 2px;
    overflow-x: auto;
    scrollbar-width: none;
  }

  button {
    background: none;
    border: none;
    border-radius: 0;
    color: inherit;
    padding: 0 12px;
  }

  button:hover:not(:disabled) {
    background: none;
  }

  .tab {
    position: relative;
    display: flex;
    align-items: center;
    height: 34px;
    color: var(--ink-mute);
    white-space: nowrap;
    font-size: 14px;
    border-radius: 6px 6px 0 0;
    transition:
      color 0.2s,
      background-color 0.2s;
  }

  .tab:hover {
    color: var(--ink);
    background: var(--wash-hover);
  }

  .tab.active {
    color: var(--ink);
    background: var(--paper);
  }

  /* The active tab wears the navbar's torn rose bar. */
  .tab.active::after {
    content: '';
    position: absolute;
    left: 10px;
    right: 10px;
    bottom: 2px;
    height: 3px;
    background: var(--link);
    filter: url(#wc-edge);
    opacity: 0.85;
    border-radius: 45% 55% 50% 50% / 55% 45%;
    pointer-events: none;
  }

  .label {
    display: flex;
    align-items: center;
    gap: 6px;
    height: 100%;
    padding-right: 4px;
  }

  .dot {
    color: var(--link);
    font-size: 10px;
  }

  .home,
  .add {
    padding: 0 10px;
  }

  .head {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--ink-mute);
  }

  .close {
    display: grid;
    place-items: center;
    width: 20px;
    height: 20px;
    margin-right: 6px;
    padding: 0;
    border-radius: 50%;
    color: var(--ink-mute);
    opacity: 0;
    transition: opacity 0.2s;
  }

  .tab:hover .close,
  .tab.active .close,
  .close:focus-visible {
    opacity: 1;
  }

  .close:hover:not(:disabled) {
    background: var(--wash-hover);
    color: var(--ink);
  }

  .palette {
    align-self: center;
    display: flex;
    align-items: center;
    gap: 8px;
    height: 28px;
    margin-right: 6px;
    padding: 0 6px 0 10px;
    border: 1px solid var(--paper-edge);
    border-radius: 6px;
    background: var(--paper);
    color: var(--ink-mute);
    font-size: 13px;
  }

  .palette:hover:not(:disabled) {
    background: var(--paper);
    color: var(--ink);
    border-color: var(--ink-faint);
  }

  .palette kbd {
    padding: 0 5px;
    border-radius: 3px;
    background: var(--wash-hover);
    font-family: var(--font-mono);
    font-size: 11px;
  }

  .theme {
    position: relative;
    align-self: center;
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    padding: 0;
    border-radius: 50%;
  }

  .theme::before {
    content: '';
    position: absolute;
    inset: 3px;
    background: var(--accent-a);
    opacity: 0;
    filter: url(#wc-edge);
    border-radius: 45% 55% 50% 50% / 55% 45%;
    transition: opacity 0.3s;
  }

  .theme:hover::before {
    opacity: 0.14;
  }

  .theme > :global(svg) {
    position: relative;
  }
</style>
