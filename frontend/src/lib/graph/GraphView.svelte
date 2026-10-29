<script lang="ts">
  import { onMount } from 'svelte'
  import { CommitRow, GraphRows, SearchGraph } from '../../../wailsjs/go/app/App.js'
  import { session } from '../session.svelte'
  import { EventsOn } from '../../../wailsjs/runtime/runtime.js'
  import { theme as colorTheme } from '../theme.svelte'
  import { drawRow, graphWidth, LANE_COUNT, layoutColumns, ROW_HEIGHT, setLaneColors, type Theme } from './draw'
  import { RowCache, type Page, type Ref, type Row } from './rowCache'

  let {
    path,
    onSelect,
    onContext,
    onDrop,
  }: {
    path: string
    /** Right-click on a row, with the refs pointing at it and the mouse position. */
    onContext?: (row: Row, refs: Ref[], x: number, y: number) => void
    /** A row dragged onto another row. */
    onDrop?: (source: Row, sourceRefs: Ref[], target: Row, targetRefs: Ref[], x: number, y: number) => void
    /** auto is true for the automatic selection of HEAD when the graph first loads. */
    onSelect?: (row: Row | undefined, auto: boolean) => void
  } = $props()

  /** Rows fetched beyond the viewport, so short scrolls never show blanks. */
  const OVERSCAN = 40
  /** The graph column never takes more than this share of the width. */
  const MAX_GRAPH_SHARE = 0.4

  let scroller: HTMLDivElement
  let canvas: HTMLCanvasElement
  let total = $state(0)
  let loading = $state(true)
  let error = $state('')
  let selected = $state(-1)
  let hovered = -1
  let frame = 0
  /** Drag state: the row pressed and where; the row under the pointer while dragging. */
  let drag: { from: number; x: number; y: number; active: boolean } | null = null
  let dropTarget = -1

  // Search state. matchSet is only used for dimming when every match fits in the result.
  let searchInput: HTMLInputElement
  let query = $state('')
  let matches = $state<number[]>([])
  let matchTotal = $state(0)
  let matchSet: Set<number> | null = null
  let searchSeq = 0

  const cache = new RowCache(
    // The generated binding types mirror rowCache's; the cast only bridges class vs interface.
    async (start, count) => (await GraphRows(path, start, count)) as unknown as Page,
    () => {
      total = cache.total
      loading = cache.loading
      selectHeadOnce()
      schedule()
    },
  )

  function theme(): Theme {
    const css = getComputedStyle(document.documentElement)
    const token = (name: string) => css.getPropertyValue(name).trim()
    setLaneColors(Array.from({ length: LANE_COUNT }, (_, i) => token(`--lane-${i}`)))
    return {
      bg: token('--paper'),
      fg: token('--ink'),
      muted: token('--ink-mute'),
      selected: token('--selected'),
      hover: token('--wash-hover'),
    }
  }

  function schedule() {
    if (!frame) frame = requestAnimationFrame(draw)
  }

  function draw() {
    frame = 0
    if (!canvas || !scroller) return
    const dpr = window.devicePixelRatio || 1
    const w = scroller.clientWidth
    const h = scroller.clientHeight
    if (canvas.width !== Math.round(w * dpr) || canvas.height !== Math.round(h * dpr)) {
      canvas.width = Math.round(w * dpr)
      canvas.height = Math.round(h * dpr)
      canvas.style.width = `${w}px`
      canvas.style.height = `${h}px`
    }
    const ctx = canvas.getContext('2d')!
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, w, h)
    ctx.font = '13px system-ui, -apple-system, "Segoe UI", sans-serif'

    const top = scroller.scrollTop
    const first = Math.floor(top / ROW_HEIGHT)
    const last = Math.min(total, Math.ceil((top + h) / ROW_HEIGHT))
    cache.ensure(first - OVERSCAN, last + OVERSCAN)

    let lanes = 1
    for (let i = first; i < last; i++) lanes = Math.max(lanes, cache.get(i)?.width ?? 0)
    const lay = layoutColumns(w, graphWidth(lanes, w * MAX_GRAPH_SHARE))
    const th = theme()
    for (let i = first; i < last; i++) {
      const row = cache.get(i)
      const y = i * ROW_HEIGHT - top
      // The drop target is highlighted like a selection while dragging.
      if (row)
        drawRow(ctx, row, cache.refsFor(row.id), y, lay, th, i === selected || i === dropTarget, i === hovered, !!matchSet && !matchSet.has(i))
    }
  }

  function rowAt(e: MouseEvent): number {
    const rect = scroller.getBoundingClientRect()
    const i = Math.floor((e.clientY - rect.top + scroller.scrollTop) / ROW_HEIGHT)
    return i >= 0 && i < total ? i : -1
  }

  /** Scrolls to and selects row i (e.g. when a branch is picked in the sidebar). */
  export function reveal(i: number) {
    if (i < 0) return
    // Center it rather than pinning it to an edge.
    scroller.scrollTop = Math.max(0, i * ROW_HEIGHT - scroller.clientHeight / 2)
    select(i)
    scroller.focus()
  }

  /**
   * Initial selection: the top row until the checked-out commit is known, then that commit.
   * Refs load alongside history, so HEAD's label can arrive after the first rows; until the
   * user picks something themselves, keep trying to move to HEAD.
   */
  let autoSelect: 'none' | 'fallback' | 'done' = 'none'

  function selectHeadOnce() {
    if (autoSelect === 'done' || !cache.get(0)) return
    for (let i = 0; i < 200 && cache.get(i); i++) {
      if (cache.refsFor(cache.get(i)!.id).some((r) => r.head)) {
        autoSelect = 'done'
        if (i !== selected) select(i, true)
        return
      }
    }
    if (autoSelect === 'none') {
      autoSelect = 'fallback'
      select(0, true)
    }
  }

  /** The commit selected before a reload, so the selection follows it to its new row. */
  let selectedId = ''

  async function followSelection() {
    if (selected < 0) return
    const id = selectedId
    const row = id ? await CommitRow(path, id).catch(() => -1) : -1
    // If the commit is gone (e.g. amended), stay on the same position but refresh details.
    const target = row >= 0 ? row : selected
    cache.ensure(target, target + 1)
    // Wait for the row to load, giving up after ~5s (e.g. the graph got shorter).
    let frames = 300
    const wait = () => {
      if (cache.get(target)) select(target, true)
      else if (frames-- > 0) requestAnimationFrame(wait)
    }
    wait()
  }

  function select(i: number, auto = false) {
    if (!auto) autoSelect = 'done'
    selected = Math.max(0, Math.min(total - 1, i))
    // Keep the selection in view.
    const y = selected * ROW_HEIGHT
    if (y < scroller.scrollTop) scroller.scrollTop = y
    else if (y + ROW_HEIGHT > scroller.scrollTop + scroller.clientHeight) {
      scroller.scrollTop = y + ROW_HEIGHT - scroller.clientHeight
    }
    selectedId = cache.get(selected)?.id ?? ''
    onSelect?.(cache.get(selected), auto)
    schedule()
  }

  async function runSearch() {
    const seq = ++searchSeq
    const q = query
    if (!q.trim()) {
      matches = []
      matchTotal = 0
      matchSet = null
      schedule()
      return
    }
    const res = await session.run(() => SearchGraph(path, q))
    if (!res || seq !== searchSeq) return
    matches = res.rows
    matchTotal = res.total
    matchSet = res.total === res.rows.length ? new Set(res.rows) : null
    schedule()
  }

  /** Moves to the next (or previous) match relative to the selection. */
  function jump(dir: 1 | -1) {
    if (matches.length === 0) return
    const next =
      dir === 1
        ? (matches.find((m) => m > selected) ?? matches[0])
        : (matches.findLast((m) => m < selected) ?? matches[matches.length - 1])
    reveal(next)
    searchInput.focus()
  }

  function onSearchKey(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      jump(e.shiftKey ? -1 : 1)
    } else if (e.key === 'Escape') {
      query = ''
      void runSearch()
      scroller.focus()
    }
  }

  function onWindowKey(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'f') {
      e.preventDefault()
      searchInput.focus()
      searchInput.select()
    }
  }

  let currentMatch = $derived(matches.indexOf(selected))

  function onKey(e: KeyboardEvent) {
    const page = Math.max(1, Math.floor(scroller.clientHeight / ROW_HEIGHT) - 1)
    const moves: Record<string, number> = {
      ArrowDown: 1,
      ArrowUp: -1,
      PageDown: page,
      PageUp: -page,
      Home: -Infinity,
      End: Infinity,
    }
    if (e.key in moves) {
      e.preventDefault()
      const d = moves[e.key]
      select(d === -Infinity ? 0 : d === Infinity ? total - 1 : (selected < 0 ? 0 : selected + d))
    }
  }

  // The canvas holds painted colours, so a theme change needs a repaint.
  $effect(() => {
    void colorTheme.resolved
    schedule()
  })

  onMount(() => {
    const ro = new ResizeObserver(schedule)
    ro.observe(scroller)
    const off = EventsOn('graph:progress', (p: { path: string; rows: number; done: boolean; refsReady: boolean; error: string }) => {
      if (p.path !== path) return
      cache.setTotal(p.rows, !p.done)
      if (p.refsReady) cache.invalidate() // labels arrived after the first rows
      if (p.done) {
        // A reload swaps in a new graph; anything cached may be stale.
        cache.invalidate()
        error = p.error
        if (query) void runSearch()
        void followSelection()
      }
      total = p.rows
      loading = !p.done
      schedule()
    })
    cache.ensure(0, 200)
    return () => {
      ro.disconnect()
      off()
      cancelAnimationFrame(frame)
    }
  })
