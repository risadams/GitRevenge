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

  let version = $state('')
  let env = $state<app.Environment | null>(null)

  onMount(async () => {
    ;[version, env] = await Promise.all([Version(), Environment()])
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
    background: var(--bg-raised);
    border-bottom: 1px solid var(--border);
  }

  .right {
    margin-left: auto;
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
    background: #5a1d1d;
  }

  .problem {
    max-width: 480px;
  }

  .muted {
    color: var(--fg-muted);
  }

  code {
    font-family: var(--font-mono);
  }
</style>
