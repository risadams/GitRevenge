<script lang="ts">
  import { onMount } from 'svelte'
  import { Environment, Version } from '../wailsjs/go/app/App.js'
  import type { app } from '../wailsjs/go/models'
  import { BrowserOpenURL } from '../wailsjs/runtime/runtime.js'
  import CloneDialog from './lib/CloneDialog.svelte'
  import CommandPalette from './lib/CommandPalette.svelte'
  import SettingsDialog from './lib/SettingsDialog.svelte'
  import ShortcutsDialog from './lib/ShortcutsDialog.svelte'
  import { commands, provideCommands, type Command } from './lib/commands.svelte'
  import OperationBanner from './lib/OperationBanner.svelte'
  import PromptDialog from './lib/PromptDialog.svelte'
  import RepoView from './lib/RepoView.svelte'
  import { session } from './lib/session.svelte'
  import StartPage from './lib/StartPage.svelte'
  import TabBar from './lib/TabBar.svelte'
  import { theme } from './lib/theme.svelte'
  import { prefs } from './lib/prefs.svelte'
  import { update } from './lib/update.svelte'
  import Icon from './lib/ui/Icon.svelte'
  import InkButton from './lib/ui/InkButton.svelte'
  import InkFilters from './lib/ui/InkFilters.svelte'
  import Mark from './lib/ui/Mark.svelte'

  let version = $state('')
  let env = $state<app.Environment | null>(null)

  provideCommands(() => {
    const list: Command[] = [
      {
        id: 'app.palette',
        group: 'App',
        title: 'Command palette',
        shortcut: 'Mod+K',
        aliases: ['Mod+Shift+P'],
        hidden: true,
        global: true,
        run: () => (commands.paletteOpen = !commands.paletteOpen),
      },
      { id: 'app.shortcuts', group: 'App', title: 'Keyboard shortcuts', keywords: 'keys help cheat sheet', shortcut: 'Mod+/', whenTyping: true, run: () => (commands.shortcutsOpen = true) },
      { id: 'app.open', group: 'App', title: 'Open repository…', shortcut: 'Mod+O', whenTyping: true, run: () => session.choose() },
      { id: 'app.clone', group: 'App', title: 'Clone a repository…', run: () => (commands.cloneOpen = true) },
      { id: 'app.settings', group: 'App', title: 'Settings…', keywords: 'preferences options configuration zoom git path', shortcut: 'Mod+,', whenTyping: true, run: () => (commands.settingsOpen = true) },
      {
        id: 'app.theme',
        group: 'View',
        title: theme.resolved === 'dark' ? 'Switch to the light theme' : 'Switch to the dark theme',
        keywords: 'appearance colour color mode',
        run: () => session.run(() => theme.toggle()),
      },
      { id: 'app.update', group: 'App', title: 'Check for updates', keywords: 'new version release download', run: () => session.run(() => update.checkNow()) },
      {
        id: 'app.update-checks',
        group: 'App',
        title: update.checksOn ? 'Turn off update checks at startup' : 'Turn on update checks at startup',
        keywords: 'new version release network privacy',
        run: () => session.run(() => update.toggleChecks()),
      },
      {
        id: 'app.density',
        group: 'View',
        title: theme.density === 'compact' ? 'Show the graph as a timeline' : 'Show the graph as compact rows',
        keywords: 'density rows history',
        run: () => session.run(() => theme.toggleDensity()),
      },
    ]
    if (session.active)
      list.push({ id: 'tabs.home', group: 'Tabs', title: 'Go to the start page', keywords: 'home groups recent', shortcut: 'Mod+Shift+H', run: () => session.select('') })
    session.tabs.forEach((t, i) => {
      if (t.path === session.active)
        list.push({ id: `tabs.close`, group: 'Tabs', title: `Close ${t.name}`, keywords: 'tab', shortcut: 'Mod+W', whenTyping: true, run: () => session.close(t.path) })
      else list.push({ id: `tabs.${t.path}`, group: 'Tabs', title: `Switch to ${t.name}`, keywords: `tab ${t.path} ${t.head}`, run: () => session.select(t.path) })
      // Mod+1 … Mod+9 jump straight to a tab, as in browsers.
      if (i < 9) list.push({ id: `tabs.n${i + 1}`, group: 'Tabs', title: `Tab ${i + 1}`, shortcut: `Mod+${i + 1}`, hidden: true, whenTyping: true, run: () => session.select(t.path) })
    })
    if (session.tabs.length > 0) {
      // Cycle through the start page and every tab.
      const order = ['', ...session.tabs.map((t) => t.path)]
      const step = (d: number) => () => session.select(order[(order.indexOf(session.active) + d + order.length) % order.length])
      list.push(
        { id: 'tabs.next', group: 'Tabs', title: 'Next tab', shortcut: 'Mod+Tab', hidden: true, whenTyping: true, run: step(1) },
        { id: 'tabs.prev', group: 'Tabs', title: 'Previous tab', shortcut: 'Mod+Shift+Tab', hidden: true, whenTyping: true, run: step(-1) },
      )
    }
    const open = new Set(session.tabs.map((t) => t.path))
    for (const r of session.recent) {
      if (r.exists && !open.has(r.path)) list.push({ id: `recent.${r.path}`, group: 'Recent', title: `Open ${r.name}`, keywords: r.path, run: () => session.open(r.path) })
    }
    return list
  })

  /** Every app-wide shortcut is a command's; the dispatcher finds and runs it. */
  function onKey(e: KeyboardEvent) {
    if (!env || env.problem) return
    commands.dispatch(e)
  }

  onMount(async () => {
    ;[version, env] = await Promise.all([Version(), Environment(), theme.load()])
    void prefs.load()
    if (!env.problem) {
      await session.restore()
      void update.startup()
    }
  })