</script>

<svelte:window onkeydown={onWindowKey} />

<div class="graph">
  <div class="status">
    {#if error}
      <span class="error">Couldn't read history: {error}</span>
    {:else if loading}
      <span>Loading history… {total.toLocaleString()} commits</span>
    {:else}
      <span>{total.toLocaleString()} commits</span>
    {/if}
    <span class="search">
      <input
        bind:this={searchInput}
        bind:value={query}
        oninput={runSearch}
        onkeydown={onSearchKey}
        type="search"
        placeholder="Search message, author or SHA"
        aria-label="Search commits"
      />
      {#if query.trim()}
        <span class="count">
          {#if matchTotal === 0}No matches{:else}{currentMatch >= 0 ? `${currentMatch + 1} of ` : ''}{matchTotal.toLocaleString()}{/if}
        </span>
        <button onclick={() => jump(-1)} disabled={!matchTotal} aria-label="Previous match">↑</button>
        <button onclick={() => jump(1)} disabled={!matchTotal} aria-label="Next match">↓</button>
      {/if}
    </span>
  </div>
  <div class="viewport">
    <canvas bind:this={canvas}></canvas>
    <div
      class="scroller"
      bind:this={scroller}
      tabindex="0"
      role="grid"
      aria-label="Commit history"
      aria-rowcount={total}
      onscroll={schedule}
      onkeydown={onKey}
      onclick={(e) => {
        const i = rowAt(e)
        if (i >= 0) select(i)
      }}
      onmousedown={(e) => {
        if (e.button !== 0 || !onDrop) return
        const i = rowAt(e)
        if (i >= 0) drag = { from: i, x: e.clientX, y: e.clientY, active: false }
      }}
      onmousemove={(e) => {
        if (drag && !drag.active && Math.hypot(e.clientX - drag.x, e.clientY - drag.y) > 6) drag.active = true
        if (drag?.active) {
          const t = rowAt(e)
          if (t !== dropTarget) {
            dropTarget = t
            schedule()
          }
        }
        const i = rowAt(e)
        if (i !== hovered) {
          hovered = i
          schedule()
        }
      }}
      onmouseup={(e) => {
        const d = drag
        const target = dropTarget
        drag = null
        dropTarget = -1
        if (!d?.active) return
        schedule()
        const s = cache.get(d.from)
        const t = target >= 0 ? cache.get(target) : undefined
        if (s && t && t.id !== s.id) onDrop?.(s, cache.refsFor(s.id), t, cache.refsFor(t.id), e.clientX, e.clientY)
      }}
      oncontextmenu={(e) => {
        const i = rowAt(e)
        const row = i >= 0 ? cache.get(i) : undefined
        if (!row || !onContext) return
        e.preventDefault()
        select(i)
        onContext(row, cache.refsFor(row.id), e.clientX, e.clientY)
      }}
      onmouseleave={() => {
        drag = null
        dropTarget = -1
        hovered = -1
        schedule()
      }}
    >
      <div class="spacer" style:height="{total * ROW_HEIGHT}px"></div>
    </div>
  </div>
</div>

<style>
  .graph {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .status {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 4px 12px;
    color: var(--ink-mute);
    border-bottom: 1px solid var(--paper-edge);
    font-size: 12px;
  }

  .search {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .search input {
    width: 260px;
    padding: 2px 6px;
    background: var(--paper);
    color: var(--ink);
    border: 1px solid var(--paper-edge);
  }

  .search button {
    background: none;
    border: 1px solid var(--paper-edge);
    color: var(--ink);
    cursor: pointer;
  }

  .count {
    white-space: nowrap;
  }

  .error {
    color: var(--danger);
  }

  /* The canvas paints the visible rows; the transparent scroller above it provides the
     native scrollbar and input, with a spacer as tall as all rows. */
  .viewport {
    position: relative;
    flex: 1;
    min-height: 0;
  }

  canvas {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }

  .scroller {
    position: absolute;
    inset: 0;
    overflow-y: auto;
    overflow-x: hidden;
    outline: none;
  }

  .spacer {
    width: 1px;
  }
</style>
