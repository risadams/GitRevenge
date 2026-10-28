<script lang="ts">
  import { session } from './session.svelte'
</script>

<section class="start">
  <h2>Open a repository</h2>
  <button onclick={() => session.choose()}>Open repository…</button>

  {#if session.recent.length > 0}
    <h3>Recent</h3>
    <ul>
      {#each session.recent as r (r.path)}
        <li class:missing={!r.exists}>
          <button class="link" disabled={!r.exists} onclick={() => session.open(r.path)} title={r.path}>
            <strong>{r.name}</strong>
            <span class="muted">{r.path}</span>
          </button>
          <button class="forget" aria-label="Remove {r.name} from recents" onclick={() => session.forget(r.path)}>×</button>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .start {
    max-width: 640px;
    margin: 48px auto;
    padding: 0 16px;
  }

  ul {
    list-style: none;
    padding: 0;
  }

  li {
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--border);
  }

  li.missing {
    opacity: 0.5;
  }

  .link {
    flex: 1;
    display: flex;
    gap: 12px;
    align-items: baseline;
    text-align: left;
    background: none;
    border: none;
    color: inherit;
    padding: 8px 0;
    cursor: pointer;
    min-width: 0;
  }

  .link .muted {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .forget {
    background: none;
    border: none;
    color: var(--fg-muted);
    cursor: pointer;
  }

  .muted {
    color: var(--fg-muted);
  }
</style>