</script>

<svelte:window onkeydown={onKey} />

<PromptDialog />
<InkFilters />
<CommandPalette />
{#if commands.shortcutsOpen}
  <ShortcutsDialog onClose={() => (commands.shortcutsOpen = false)} />
{/if}
{#if commands.settingsOpen}
  <SettingsDialog onClose={() => (commands.settingsOpen = false)} />
{/if}

{#if commands.cloneOpen}
  <CloneDialog onClose={() => (commands.cloneOpen = false)} />
{/if}

<div class="shell">
  {#if env?.problem}
    <main class="center">
      <div class="problem">
        <Mark size={56} label="Git Revenge" />
        <span class="hand">one more thing</span>
        <h1>Git Revenge needs git</h1>
        <p>{env.problem}</p>
        <p class="muted">
          Install git {env.minVersion} or newer, then restart Git Revenge.
          {#if env.gitPath}Found: <code>{env.gitPath}</code>{/if}
        </p>
        <InkButton arrow onclick={() => BrowserOpenURL('https://git-scm.com/downloads')}>Download git</InkButton>
      </div>
    </main>
  {:else if env}
    <TabBar />
    {#if session.error}
      <div class="banner" role="alert">
        <Icon name="warning" />
        <span class="message">{session.error}</span>
        <button class="tool" onclick={() => (session.error = '')}>Dismiss</button>
      </div>
    {/if}
    {#if update.available || update.upToDate}
      <div class="banner notice" role="status">
        <Icon name="update" />
        {#if update.available}
          <span class="message">Git Revenge {update.available.latest} is available. You have {update.available.current}.</span>
          <button class="tool" onclick={() => BrowserOpenURL(update.available!.url)}><Icon name="external" />Download</button>
        {:else}
          <span class="message">{update.upToDate}</span>
        {/if}
        <button class="tool" onclick={() => update.dismiss()}>Dismiss</button>
      </div>
    {/if}
    {#if session.activeTab}
      {#key session.activeTab.path}
        <OperationBanner path={session.activeTab.path} />
      {/key}
    {/if}
    <main>
      {#if session.activeTab}
        {#key session.activeTab.path}
          <RepoView path={session.activeTab.path} />
        {/key}
      {:else}
        <StartPage />
      {/if}
    </main>
    <footer class="status">
      <span class="meta">git {env.gitVersion}</span>
      <span class="meta">{env.lfsVersion ? env.lfsVersion.replace(/^git-lfs\/(\S+).*/, 'lfs $1') : 'git-lfs not found'}</span>
      <span class="spacer"></span>
      <span class="meta">Git Revenge {version}</span>
    </footer>
  {/if}
</div>

<style>
  .shell {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  main {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  .center {
    display: grid;
    place-items: center;
    padding: 24px;
  }

  .problem {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    max-width: 520px;
  }

  .problem > :global(svg) {
    margin-bottom: 12px;
  }

  .problem h1 {
    margin: 6px 0 12px;
    font-size: 36px;
    font-weight: 400;
    letter-spacing: -0.02em;
    line-height: 1.1;
    font-variation-settings:
      'SOFT' 50,
      'WONK' 1;
  }

  .problem p {
    font-size: 16px;
    line-height: 1.6;
  }

  .banner {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 8px 6px 14px;
    background: var(--danger-wash);
    border-bottom: 1px solid var(--paper-edge);
    color: var(--ink);
  }

  .banner > :global(svg) {
    color: var(--danger);
  }

  .notice {
    background: color-mix(in srgb, var(--accent-a) 12%, transparent);
  }

  .notice > :global(svg) {
    color: var(--accent-a);
  }

  .message {
    flex: 1;
    min-width: 0;
    white-space: pre-wrap;
  }

  .status {
    display: flex;
    align-items: center;
    gap: 16px;
    min-height: 24px;
    padding: 0 14px;
    background: var(--paper-warm);
    border-top: 1px solid var(--paper-edge);
  }

  .spacer {
    flex: 1;
  }

  .muted {
    color: var(--ink-mute);
  }

  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
    color: var(--accent-b);
  }
</style>
