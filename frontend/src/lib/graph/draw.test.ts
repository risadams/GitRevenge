import { describe, expect, it } from 'vitest'
import { clip, graphWidth, LANE_WIDTH, laneColor, laneX, PALETTE, sortRefs } from './draw'

describe('graph geometry', () => {
  it('spaces lanes evenly', () => {
    expect(laneX(1) - laneX(0)).toBe(LANE_WIDTH)
  })

  it('caps the graph column width', () => {
    expect(graphWidth(3, 1000)).toBeLessThan(graphWidth(10, 1000))
    expect(graphWidth(500, 300)).toBe(300)
  })

  it('cycles colors, including for negative indexes', () => {
    expect(laneColor(PALETTE.length)).toBe(PALETTE[0])
    expect(laneColor(-1)).toBe(PALETTE[PALETTE.length - 1])
  })
})

describe('clip', () => {
  // One pixel per character.
  const ctx = { measureText: (s: string) => ({ width: s.length }) as TextMetrics }

  it('leaves text that fits alone', () => {
    expect(clip(ctx, 'hello', 10)).toBe('hello')
  })

  it('truncates with an ellipsis', () => {
    const out = clip(ctx, 'hello world', 6)
    expect(out).toBe('hello…')
  })

  it('handles no room at all', () => {
    expect(clip(ctx, 'hello', 0)).toBe('')
  })
})

describe('sortRefs', () => {
  const ref = (short: string, kind: 'local' | 'remote' | 'tag' | 'stash', head = false) =>
    ({ name: short, short, kind, head, commit: '', upstream: '', ahead: 0, behind: 0, gone: false }) as const

  it('puts the checked-out branch first, then local, remote, tag', () => {
    const out = sortRefs([ref('v1', 'tag'), ref('origin/main', 'remote'), ref('dev', 'local'), ref('main', 'local', true)])
    expect(out.map((r) => r.short)).toEqual(['main', 'dev', 'origin/main', 'v1'])
  })
})
