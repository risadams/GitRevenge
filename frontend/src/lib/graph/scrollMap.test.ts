import { describe, expect, it } from 'vitest'
import { scrollMap, toScroll, toVirtual } from './scrollMap'

describe('scrollMap', () => {
  it('leaves a history that fits unscaled', () => {
    const m = scrollMap(5_000, 600, 10_000)
    expect(m).toEqual({ spacer: 5_000, ratio: 1, maxTop: 4_400 })
    expect(toVirtual(1_234, m)).toBe(1_234)
  })

  it('caps a taller history and maps the ends to the ends', () => {
    const rows = 1_300_000
    const m = scrollMap(rows * 52, 800, 8_000_000)
    expect(m.spacer).toBe(8_000_000)
    expect(toVirtual(0, m)).toBe(0)
    // Scrolled to the bottom of the capped area, the last row sits at the bottom of the view.
    expect(toVirtual(8_000_000 - 800, m)).toBeCloseTo(rows * 52 - 800)
  })

  it('round-trips a virtual position', () => {
    const m = scrollMap(68_000_000, 800, 8_000_000)
    expect(toVirtual(toScroll(30_000_000, m), m)).toBeCloseTo(30_000_000)
  })
})
