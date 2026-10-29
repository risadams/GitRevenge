<script lang="ts">
  import { CommitDetail, IssueURL } from '../../wailsjs/go/app/App.js'
  import type { app, git } from '../../wailsjs/go/models'
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js'
  import { formatTime } from './graph/draw'
  import type { DiffTarget } from './diff/DiffView.svelte'
  import { session } from './session.svelte'
  import { reflow } from './text'
  import { asPerson, valueParts } from './trailers'

  let {
    path,
    commitId,
    openFile,
    current,
  }: {
    path: string
    commitId: string
    openFile: (t: DiffTarget) => void
    /** The diff currently open, to highlight its file. */
    current?: DiffTarget | null
  } = $props()

  let info = $state<app.CommitInfo | null>(null)
  /** Where "#123" in a trailer links to, from the repository's remote; "" for nowhere. */
  let issueURL = $state('')

  $effect(() => {
    const p = path
    issueURL = ''
    void IssueURL(p)
      .then((u) => {
        if (p === path) issueURL = u
      })
      .catch(() => {})
  })

  function openLink(e: MouseEvent, href: string) {
    e.preventDefault()
    BrowserOpenURL(href)
  }

  const statusNames: Record<string, string> = {
    A: 'Added',
    M: 'Modified',
    D: 'Deleted',
    R: 'Renamed',
    C: 'Copied',
    T: 'Type changed',
  }

  // Load whenever the selected commit changes; drop answers for a commit no longer selected.
  $effect(() => {
    const id = commitId
    info = null
    void session.run(() => CommitDetail(path, id)).then((res) => {
      if (res && id === commitId) info = res
    })
  })

  function showFile(f: git.FileChange) {
    openFile({ mode: 'commit', commit: commitId, file: f.path, oldPath: f.oldPath })
  }
</script>

