<script lang="ts">
  import type { git } from '../../wailsjs/go/models'

  let { diff }: { diff: git.FileDiff | null } = $props()

  function lineClass(line: string): string {
    if (line.startsWith('+') && !line.startsWith('+++')) return 'add'
    if (line.startsWith('-') && !line.startsWith('---')) return 'del'
    if (line.startsWith('@@')) return 'hunk'
    return ''
  }
</script>

<!-- Plain patch view until the diff viewer lands (#30). -->
<div class="diff">
  {#if !diff}
    <p class="muted">Loading diff…</p>
  {:else if diff.binary}
    <p class="muted">Binary file, no text diff.</p>
  {:else if !diff.patch}
    <p class="muted">No changes.</p>
  {:else}
    <pre>{#each diff.patch.split('\n') as line, i (i)}<span class={lineClass(line)}>{line}
</span>{/each}</pre>
    {#if diff.truncated}<p class="muted">Diff truncated at 1 MB.</p>{/if}
  {/if}
</div>

<style>
  .diff {
    border-top: 1px solid var(--border);
  }

  pre {
    margin: 0;
    padding: 8px 0;
    overflow-x: auto;
    font-family: var(--font-mono);
    font-size: 12px;
  }

  span {
    display: block;
    padding: 0 12px;
    white-space: pre;
  }

  .add {
    background: rgba(80, 160, 80, 0.2);
  }

  .del {
    background: rgba(200, 80, 80, 0.2);
  }

  .hunk {
    color: var(--fg-muted);
  }

  .muted {
    padding: 0 12px;
    color: var(--fg-muted);
  }
</style>
