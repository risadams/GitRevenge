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
  import Icon from './lib/ui/Icon.svelte'
  import InkButton from './lib/ui/InkButton.svelte'
  import InkFilters from './lib/ui/InkFilters.svelte'

  let version = $state('')
  let env = $state<app.Environment | null>(null)

  onMount(async () => {
    ;[version, env] = await Promise.all([Version(), Environment(), theme.load()])
    if (!env.problem) await session.restore()
  })
</script>

<PromptDialog />
<InkFilters />

<div class="shell">
  {#if env?.problem}
    <main class="center">
      <div class="problem">
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
    max-width: 520px;
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
