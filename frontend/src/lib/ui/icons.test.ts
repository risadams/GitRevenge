import { describe, expect, it } from 'vitest'
import { PATHS } from './icons'

const ARGS: Record<string, number> = { m: 2, l: 2, t: 2, h: 1, v: 1, c: 6, s: 4, q: 4, a: 7, z: 0 }

/** Every end and control point a path visits, in absolute coordinates. */
function points(d: string): [number, number][] {
  const out: [number, number][] = []
  let x = 0
  let y = 0
  let startX = 0
  let startY = 0
  for (const [, cmd, rest] of d.matchAll(/([a-zA-Z])([^a-zA-Z]*)/g)) {
    const nums = (rest.match(/-?(?:\d+\.?\d*|\.\d+)/g) ?? []).map(Number)
    const lower = cmd.toLowerCase()
    const rel = cmd === lower
    if (lower === 'z') {
      x = startX
      y = startY
      continue
    }
    for (let i = 0; i < nums.length; i += ARGS[lower]) {
      const a = nums.slice(i, i + ARGS[lower])
      const ox = rel ? x : 0
      const oy = rel ? y : 0
      if (lower === 'h') x = ox + a[0]
      else if (lower === 'v') y = oy + a[0]
      else if (lower === 'a') {
        x = ox + a[5]
        y = oy + a[6]
      } else {
        // Control points first, then the end point, all relative to the same start.
        for (let j = 0; j < a.length - 2; j += 2) out.push([ox + a[j], oy + a[j + 1]])
        x = ox + a[a.length - 2]
        y = oy + a[a.length - 1]
      }
      if (lower === 'm' && i === 0) {
        startX = x
        startY = y
      }
      out.push([x, y])
    }
  }
  return out
}

describe('icons', () => {
  // A part that lands outside the grid is clipped by the 24px box, as push, pull, back and
  // search were when a joined part's leading "m" was read as relative.
  it.each(Object.entries(PATHS))('%s stays inside the 24 unit grid', (_, d) => {
    const outside = points(d).filter(([x, y]) => x < 1 || x > 23 || y < 1 || y > 23)
    expect(outside).toEqual([])
  })
})
