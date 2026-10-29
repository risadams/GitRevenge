<script lang="ts">
  // A commit message body rendered as markdown. marked only tokenizes; every token becomes a
  // Svelte element, so nothing is ever inserted as HTML. Raw HTML in a message shows as its
  // text, and links open in the system browser.
  import { marked, type Token, type Tokens } from 'marked'
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js'
  import { valueParts } from './trailers'

  let { text, issueURL = '' }: { text: string; issueURL?: string } = $props()

  let tokens = $derived(marked.lexer(text, { gfm: true }))

  const SAFE_LINK = /^(https?:|mailto:)/i

  function open(e: MouseEvent, href: string) {
    e.preventDefault()
    if (SAFE_LINK.test(href)) BrowserOpenURL(href)
  }
</script>

{#snippet links(value: string)}{#each valueParts(value, issueURL) as part, k (k)}{#if part.href}<a
        href={part.href}
        title={part.href}
        onclick={(e) => open(e, part.href!)}>{part.text}</a
      >{:else}{part.text}{/if}{/each}{/snippet}

{#snippet inline(list: Token[])}{#each list as t, i (i)}{#if t.type === 'text' && 'tokens' in t && t.tokens?.length}{@render inline(
        t.tokens,
      )}{:else if t.type === 'text' || t.type === 'escape'}{@render links(t.text)}{:else if t.type === 'strong'}<strong
        >{@render inline((t as Tokens.Strong).tokens)}</strong
      >{:else if t.type === 'em'}<em>{@render inline((t as Tokens.Em).tokens)}</em>{:else if t.type === 'del'}<del
        >{@render inline((t as Tokens.Del).tokens)}</del
      >{:else if t.type === 'codespan'}<code>{t.text}</code>{:else if t.type === 'br'}<br />{:else if t.type === 'link'}<a
        href={(t as Tokens.Link).href}
        title={(t as Tokens.Link).href}
        onclick={(e) => open(e, (t as Tokens.Link).href)}>{@render inline((t as Tokens.Link).tokens)}</a
      >{:else if t.type === 'image'}<a href={(t as Tokens.Image).href} onclick={(e) => open(e, (t as Tokens.Image).href)}
        >{(t as Tokens.Image).text || (t as Tokens.Image).href}</a
      >{:else if t.type === 'checkbox'}<input type="checkbox" checked={(t as Tokens.Checkbox).checked} disabled />{:else}{t.raw}{/if}{/each}{/snippet}

{#snippet blocks(list: Token[])}
  {#each list as t, i (i)}
    {#if t.type === 'paragraph'}
      <p>{@render inline((t as Tokens.Paragraph).tokens)}</p>
    {:else if t.type === 'heading'}
      <h3>{@render inline((t as Tokens.Heading).tokens)}</h3>
    {:else if t.type === 'code'}
      <pre><code>{(t as Tokens.Code).text}</code></pre>
    {:else if t.type === 'blockquote'}
      <blockquote>{@render blocks((t as Tokens.Blockquote).tokens)}</blockquote>
    {:else if t.type === 'list'}
      {@const l = t as Tokens.List}
      {#if l.ordered}
        <ol start={Number(l.start) || 1}>
          {#each l.items as item, k (k)}<li>{@render blocks(item.tokens)}</li>{/each}
        </ol>
      {:else}
        <ul class:tasks={l.items.some((it) => it.task)}>
          {#each l.items as item, k (k)}<li>{@render blocks(item.tokens)}</li>{/each}
        </ul>
      {/if}
    {:else if t.type === 'table'}
      {@const tb = t as Tokens.Table}
      <div class="table">
        <table>
          <thead>
            <tr>{#each tb.header as cell, k (k)}<th>{@render inline(cell.tokens)}</th>{/each}</tr>
          </thead>
          <tbody>
            {#each tb.rows as row, r (r)}
              <tr>{#each row as cell, k (k)}<td>{@render inline(cell.tokens)}</td>{/each}</tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else if t.type === 'hr'}
      <hr />
    {:else if t.type === 'text' || t.type === 'checkbox'}
      {@render inline([t])}
    {:else if t.type === 'html'}
      <p>{t.raw}</p>
    {:else if t.type !== 'space' && t.type !== 'def'}
      <p>{t.raw}</p>
    {/if}
  {/each}
{/snippet}

<div class="markdown">{@render blocks(tokens)}</div>

<style>
  .markdown {
    font-size: 14.5px;
    line-height: 1.6;
    color: var(--ink-soft);
    overflow-wrap: anywhere;
  }

  .markdown > :global(:first-child) {
    margin-top: 0;
  }

  p,
  ul,
  ol,
  pre,
  blockquote,
  .table {
    margin: 0 0 10px;
  }

  h3 {
    margin: 14px 0 6px;
    font-size: 15px;
    font-weight: 600;
    color: var(--ink);
  }

  ul,
  ol {
    padding-left: 22px;
  }

  ul.tasks {
    list-style: none;
    padding-left: 4px;
  }

  li > :global(p) {
    margin: 0;
  }

  code {
    font-family: var(--font-mono);
    font-size: 12.5px;
    color: var(--accent-b);
    background: var(--paper-warm);
    padding: 1px 4px;
    border-radius: 3px;
  }

  pre {
    overflow-x: auto;
    padding: 10px 12px;
    background: var(--paper-warm);
    border-radius: 4px;
  }

  pre code {
    padding: 0;
    background: none;
    color: var(--ink);
    white-space: pre;
  }

  blockquote {
    padding-left: 12px;
    border-left: 3px solid var(--paper-edge);
    color: var(--ink-mute);
  }

  .table {
    overflow-x: auto;
  }

  table {
    border-collapse: collapse;
    font-size: 13px;
  }

  th,
  td {
    padding: 3px 8px;
    border: 1px solid var(--paper-edge);
    text-align: left;
  }

  hr {
    border: none;
    border-top: 1px dashed var(--hairline);
    margin: 12px 0;
  }

  a {
    color: var(--link);
  }

  a:hover {
    color: var(--link-hover);
  }

  input[type='checkbox'] {
    margin: 0 4px 0 0;
    vertical-align: -1px;
  }
</style>
