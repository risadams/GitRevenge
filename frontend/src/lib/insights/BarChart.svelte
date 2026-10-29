<script lang="ts" module>
  export interface Bar {
    label: string
    value: number
    /** Tooltip text; defaults to "label: value". */
    detail?: string
  }

  export interface Colors {
    mark: string
    text: string
    muted: string
    grid: string
  }
</script>

<script lang="ts">
  // A single-series bar chart in SVG: columns over time, or ranked horizontal bars. Colours
  // arrive as values (not CSS variables) so an exported PNG draws the same.

  let {
    bars,
    colors,
    horizontal = false,
    label,
    format = (n: number) => n.toLocaleString(),
    svg = $bindable(),
  }: {
    bars: Bar[]
    colors: Colors
    horizontal?: boolean
    /** What the chart shows, for screen readers. */
    label: string
    format?: (n: number) => string
    svg?: SVGSVGElement
  } = $props()

  let width = $state(600)
  let hovered = $state(-1)

  const max = $derived(Math.max(1, ...bars.map((b) => b.value)))

  // Columns: plot area below a little headroom, with three recessive gridlines.
  const H = 180
  const PAD = { top: 10, right: 8, bottom: 22, left: 40 }
  const plotW = $derived(Math.max(10, width - PAD.left - PAD.right))
  const plotH = H - PAD.top - PAD.bottom
  const slot = $derived(plotW / Math.max(1, bars.length))
  const barW = $derived(Math.max(1, Math.min(28, slot - 2)))
  const ticks = $derived([0, 0.5, 1].map((f) => Math.round(max * f)))
  const xLabels = $derived(bars.length > 2 ? [0, Math.floor(bars.length / 2), bars.length - 1] : bars.map((_, i) => i))

  // Ranked bars: a label column, then the bar, then its value.
  const ROW = 24
  const LABEL_W = $derived(Math.min(240, width * 0.38))
  const VALUE_W = 64
  const hbarMax = $derived(Math.max(10, width - LABEL_W - VALUE_W - 12))
  /** Characters that fit the label column at 12px monospace (about 7.3px each). */
  const labelChars = $derived(Math.max(6, Math.floor((LABEL_W - 12) / 7.3)))
  /** Long labels keep their end, the useful part of a path. */
  const fit = (s: string) => (s.length > labelChars ? '…' + s.slice(-(labelChars - 1)) : s)

  /** A bar with its data end rounded (4px) and its baseline end square. */
  function column(x: number, y: number, w: number, h: number): string {
    const r = Math.min(4, w / 2, h)
    return `M${x},${y + h} V${y + r} Q${x},${y} ${x + r},${y} H${x + w - r} Q${x + w},${y} ${x + w},${y + r} V${y + h} Z`
  }

  function hbar(x: number, y: number, w: number, h: number): string {
    const r = Math.min(4, h / 2, w)
    return `M${x},${y} H${x + w - r} Q${x + w},${y} ${x + w},${y + r} V${y + h - r} Q${x + w},${y + h} ${x + w - r},${y + h} H${x} Z`
  }

  const tip = (b: Bar) => b.detail ?? `${b.label}: ${format(b.value)}`
</script>

<div class="chart" bind:clientWidth={width}>
  {#if horizontal}
    <svg bind:this={svg} {width} height={bars.length * ROW + 4} role="img" aria-label={label}>
      {#each bars as b, i (b.label)}
        {@const y = i * ROW + 4}
        {@const w = (b.value / max) * hbarMax}
        <g
          role="presentation"
          onmouseenter={() => (hovered = i)}
          onmouseleave={() => (hovered = -1)}
          opacity={hovered >= 0 && hovered !== i ? 0.55 : 1}
        >
          <rect x="0" {y} {width} height={ROW} fill="transparent" />
          <text x={LABEL_W - 8} y={y + ROW / 2 + 4} text-anchor="end" font-size="12" fill={colors.text}
            >{fit(b.label)}</text
          >
          <path d={hbar(LABEL_W, y + 5, Math.max(2, w), ROW - 10)} fill={colors.mark} />
          <text x={LABEL_W + Math.max(2, w) + 6} y={y + ROW / 2 + 4} font-size="12" fill={colors.muted}>{format(b.value)}</text>
          <title>{tip(b)}</title>
        </g>
      {/each}
    </svg>
  {:else}
    <svg bind:this={svg} {width} height={H} role="img" aria-label={label}>
      {#each ticks as t, k (k)}
        {@const y = PAD.top + plotH - (t / max) * plotH}
        <line x1={PAD.left} x2={width - PAD.right} y1={y} y2={y} stroke={colors.grid} stroke-width="1" />
        <text x={PAD.left - 6} y={y + 4} text-anchor="end" font-size="11" fill={colors.muted}>{format(t)}</text>
      {/each}
      {#each bars as b, i (i)}
        {@const h = (b.value / max) * plotH}
        {@const x = PAD.left + i * slot + (slot - barW) / 2}
        <g role="presentation" onmouseenter={() => (hovered = i)} onmouseleave={() => (hovered = -1)}>
          <!-- The hit target is the whole column, bigger than the mark. -->
          <rect x={PAD.left + i * slot} y={PAD.top} width={slot} height={plotH} fill="transparent" />
          {#if b.value > 0}<path d={column(x, PAD.top + plotH - h, barW, h)} fill={colors.mark} opacity={hovered >= 0 && hovered !== i ? 0.55 : 1} />{/if}
          <title>{tip(b)}</title>
        </g>
      {/each}
      {#each xLabels as i (i)}
        <text
          x={PAD.left + i * slot + slot / 2}
          y={H - 6}
          text-anchor={i === 0 ? 'start' : i === bars.length - 1 ? 'end' : 'middle'}
          font-size="11"
          fill={colors.muted}>{bars[i]?.label}</text
        >
      {/each}
    </svg>
  {/if}
  {#if hovered >= 0 && bars[hovered]}
    <div class="tooltip" role="status">{tip(bars[hovered])}</div>
  {/if}
</div>

<style>
  .chart {
    position: relative;
    width: 100%;
  }

  svg {
    display: block;
    font-family: var(--font-mono);
  }

  .tooltip {
    position: absolute;
    top: 0;
    right: 0;
    padding: 4px 8px;
    background: var(--paper);
    border: 1px solid var(--paper-edge);
    border-radius: 4px;
    box-shadow: var(--shadow);
    font-size: 12.5px;
    color: var(--ink);
    pointer-events: none;
  }
</style>
