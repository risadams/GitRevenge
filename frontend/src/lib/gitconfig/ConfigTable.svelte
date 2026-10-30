<script lang="ts">
  // The Settings view of the Git Config tab (#225): the selected file's keys by section,
  // edited in place, with what other config files set for the same keys.
  import { tick } from 'svelte'
  import { AddGitConfigValue, GitConfigKeys, ReplaceGitConfigValue, SetGitConfig, UnsetGitConfig } from '../../../wailsjs/go/app/App.js'
  import { commands } from '../commands.svelte'
  import { session } from '../session.svelte'
  import Icon from '../ui/Icon.svelte'
  import { CATALOG, canonicalKey, keyInfo, splitKey } from './catalog'
  import { gitConfig } from './config.svelte'
  import { buildRows, filterSections, hasSecret, maskUrls, maskValue, overrideNote, samePath, type Inherited, type Row } from './rows'
  import ValueEditor from './ValueEditor.svelte'

  let filter = $state('')
  let showInherited = $state(true)
  let collapsed = $state<Record<string, boolean>>({})
  /** Rows whose secrets are shown, by canonical key. Forgotten when the view unmounts. */
  let revealed = $state<Record<string, boolean>>({})
  /** The row taking a new value, by canonical key. */
  let adding = $state('')
  let newValue = $state('')
  let newKey = $state('')
  let newKeyValue = $state('')
  let known = $state<string[]>([])

  let scope = $derived(gitConfig.scope)
  let doc = $derived(gitConfig.doc)
  let sections = $derived(doc ? buildRows(doc.entries, gitConfig.effective, doc.path) : [])
  let shown = $derived(filterSections(sections, filter, showInherited))
  let suggestions = $derived([...new Set([...CATALOG.map((c) => c.key), ...known])])
  let newInfo = $derived(keyInfo(newKey))

  $effect(() => {
    GitConfigKeys()
      .then((k) => (known = k ?? []))
      .catch(() => {})
  })

  const edit = (fn: (path: string) => Promise<unknown>) => gitConfig.apply(fn as never)

  function setValue(r: Row, old: string, value: string) {
    if (r.multi) return edit((p) => ReplaceGitConfigValue(p, r.key, old, value))
    return edit((p) => SetGitConfig(p, r.key, value))
  }

  function remove(r: Row, value = '') {
    return edit((p) => UnsetGitConfig(p, r.key, value))
  }

  /** Copies the inherited value into this file, where it can be changed. */
  function override(r: Row) {
    const last = r.inherited[r.inherited.length - 1]
    const value = last.implicit ? 'true' : last.value
    return edit((p) => (r.multi ? AddGitConfigValue(p, r.key, value) : SetGitConfig(p, r.key, value)))
  }

  async function addValue(r: Row) {
    if (!newValue) return
    if (await edit((p) => AddGitConfigValue(p, r.key, newValue))) {
      adding = ''
      newValue = ''
    }
  }

  async function addKey(e: SubmitEvent) {
    e.preventDefault()
    const key = newKey.trim()
    if (!key) return
    if (key.includes('<')) {
      gitConfig.error = `Replace the placeholder in ${key} with a name, such as remote.origin.url.`
      return
    }
    const exists = sections.some((s) => s.rows.some((r) => r.values.length > 0 && canonicalKey(r.key) === canonicalKey(key)))
    const multi = keyInfo(key)?.multi
    if (await edit((p) => (exists && multi ? AddGitConfigValue(p, key, newKeyValue) : SetGitConfig(p, key, newKeyValue)))) {
      newKey = ''
      newKeyValue = ''
    }
  }

  /** A file that inherited values come from and that isn't already in the switcher. */
  function editableSource(i: Inherited): boolean {
    return i.scope !== 'system' && i.file !== '' && !gitConfig.scopes.some((s) => samePath(s.path, i.file))
  }

  async function manageRemotes(repo: string) {
    await session.select(repo)
    await tick()
    commands.all().find((c) => c.id === 'repo.remotes')?.run()
  }

  const shownValue = (r: Row, v: string) => (revealed[canonicalKey(r.key)] ? v : maskValue(r.key, v))
  const shownKey = (r: Row) => (revealed[canonicalKey(r.key)] ? r.key : maskUrls(r.key))
  const secret = (r: Row) => [...r.values.map((v) => v.value), ...r.inherited.map((i) => i.value)].some((v) => hasSecret(r.key, v)) || hasSecret(r.key, '')
  const name = (key: string) => splitKey(key)[2]
