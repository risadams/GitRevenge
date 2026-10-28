<script lang="ts">
  import { onMount } from 'svelte'
  import { GraphRows } from '../../../wailsjs/go/app/App.js'
  import { EventsOn } from '../../../wailsjs/runtime/runtime.js'
  import { drawRow, graphWidth, layoutColumns, ROW_HEIGHT, type Theme } from './draw'
  import { RowCache, type Page, type Row } from './rowCache'

  let { path, onSelect }: { path: string; onSelect?: (row: Row | undefined) => void } = $props()

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
    return {
      bg: css.getPropertyValue('--bg').trim(),
      fg: css.getPropertyValue('--fg').trim(),
      muted: css.getPropertyValue('--fg-muted').trim(),
      selected: css.getPropertyValue('--bg-selected').trim() || '#264f78',
      hover: css.getPropertyValue('--bg-hover').trim() || '#2a2d2e',
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
      if (row) drawRow(ctx, row, cache.refsFor(row.id), y, lay, th, i === selected, i === hovered)
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

  let autoSelected = false

  /** On first load, select the checked-out commit (or the top row) so details show. */
  function selectHeadOnce() {
    if (autoSelected || selected >= 0 || !cache.get(0)) return
    autoSelected = true
    for (let i = 0; i < 200 && cache.get(i); i++) {
      if (cache.refsFor(cache.get(i)!.id).some((r) => r.head)) {
        select(i)
        return
      }
    }
    select(0)
  }

  function select(i: number) {
    selected = Math.max(0, Math.min(total - 1, i))
    // Keep the selection in view.
    const y = selected * ROW_HEIGHT
    if (y < scroller.scrollTop) scroller.scrollTop = y
    else if (y + ROW_HEIGHT > scroller.scrollTop + scroller.clientHeight) {
      scroller.scrollTop = y + ROW_HEIGHT - scroller.clientHeight
    }
    onSelect?.(cache.get(selected))
    schedule()
  }

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

  onMount(() => {
    const ro = new ResizeObserver(schedule)
    ro.observe(scroller)
    const off = EventsOn('graph:progress', (p: { path: string; rows: number; done: boolean; error: string }) => {
      if (p.path !== path) return
      cache.setTotal(p.rows, !p.done)
      if (p.done) {
        // A reload swaps in a new graph; anything cached may be stale.
        cache.invalidate()
        error = p.error
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

<div class="graph">
  <div class="status">
    {#if error}
      <span class="error">Couldn't read history: {error}</span>
    {:else if loading}
      <span>Loading history… {total.toLocaleString()} commits</span>
    {:else}
      <span>{total.toLocaleString()} commits</span>
    {/if}
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
      onmousemove={(e) => {
        const i = rowAt(e)
        if (i !== hovered) {
          hovered = i
          schedule()
        }
      }}
      onmouseleave={() => {
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
    padding: 4px 12px;
    color: var(--fg-muted);
    border-bottom: 1px solid var(--border);
    font-size: 12px;
  }

  .error {
    color: #f0a0a0;
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
