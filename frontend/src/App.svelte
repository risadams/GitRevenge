<script lang="ts">
  import { onMount } from 'svelte'
  import { Environment, Version } from '../wailsjs/go/app/App.js'
  import type { app } from '../wailsjs/go/models'
  import { BrowserOpenURL } from '../wailsjs/runtime/runtime.js'
  import OperationBanner from './lib/OperationBanner.svelte'
  import PromptDialog from './lib/PromptDialog.svelte'
  import RepoView from './lib/RepoView.svelte'
  import { session } from './lib/session.svelte'
  import StartPage from './lib/StartPage.svelte'
  import TabBar from './lib/TabBar.svelte'
  import { theme } from './lib/theme.svelte'

  let version = $state('')
  let env = $state<app.Environment | null>(null)

  onMount(async () => {
    ;[version, env] = await Promise.all([Version(), Environment(), theme.load()])
    if (!env.problem) await session.restore()
  })
</script>

<PromptDialog />

<div class="shell">
  <header>
    <strong>Git Revenge</strong>
    <span class="muted">{version}</span>
    {#if env && !env.problem}
      <span class="muted right">git {env.gitVersion}{env.lfsVersion ? '' : ' · git-lfs not found'}</span>
    {/if}
    <button
      class="theme"
      class:right={!env || !!env.problem}
      onclick={() => session.run(() => theme.toggle())}
      aria-label={theme.resolved === 'dark' ? 'Switch to the light theme' : 'Switch to the dark theme'}
      title={theme.resolved === 'dark' ? 'Switch to the light theme' : 'Switch to the dark theme'}
    >
      {#if theme.resolved === 'dark'}
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" aria-hidden="true"
          ><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></svg
        >
      {:else}
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"
          ><path d="M20 14.5A8 8 0 1 1 9.5 4a6.5 6.5 0 0 0 10.5 10.5z" /></svg
        >
      {/if}
    </button>
  </header>

  {#if env?.problem}
    <main class="center">
      <div class="problem">
        <h2>Git is required</h2>
        <p>{env.problem}</p>
        <p class="muted">
          Install git {env.minVersion} or newer, then restart Git Revenge.
          {#if env.gitPath}Found: <code>{env.gitPath}</code>{/if}
        </p>
        <button onclick={() => BrowserOpenURL('https://git-scm.com/downloads')}>Download git</button>
      </div>
    </main>
  {:else if env}
    <TabBar />
    {#if session.error}
      <div class="banner" role="alert">
        <span>{session.error}</span>
        <button onclick={() => (session.error = '')}>Dismiss</button>
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
  {/if}
</div>

<style>
  .shell {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  header {
    display: flex;
    gap: 8px;
    align-items: baseline;
    padding: 8px 12px;
    background: var(--paper-warm);
    border-bottom: 1px solid var(--paper-edge);
  }

  .right {
    margin-left: auto;
  }

  .theme {
    display: grid;
    place-items: center;
    align-self: center;
    width: 28px;
    height: 28px;
    padding: 0;
    background: none;
    border: none;
    border-radius: 50%;
    color: var(--ink);
    cursor: pointer;
  }

  .theme:hover {
    background: var(--wash-hover);
  }

  main {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  .center {
    display: grid;
    place-items: center;
    height: 100%;
  }

  .banner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 12px;
    background: var(--danger-wash);
  }

  .problem {
    max-width: 480px;
  }

  .muted {
    color: var(--ink-mute);
  }

  code {
    font-family: var(--font-mono);
  }
</style>
