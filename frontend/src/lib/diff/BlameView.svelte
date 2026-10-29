<script lang="ts">
  import { Blame } from '../../../wailsjs/go/app/App.js'
  import type { git } from '../../../wailsjs/go/models'
  import { formatTime, laneVar } from '../graph/draw'
  import { session } from '../session.svelte'
  import type { DiffTarget } from './DiffView.svelte'
  import { highlightLines, languageFor, type Token } from './highlight'

  let {
    path,
    file,
    rev = '',
    onClose,
    onOpen,
    onReveal,
  }: {
    path: string
    file: string
    rev?: string
    onClose: () => void
    onOpen: (t: DiffTarget) => void
    /** Shows a commit in the graph. */
    onReveal: (commit: string) => void
  } = $props()

  let lines = $state<git.BlameLine[] | null>(null)
  let tokens = $state<Token[][]>([])

  /** A stable color per commit, so a commit's blocks read as one. */
  let colorOf = $derived.by(() => {
    const m: Record<string, number> = {}
    let n = 0
    for (const l of lines ?? []) if (!(l.commit in m)) m[l.commit] = n++
    return m
  })

  $effect(() => {
    const [f, r] = [file, rev]
    lines = null
    void session.run(() => Blame(path, f, r)).then(async (res) => {
      if (f !== file || r !== rev) return
      lines = res ?? []
      const text = lines.map((l) => l.line).join('\n')
      tokens = highlightLines(text, null)
      const lang = await languageFor(f)
      if (lang && f === file) tokens = highlightLines(text, lang)
    })
  })

  const uncommitted = (c: string) => /^0+$/.test(c)
</script>

<div class="blame">
  <header>
    <button class="back" onclick={onClose} title="Back">←</button>
    <span class="file">Blame {file}{rev ? ` at ${rev.slice(0, 7)}` : ''}</span>
    <span class="spacer"></span>
    <button onclick={() => onOpen({ mode: 'history', file })}>History</button>
  </header>
  <div class="body">
    {#if !lines}
      <p class="muted">Loading…</p>
    {:else}
      <table>
        <tbody>
          {#each lines as l, i (i)}
            {@const start = i === 0 || lines[i - 1].commit !== l.commit}
            <tr class:start>
              <td class="gutter" style:border-left-color={uncommitted(l.commit) ? 'transparent' : laneVar(colorOf[l.commit] ?? 0)}>
                {#if start}
                  {#if uncommitted(l.commit)}
                    <span class="muted">Not committed yet</span>
                  {:else}
                    <button class="sha" onclick={() => onReveal(l.commit)} title="{l.summary}&#10;Show in graph">{l.commit.slice(0, 7)}</button>
                    <span class="who">{l.author}</span>
                    <span class="muted">{formatTime(l.time)}</span>
                  {/if}
                {/if}
              </td>
              <td class="num">{i + 1}</td>
              <td class="code"
                >{#if tokens[i]}{#each tokens[i] as t, k (k)}<span class={t.cls}>{t.text}</span>{/each}{:else}{l.line}{/if}</td
              >
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<style>
  .blame {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 8px;
    border-bottom: 1px solid var(--paper-edge);
  }

  .file {
    font-family: var(--font-mono);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .spacer {
    flex: 1;
  }

  header button {
    padding: 2px 8px;
    background: none;
    border: 1px solid var(--paper-edge);
    color: var(--ink);
    cursor: pointer;
  }

  .back {
    border: none !important;
    font-size: 16px;
  }

  .body {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  table {
    border-collapse: collapse;
    font-size: 12px;
  }

  tr.start td {
    border-top: 1px solid var(--paper-edge);
  }

  td {
    padding: 0 8px;
    vertical-align: top;
    white-space: pre;
  }

  .gutter {
    width: 280px;
    max-width: 280px;
    border-left: 3px solid transparent;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: var(--font-serif);
  }

  .sha {
    background: none;
    border: none;
    color: var(--accent-a);
    font-family: var(--font-mono);
    padding: 0;
    cursor: pointer;
  }

  .who {
    margin: 0 6px;
  }

  .num {
    text-align: right;
    color: var(--ink-mute);
    user-select: none;
  }

  .code {
    font-family: var(--font-mono);
    tab-size: 4;
  }

  .muted {
    color: var(--ink-mute);
  }

  p.muted {
    padding: 8px;
  }
</style>