</script>

<div class="table">
  <div class="toolbar">
    <label class="filter">
      <Icon name="search" size={14} />
      <input type="search" placeholder="Filter keys and values" aria-label="Filter keys and values" bind:value={filter} spellcheck="false" />
    </label>
    <label class="form-check"><input type="checkbox" bind:checked={showInherited} /> Show inherited</label>
  </div>

  <form class="add" onsubmit={addKey}>
    <input list="gitconfig-keys" placeholder="Add a key, e.g. pull.rebase" aria-label="New key" bind:value={newKey} spellcheck="false" autocomplete="off" />
    <datalist id="gitconfig-keys">
      {#each suggestions as s (s)}<option value={s}>{keyInfo(s)?.description ?? ''}</option>{/each}
    </datalist>
    {#if newInfo?.values}
      <select aria-label="New value" bind:value={newKeyValue}>
        <option value="" disabled>Value</option>
        {#each newInfo.values as v (v)}<option value={v}>{v}</option>{/each}
      </select>
    {:else if newInfo?.type === 'bool'}
      <select aria-label="New value" bind:value={newKeyValue}>
        <option value="" disabled>Value</option>
        <option value="true">true</option>
        <option value="false">false</option>
      </select>
    {:else}
      <input placeholder="Value" aria-label="New value" bind:value={newKeyValue} spellcheck="false" autocomplete="off" />
    {/if}
    <button class="tool" type="submit" disabled={!newKey.trim() || gitConfig.busy}><Icon name="plus" size={14} />Add</button>
    {#if newInfo}<span class="sub hint">{newInfo.description}</span>{/if}
  </form>

  {#each shown as s (s.name)}
    {@const open = !collapsed[s.name]}
    <section class="section">
      <h3>
        <button class="toggle" aria-expanded={open} onclick={() => (collapsed[s.name] = open)}>
          <Icon name={open ? 'chevron' : 'right'} size={14} />
          <code>[{maskUrls(s.name)}]</code>
        </button>
        {#if scope?.kind === 'local' && s.name.startsWith('remote "')}
          <button class="link" onclick={() => manageRemotes(scope.repo)}>Also managed in Remotes</button>
        {:else if scope?.kind === 'local' && s.name.startsWith('branch "')}
          <span class="sub">Upstreams are also set from the branch menu.</span>
        {/if}
      </h3>
      {#if open}
        {#each s.rows as r (canonicalKey(r.key))}
          {@const info = keyInfo(r.key)}
          {@const id = canonicalKey(r.key)}
          {@const own = r.values.length > 0}
          {@const hidden = secret(r) && !revealed[id]}
          <div class="row" class:inherited={!own}>
            <div class="key" title={info?.description ?? ''}>
              <code>{name(r.key)}</code>
              {#if r.key.split('.').length > 2}<span class="sub full">{shownKey(r)}</span>{/if}
              {#if info}<span class="sub desc">{info.description}</span>{/if}
            </div>
            <div class="values">
              {#if own}
                {#each r.values as v, i (i)}
                  <div class="value">
                    {#if hidden}
                      <code class="masked">{shownValue(r, v.value)}</code>
                    {:else}
                      <ValueEditor label={r.key} value={v.value} implicit={v.implicit} {info} disabled={gitConfig.busy} onCommit={(nv) => setValue(r, v.value, nv)} />
                    {/if}
                    {#if r.multi}
                      <button class="tool icon" title="Remove this value" aria-label="Remove {v.value} from {r.key}" onclick={() => remove(r, v.value)}><Icon name="minus" size={14} /></button>
                    {/if}
                  </div>
                {/each}
                {#if r.multi}
                  {#if adding === id}
                    <form class="value" onsubmit={(e) => (e.preventDefault(), addValue(r))}>
                      <input aria-label="Another value for {r.key}" bind:value={newValue} spellcheck="false" autocomplete="off" />
                      <button class="tool" type="submit" disabled={!newValue}>Add</button>
                      <button class="tool" type="button" onclick={() => (adding = '')}>Cancel</button>
                    </form>
                  {:else}
                    <button class="tool add-value" onclick={() => ((adding = id), (newValue = ''))}><Icon name="plus" size={14} />Add value</button>
                  {/if}
                {/if}
                {#if overrideNote(r)}<p class="sub note">{overrideNote(r)}</p>{/if}
              {/if}
              {#if !own}
                {#each r.inherited as i, n (n)}
                  <div class="value from">
                    <code>{i.implicit ? 'true' : shownValue(r, i.value)}</code>
                    <span class="view-badge" title={i.file || 'set on the command line'}>{i.scope}</span>
                    {#if editableSource(i)}
                      <button class="link" title={i.file} onclick={() => gitConfig.showFile(i.file, scope?.repo ?? '')}>Edit this file</button>
                    {/if}
                  </div>
                {/each}
              {/if}
            </div>
            <div class="actions">
              {#if secret(r)}
                <button
                  class="tool icon"
                  aria-pressed={!!revealed[id]}
                  title={revealed[id] ? 'Hide' : 'Show'}
                  aria-label="{revealed[id] ? 'Hide' : 'Show'} the value of {r.key}"
                  onclick={() => (revealed[id] = !revealed[id])}
                >
                  <Icon name="eye" size={14} />
                </button>
              {/if}
              {#if own}
                <button class="tool" onclick={() => remove(r)} title={r.inherited.length ? 'Remove from this file, so the inherited value applies' : 'Remove from this file'}>
                  {r.inherited.length ? 'Revert to inherited' : 'Remove'}
                </button>
              {:else}
                <button class="tool" onclick={() => override(r)}>{r.multi ? 'Add here' : 'Override here'}</button>
              {/if}
            </div>
          </div>
        {/each}
      {/if}
    </section>
  {:else}
    <p class="sub empty">{filter ? 'Nothing matches the filter.' : 'This file has no settings yet. Add one above.'}</p>
  {/each}
</div>

<style>
  .table {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .toolbar,
  .add {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .filter {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    max-width: 360px;
    color: var(--ink-mute);
  }

  .filter input {
    flex: 1;
  }

  .add input {
    font-family: var(--font-mono);
    font-size: 12px;
  }

  .add input[list] {
    width: 260px;
  }

  .hint {
    flex-basis: 100%;
  }

  .section h3 {
    display: flex;
    align-items: center;
    gap: 12px;
    margin: 0;
    font-size: 14px;
    font-weight: 400;
  }

  .toggle {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 2px 4px;
    background: none;
    border: none;
  }

  .toggle code {
    font-family: var(--font-mono);
    color: var(--accent-a);
  }

  .link {
    padding: 0;
    background: none;
    border: none;
    color: var(--link);
    font-size: 12px;
    text-decoration: underline;
  }

  .row {
    display: grid;
    grid-template-columns: minmax(160px, 1fr) minmax(200px, 2fr) 170px;
    gap: 12px;
    align-items: start;
    padding: 6px 4px 6px 24px;
    border-bottom: 1px solid var(--hairline);
  }

  .row.inherited {
    color: var(--ink-mute);
  }

  .key {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .key code {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--accent-b);
  }

  .row.inherited .key code {
    color: var(--ink-mute);
  }

  .sub {
    color: var(--ink-mute);
    font-size: 12px;
    line-height: 1.4;
  }

  .full {
    font-family: var(--font-mono);
    font-size: 11px;
    overflow-wrap: anywhere;
  }

  .values {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .value {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  .value code {
    font-family: var(--font-mono);
    font-size: 12px;
    overflow-wrap: anywhere;
  }

  .value input {
    flex: 1;
    font-family: var(--font-mono);
    font-size: 12px;
  }

  .note {
    margin: 0;
  }

  .add-value {
    align-self: flex-start;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 4px;
  }

  .icon {
    padding: 2px 6px;
  }

  .empty {
    padding: 12px 0;
  }
</style>
