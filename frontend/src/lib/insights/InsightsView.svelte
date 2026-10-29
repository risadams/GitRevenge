<script lang="ts">
  // Insights ("Scrum Master view"): what happened in a stretch of history. Commit activity,
  // contributors, churn hotspots, merge cycle time and branch ages, each exportable as CSV,
  // the activity chart also as PNG. Aggregation runs in Go over one streamed git log.
  import { BranchAges, Refs, RepoStats, SaveFile } from '../../../wailsjs/go/app/App.js'
  import type { git } from '../../../wailsjs/go/models'
  import { errorText } from '../errors'
  import type { Ref } from '../graph/rowCache'
  import { session } from '../session.svelte'
  import { theme } from '../theme.svelte'
  import Icon from '../ui/Icon.svelte'
  import BarChart, { type Colors } from './BarChart.svelte'
  import { formatHours, svgToPngBase64, textToBase64, toCSV } from './export'

  let { path, onClose }: { path: string; onClose: () => void } = $props()

  const DAY = 24 * 60 * 60
  const PRESETS = [
    { id: '7', label: 'Last 7 days', days: 7 },
    { id: '30', label: 'Last 30 days', days: 30 },
    { id: '90', label: 'Last 90 days', days: 90 },
    { id: '365', label: 'Last year', days: 365 },
    { id: 'custom', label: 'Custom…', days: 0 },
  ]

  let preset = $state('30')
  const isoDay = (unix: number) => new Date(unix * 1000).toISOString().slice(0, 10)
  let fromDay = $state(isoDay(Date.now() / 1000 - 30 * DAY))
  let toDay = $state(isoDay(Date.now() / 1000))
  let rev = $state('')
  let branches = $state<string[]>([])
  let stats = $state<git.Stats | null>(null)
  let ages = $state<git.BranchAge[]>([])
  let loading = $state(false)
  let error = $state('')
  let showTables = $state(false)
  let activitySvg = $state<SVGSVGElement>()

  let range = $derived.by(() => {
    const now = Math.floor(Date.now() / 1000)
    const p = PRESETS.find((x) => x.id === preset)!
    if (p.days) return { since: now - p.days * DAY, until: now }
    return { since: Math.floor(Date.parse(fromDay) / 1000), until: Math.floor(Date.parse(toDay) / 1000) + DAY - 1 }
  })

  // Chart colours as values: resolved from the theme so an exported PNG matches.
  let colors = $derived.by((): Colors => {
    void theme.resolved
    const css = getComputedStyle(document.documentElement)
    const token = (n: string) => css.getPropertyValue(n).trim()
    return { mark: token('--accent'), text: token('--ink'), muted: token('--ink-mute'), grid: token('--hairline') }
  })

  $effect(() => {
    void Refs(path)
      .then((all) => (branches = ((all ?? []) as Ref[]).filter((r) => r.kind === 'local').map((r) => r.short)))
      .catch(() => {})
    void BranchAges(path)
      .then((a) => (ages = a ?? []))
      .catch(() => {})
  })

  let seq = 0
  $effect(() => {
    const { since, until } = range
    const r = rev
    if (!Number.isFinite(since) || !Number.isFinite(until)) return
    const mine = ++seq
    loading = true
    error = ''
    RepoStats(path, since, until, r)
      .then((s) => {
        if (mine === seq) stats = s
      })
      .catch((e) => {
        if (mine === seq) error = errorText(e)
      })
      .finally(() => {
        if (mine === seq) loading = false
      })
  })

  const dayFmt = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' })
  let activityBars = $derived(
    (stats?.activity ?? []).map((b) => ({
      label: dayFmt.format(new Date(b.start * 1000)),
      value: b.commits,
      detail: `${stats?.bucket === 'week' ? 'Week of ' : ''}${dayFmt.format(new Date(b.start * 1000))}: ${b.commits.toLocaleString()} commit${b.commits === 1 ? '' : 's'}`,
    })),
  )
  let contributorBars = $derived((stats?.contributors ?? []).slice(0, 10).map((c) => ({ label: c.name, value: c.commits })))
  let hotspotBars = $derived((stats?.hotspots ?? []).slice(0, 15).map((h) => ({ label: h.path, value: h.changed, detail: `${h.path}: ${h.changed.toLocaleString()} lines changed in ${h.commits} commits` })))

  const now = Date.now() / 1000
  const ageDays = (t: number) => Math.floor((now - t) / DAY)

  async function save(name: string, base64: string) {
    await session.run(() => SaveFile('Export', name, base64))
  }

  function exportCSV(which: 'activity' | 'contributors' | 'hotspots' | 'branches') {
    if (!stats && which !== 'branches') return
    const s = stats!
    const rows: (string | number)[][] =
      which === 'activity'
        ? [[s.bucket === 'week' ? 'week starting' : 'day', 'commits'], ...s.activity.map((b) => [isoDay(b.start), b.commits])]
        : which === 'contributors'
          ? [['name', 'email', 'commits', 'lines added', 'lines deleted'], ...s.contributors.map((c) => [c.name, c.email, c.commits, c.added, c.deleted])]
          : which === 'hotspots'
            ? [['path', 'commits', 'lines changed'], ...s.hotspots.map((h) => [h.path, h.commits, h.changed])]
            : [['branch', 'kind', 'last commit', 'days since', 'author'], ...ages.map((b) => [b.name, b.kind, isoDay(b.last), ageDays(b.last), b.author])]
    void save(`${which}.csv`, textToBase64(toCSV(rows)))
  }

  async function exportPNG() {
    if (!activitySvg) return
    try {
      await save('activity.png', await svgToPngBase64(activitySvg, getComputedStyle(document.documentElement).getPropertyValue('--paper').trim()))
    } catch (e) {
      session.error = errorText(e)
    }
  }
