import type { Ref, Row } from './rowCache'

export const ROW_HEIGHT = 26
export const LANE_WIDTH = 16
export const GRAPH_PAD = 10
const NODE_RADIUS = 5

/** Lane colors, cycled by color index. Plain until the 0.6 design pass. */
export const PALETTE = [
  '#4fc1ff',
  '#f78c6c',
  '#c3e88d',
  '#c792ea',
  '#ffcb6b',
  '#ff5370',
  '#89ddff',
  '#f07178',
  '#82aaff',
  '#b2ccd6',
]

export function laneColor(i: number): string {
  return PALETTE[((i % PALETTE.length) + PALETTE.length) % PALETTE.length]
}

/** X coordinate of the center of lane column col. */
export function laneX(col: number): number {
  return GRAPH_PAD + col * LANE_WIDTH + LANE_WIDTH / 2
}

/** Width in pixels of the graph column for rows using up to `lanes` columns. */
export function graphWidth(lanes: number, max: number): number {
  return Math.min(max, GRAPH_PAD * 2 + Math.max(1, lanes) * LANE_WIDTH)
}

const dateFmt = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' })

export function formatTime(unixSeconds: number): string {
  return dateFmt.format(new Date(unixSeconds * 1000))
}

export interface Theme {
  bg: string
  fg: string
  muted: string
  selected: string
  hover: string
}

export interface Layout {
  width: number
  graphW: number
  /** Right-hand columns: author, date, sha. */
  authorW: number
  dateW: number
  shaW: number
}

export function layoutColumns(width: number, graphW: number): Layout {
  const shaW = 72
  const dateW = width > 900 ? 150 : 0
  const authorW = width > 700 ? 150 : 0
  return { width, graphW, authorW, dateW, shaW }
}

/** Draws one row whose top edge is at y. */
export function drawRow(
  ctx: CanvasRenderingContext2D,
  row: Row,
  refs: Ref[],
  y: number,
  lay: Layout,
  theme: Theme,
  selected: boolean,
  hovered: boolean,
  dimmed = false,
) {
  const h = ROW_HEIGHT
  ctx.save()
  // Rows that don't match an active search fade back.
  if (dimmed && !selected) ctx.globalAlpha = 0.3
  if (selected || hovered) {
    ctx.fillStyle = selected ? theme.selected : theme.hover
    ctx.fillRect(0, y, lay.width, h)
  }

  // Lines, clipped to the graph column so wide graphs don't run under the text.
  ctx.save()
  ctx.beginPath()
  ctx.rect(0, y, lay.graphW, h)
  ctx.clip()
  ctx.lineWidth = 2
  const mid = y + h / 2
  for (const s of row.up) segment(ctx, laneX(s.from), y, laneX(s.to), mid, laneColor(s.color))
  for (const s of row.down) segment(ctx, laneX(s.from), mid, laneX(s.to), y + h, laneColor(s.color))

  const isStash = refs.some((r) => r.kind === 'stash')
  if (isStash) {
    // Stashes are drawn as hollow squares so they read as "not a real commit".
    const s = NODE_RADIUS + 0.5
    ctx.strokeStyle = laneColor(row.color)
    ctx.lineWidth = 2
    ctx.fillStyle = theme.bg
    ctx.fillRect(laneX(row.col) - s, mid - s, s * 2, s * 2)
    ctx.strokeRect(laneX(row.col) - s, mid - s, s * 2, s * 2)
  } else {
    ctx.beginPath()
    ctx.arc(laneX(row.col), mid, NODE_RADIUS, 0, Math.PI * 2)
    ctx.fillStyle = laneColor(row.color)
    ctx.fill()
  }
  if (row.parents > 1 && !isStash) {
    // Merge commits get a hollow center.
    ctx.beginPath()
    ctx.arc(laneX(row.col), mid, NODE_RADIUS - 2.5, 0, Math.PI * 2)
    ctx.fillStyle = selected ? theme.selected : hovered ? theme.hover : theme.bg
    ctx.fill()
  }
  ctx.restore()

  // Text columns.
  const textY = mid
  let right = lay.width - 8
  ctx.textBaseline = 'middle'
  ctx.fillStyle = theme.muted
  ctx.textAlign = 'right'
  ctx.fillText(row.id.slice(0, 7), right, textY)
  right -= lay.shaW
  if (lay.dateW) {
    ctx.fillText(formatTime(row.authorTime), right, textY)
    right -= lay.dateW
  }
  if (lay.authorW) {
    ctx.fillText(clip(ctx, row.author, lay.authorW - 12), right, textY)
    right -= lay.authorW
  }
  ctx.textAlign = 'left'
  let left = lay.graphW + 8
  const subjectRight = right - 12
  left = drawLabels(ctx, refs, left, mid, Math.min(subjectRight - left, (subjectRight - left) * 0.6), laneColor(row.color), theme)
  ctx.fillStyle = theme.fg
  ctx.fillText(clip(ctx, row.subject, subjectRight - left), left, textY)
  ctx.restore()
}