<aside class="panel" aria-label="Commit details">
  <span class="blob" aria-hidden="true"></span>
  {#if !info}
    <p class="muted pad">Loading…</p>
  {:else}
    {@const c = info.commit}
    <header>
      <span class="hand">this commit</span>
      <h2>{c.subject}</h2>
      {#if c.body}<p class="body">{reflow(c.body)}</p>{/if}
      {#if c.trailers.length}
        <dl class="trailers" aria-label="Trailers">
          {#each c.trailers as t, i (i)}
            {@const who = asPerson(t.value)}
            <dt class="meta">{t.key}</dt>
            <dd>
              {#if who}
                {who.name}<span class="email" title={who.email}>{who.email}</span>
              {:else}
                {#each valueParts(t.value, issueURL) as part, k (k)}{#if part.href}<a
                      href={part.href}
                      title={part.href}
                      onclick={(e) => openLink(e, part.href!)}>{part.text}</a
                    >{:else}{part.text}{/if}{/each}
              {/if}
            </dd>
          {/each}
        </dl>
      {/if}
      <dl>
        <dt class="meta">Author</dt>
        <dd>
          {c.author.name}
          <span class="email" title={c.author.email}>{c.author.email}</span>
          <span class="when">{formatTime(c.author.time)}</span>
        </dd>
        {#if c.committer.name !== c.author.name || c.committer.time !== c.author.time}
          <dt class="meta">Committer</dt>
          <dd>
          {c.committer.name}
          <span class="when">{formatTime(c.committer.time)}</span>
        </dd>
        {/if}
        <dt class="meta">Commit</dt>
        <dd><code>{c.id}</code></dd>
        {#if c.parents.length}
          <dt class="meta">{c.parents.length > 1 ? 'Parents' : 'Parent'}</dt>
          <dd class="parents">{c.parents.map((p) => p.slice(0, 10)).join(' ')}</dd>
        {/if}
      </dl>
      {#if c.parents.length > 1}
        <p class="note">Changes are shown against the first parent.</p>
      {/if}
    </header>

    <div class="files-head">
      <span class="meta label">Files changed</span>
      <span class="meta">{info.files.length}</span>
    </div>
    <ul class="files">
      {#each info.files as f (f.path)}
        <li>
          <button
            class:active={current?.mode === 'commit' && current.commit === commitId && current.file === f.path}
            onclick={() => showFile(f)}
            title={f.oldPath ? `${f.oldPath} → ${f.path}` : f.path}
          >
            <span class="status s-{f.status}">{statusNames[f.status] ?? f.status}</span>
            <span class="file">{f.oldPath ? `${f.oldPath} → ${f.path}` : f.path}</span>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</aside>

<style>
  .panel {
    position: relative;
    display: flex;
    flex-direction: column;
    width: 420px;
    min-width: 300px;
    padding-bottom: 16px;
    border-left: 1px solid var(--paper-edge);
    background: var(--paper);
    overflow-x: hidden;
    overflow-y: auto;
  }

  /* One watercolour blob behind the title: decoration only. */
  .blob {
    position: absolute;
    top: -50px;
    right: -40px;
    width: 220px;
    height: 170px;
    border-radius: 45% 55% 50% 50% / 55% 45%;
    background: var(--ink-rose);
    filter: blur(60px);
    opacity: 0.16;
    pointer-events: none;
  }

  :global([data-theme='dark']) .blob {
    opacity: 0.12;
  }

  header {
    position: relative;
    padding: 18px 20px 4px;
  }

  h2 {
    margin: 4px 0 8px;
    font-size: 21px;
    font-weight: 500;
    line-height: 1.25;
    letter-spacing: -0.015em;
    overflow-wrap: anywhere;
    font-variation-settings:
      'SOFT' 50,
      'WONK' 1;
  }

  .body {
    margin: 0 0 12px;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    font-size: 14.5px;
    line-height: 1.6;
    color: var(--ink-soft);
  }

  dl {
    display: grid;
    grid-template-columns: 70px minmax(0, 1fr);
    gap: 8px 12px;
    align-items: baseline;
    margin: 12px 0 8px;
  }

  dd {
    margin: 0;
    overflow-wrap: anywhere;
    font-size: 14px;
  }

  /* Trailers read as a short record under the message: key, then person or value. */
  dl.trailers {
    grid-template-columns: max-content minmax(0, 1fr);
    margin: 0 0 12px;
    padding-top: 10px;
    border-top: 1px dashed var(--hairline);
  }

  .trailers dt {
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .trailers a {
    color: var(--link);
  }

  .trailers a:hover {
    color: var(--link-hover);
  }

  .email {
    display: block;
    font-size: 13px;
    color: var(--ink-mute);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .when {
    display: block;
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--ink-mute);
  }

  code {
    padding: 1px 4px;
    border-radius: 3px;
    background: var(--paper-warm);
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--accent-b);
    overflow-wrap: anywhere;
  }

  .parents {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--accent-b);
  }

  .note {
    margin: 0 0 8px;
    font-size: 13px;
    font-style: italic;
    color: var(--ink-mute);
  }

  .files-head {
    position: relative;
    display: flex;
    gap: 10px;
    align-items: baseline;
    margin: 12px 20px 0;
    padding-bottom: 6px;
    border-bottom: 1px solid var(--paper-edge);
  }

  .label {
    color: var(--ink-soft);
  }

  .files {
    list-style: none;
    margin: 0 8px;
    padding: 0;
  }

  .files li {
    border-bottom: 1px dashed var(--hairline);
  }

  .files button {
    display: grid;
    grid-template-columns: 78px minmax(0, 1fr);
    align-items: baseline;
    gap: 10px;
    width: 100%;
    padding: 7px 12px;
    background: none;
    border: none;
    border-radius: 4px;
    text-align: left;
    transition:
      background-color 0.2s,
      transform 0.3s var(--ease-ink);
  }

  .files button:hover:not(:disabled) {
    background: var(--wash-hover);
    transform: translateX(3px);
  }

  .files button.active {
    background: var(--selected);
  }

  .status {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .s-A {
    color: var(--success);
  }

  .s-D {
    color: var(--danger);
  }

  .s-M,
  .s-T {
    color: var(--accent-c);
  }

  .s-R,
  .s-C {
    color: var(--accent-a);
  }

  .file {
    font-family: var(--font-mono);
    font-size: 12.5px;
    overflow-wrap: anywhere;
  }

  .muted {
    color: var(--ink-mute);
  }

  .pad {
    padding: 16px 20px;
    font-style: italic;
  }

  @media (prefers-reduced-motion: reduce) {
    .files button {
      transition: none;
    }

    .files button:hover:not(:disabled) {
      transform: none;
    }
  }
</style>