</script>

<section class="insights" aria-label="Insights">
  <header class="view-head">
    <button class="tool" onclick={onClose} title="Back to the graph"><Icon name="back" />Back</button>
    <span class="view-badge">Insights</span>
    <span class="spacer"></span>
  </header>

  <div class="filters" role="group" aria-label="Filters">
    <label class="form-field">
      <span>Range</span>
      <select bind:value={preset}>{#each PRESETS as p (p.id)}<option value={p.id}>{p.label}</option>{/each}</select>
    </label>
    {#if preset === 'custom'}
      <label class="form-field"><span>From</span><input type="date" bind:value={fromDay} max={toDay} /></label>
      <label class="form-field"><span>To</span><input type="date" bind:value={toDay} min={fromDay} /></label>
    {/if}
    <label class="form-field">
      <span>Branch</span>
      <select bind:value={rev}>
        <option value="">All branches</option>
        {#each branches as b (b)}<option value={b}>{b}</option>{/each}
      </select>
    </label>
    <label class="check"><input type="checkbox" bind:checked={showTables} /> Show tables</label>
    {#if loading}<span class="meta" role="status">Reading history…</span>{/if}
  </div>

  {#if error}
    <p class="error" role="alert">{error}</p>
  {:else if stats}
    <div class="tiles">
      <div class="tile"><span class="meta">Commits</span><span class="big">{stats.commits.toLocaleString()}</span></div>
      <div class="tile"><span class="meta">Contributors</span><span class="big">{stats.contributors.length.toLocaleString()}{stats.contributors.length === 50 ? '+' : ''}</span></div>
      <div class="tile"><span class="meta">Merges</span><span class="big">{stats.merges.toLocaleString()}</span></div>
      <div class="tile" title="From a merged branch's first commit to its merge, over the {stats.cycle.sample} most recent merges">
        <span class="meta">Merge cycle time</span>
        <span class="big">{stats.cycle.sample ? formatHours(stats.cycle.median) : '—'}</span>
        {#if stats.cycle.sample}<span class="sub">median · p90 {formatHours(stats.cycle.p90)} · {stats.cycle.sample} merges</span>{/if}
      </div>
    </div>

    <section class="card">
      <div class="card-head">
        <h3>Commit activity <span class="meta">per {stats.bucket}</span></h3>
        <button class="tool" onclick={() => exportCSV('activity')}>CSV</button>
        <button class="tool" onclick={exportPNG}>PNG</button>
      </div>
      <BarChart bars={activityBars} {colors} label="Commits per {stats.bucket}" bind:svg={activitySvg} />
      {#if showTables}
        <table>
          <thead><tr><th>{stats.bucket === 'week' ? 'Week starting' : 'Day'}</th><th>Commits</th></tr></thead>
          <tbody>{#each stats.activity as b (b.start)}<tr><td>{isoDay(b.start)}</td><td>{b.commits}</td></tr>{/each}</tbody>
        </table>
      {/if}
    </section>

    <div class="pair">
      <section class="card">
        <div class="card-head">
          <h3>Contributors <span class="meta">by commits</span></h3>
          <button class="tool" onclick={() => exportCSV('contributors')}>CSV</button>
        </div>
        {#if contributorBars.length}<BarChart horizontal bars={contributorBars} {colors} label="Commits by contributor" />{:else}<p class="muted">No commits in this range.</p>{/if}
        {#if showTables}
          <table>
            <thead><tr><th>Name</th><th>Commits</th><th>Added</th><th>Deleted</th></tr></thead>
            <tbody>{#each stats.contributors as c (c.email)}<tr><td>{c.name}</td><td>{c.commits}</td><td>+{c.added}</td><td>−{c.deleted}</td></tr>{/each}</tbody>
          </table>
        {/if}
      </section>

      <section class="card">
        <div class="card-head">
          <h3>Churn hotspots <span class="meta">lines changed</span></h3>
          <button class="tool" onclick={() => exportCSV('hotspots')}>CSV</button>
        </div>
        {#if hotspotBars.length}<BarChart horizontal bars={hotspotBars} {colors} label="Lines changed by file" />{:else}<p class="muted">No changes in this range.</p>{/if}
        {#if showTables}
          <table>
            <thead><tr><th>File</th><th>Commits</th><th>Lines changed</th></tr></thead>
            <tbody>{#each stats.hotspots as h (h.path)}<tr><td class="path">{h.path}</td><td>{h.commits}</td><td>{h.changed}</td></tr>{/each}</tbody>
          </table>
        {/if}
      </section>
    </div>
  {/if}

  <section class="card">
    <div class="card-head">
      <h3>Branch age <span class="meta">oldest first</span></h3>
      <button class="tool" onclick={() => exportCSV('branches')}>CSV</button>
    </div>
    <table>
      <thead><tr><th>Branch</th><th>Last commit</th><th>Age</th><th>By</th></tr></thead>
      <tbody>
        {#each ages as b (b.kind + b.name)}
          <tr>
            <td>{b.name}{#if b.kind === 'remote'}<span class="meta kind"> remote</span>{/if}</td>
            <td>{isoDay(b.last)}</td>
            <td>{ageDays(b.last)} days{#if ageDays(b.last) > 90}<span class="stale"><Icon name="warning" size={12} /> stale</span>{/if}</td>
            <td>{b.author}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </section>
</section>

<style>
  .insights {
    display: flex;
    flex-direction: column;
    gap: 14px;
    height: 100%;
    min-height: 0;
    overflow: auto;
    padding: 0 16px 24px;
  }

  .spacer {
    flex: 1;
  }

  .filters {
    display: flex;
    flex-wrap: wrap;
    align-items: end;
    gap: 10px 16px;
  }

  .check {
    font-size: 13px;
    padding-bottom: 6px;
  }

  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 10px;
  }

  .tile {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 12px 14px;
    background: var(--paper-warm);
    border-radius: 6px;
  }

  .big {
    font-size: 28px;
    font-weight: 500;
    letter-spacing: -0.01em;
  }

  .sub {
    font-size: 12px;
    color: var(--ink-mute);
  }

  .pair {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
    gap: 14px;
  }

  .card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
  }

  .card-head {
    display: flex;
    align-items: baseline;
    gap: 4px;
    border-bottom: 1px dashed var(--hairline);
  }

  h3 {
    flex: 1;
    margin: 0 0 4px;
    font-size: 17px;
    font-weight: 500;
  }

  .card-head .tool {
    height: 24px;
    font-size: 12.5px;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }

  th {
    text-align: left;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 400;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--ink-mute);
    padding: 4px 6px;
  }

  td {
    padding: 3px 6px;
    border-top: 1px dashed var(--hairline);
  }

  .path {
    font-family: var(--font-mono);
    font-size: 12px;
    overflow-wrap: anywhere;
  }

  .stale {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    margin-left: 8px;
    color: var(--warning);
  }

  .muted {
    color: var(--ink-mute);
    font-style: italic;
  }

  .error {
    color: var(--danger);
  }
</style>