/** Sort order for labels: checked-out branch, local, remote, tags, stashes. */
const KIND_ORDER: Record<Ref['kind'], number> = { local: 1, remote: 2, tag: 3, stash: 4 }

export function sortRefs(refs: Ref[]): Ref[] {
  return [...refs].sort((a, b) => Number(b.head) - Number(a.head) || KIND_ORDER[a.kind] - KIND_ORDER[b.kind])
}

/** Draws ref pills starting at x; returns the x where the subject should start. */
function drawLabels(
  ctx: CanvasRenderingContext2D,
  refs: Ref[],
  x: number,
  mid: number,
  maxWidth: number,
  color: string,
  theme: Theme,
): number {
  if (refs.length === 0) return x
  const end = x + maxWidth
  const h = 18
  const sorted = sortRefs(refs)
  ctx.save()
  ctx.font = '11px system-ui, -apple-system, "Segoe UI", sans-serif'
  for (let i = 0; i < sorted.length; i++) {
    const r = sorted[i]
    const remaining = sorted.length - i
    const more = remaining > 1 ? ` +${remaining - 1}` : ''
    const prefix = r.kind === 'tag' ? '◆ ' : r.kind === 'stash' ? '≡ ' : ''
    const text = prefix + r.short
    let w = ctx.measureText(text).width + 12
    if (x + w > end) {
      // Out of room: one clipped pill that also counts the rest.
      const label = clip(ctx, text, Math.max(0, end - x - 12 - ctx.measureText(more).width)) + more
      w = ctx.measureText(label).width + 12
      if (w > 24) pill(ctx, label, x, mid, w, h, r, color, theme)
      x += w + 4
      break
    }
    pill(ctx, text, x, mid, w, h, r, color, theme)
    x += w + 4
  }
  ctx.restore()
  return x + 4
}

function pill(
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  mid: number,
  w: number,
  h: number,
  r: Ref,
  color: string,
  theme: Theme,
) {
  const y = mid - h / 2
  ctx.beginPath()
  ctx.roundRect(x, y, w, h, 4)
  // Local branches are filled with the lane color; remotes, tags and stashes are outlined.
  if (r.kind === 'local') {
    ctx.fillStyle = color
    ctx.fill()
    ctx.fillStyle = theme.bg
  } else {
    ctx.strokeStyle = color
    ctx.lineWidth = 1
    ctx.stroke()
    ctx.fillStyle = color
  }
  ctx.font = `${r.head ? 'bold ' : ''}11px system-ui, -apple-system, "Segoe UI", sans-serif`
  ctx.textBaseline = 'middle'
  ctx.fillText(text, x + 6, mid)
}

function segment(ctx: CanvasRenderingContext2D, x1: number, y1: number, x2: number, y2: number, color: string) {
  ctx.strokeStyle = color
  ctx.beginPath()
  ctx.moveTo(x1, y1)
  if (x1 === x2) {
    ctx.lineTo(x2, y2)
  } else {
    // Vertical tangents at both ends so lines enter and leave rows straight.
    const cy = (y1 + y2) / 2
    ctx.bezierCurveTo(x1, cy, x2, cy, x2, y2)
  }
  ctx.stroke()
}

/** Truncates text with an ellipsis to fit maxWidth pixels. */
export function clip(ctx: Pick<CanvasRenderingContext2D, 'measureText'>, text: string, maxWidth: number): string {
  if (maxWidth <= 0) return ''
  if (ctx.measureText(text).width <= maxWidth) return text
  let lo = 0
  let hi = text.length
  while (lo < hi) {
    const m = (lo + hi + 1) >> 1
    if (ctx.measureText(text.slice(0, m) + '…').width <= maxWidth) lo = m
    else hi = m - 1
  }
  return text.slice(0, lo) + '…'
}
