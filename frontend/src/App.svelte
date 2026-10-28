<script lang="ts">
  import { onMount } from 'svelte'
  import { Version } from '../wailsjs/go/app/App.js'

  let version = $state('')

  onMount(async () => {
    version = await Version()
  })
</script>

<div class="shell">
  <header>
    <strong>Git Revenge</strong>
    <span class="muted">{version}</span>
  </header>
  <main>
    <p class="muted">No repository open.</p>
  </main>
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

  main {
    flex: 1;
    display: grid;
    place-items: center;
  }

  .muted {
    color: var(--fg-muted);
  }
</style>
