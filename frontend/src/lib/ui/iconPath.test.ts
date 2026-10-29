import { describe, expect, it } from 'vitest'
import { joinParts } from './iconPath'

describe('joinParts', () => {
  it('makes a leading relative move absolute, so a part starts where it was drawn', () => {
    expect(joinParts(['M12 3v12', 'm7 10 5 5 5-5'])).toBe('M12 3v12 M7 10l5 5 5-5')
  })

  it('keeps the pairs after a leading move relative', () => {
    expect(joinParts(['m6 9 6 6 6-6'])).toBe('M6 9l6 6 6-6')
  })

  it('adds no line-to when a command follows the move', () => {
    expect(joinParts(['m20 20-3.5-3.5'])).toBe('M20 20l-3.5-3.5')
    expect(joinParts(['m9 12 2 2 4-4'])).toBe('M9 12l2 2 4-4')
    expect(joinParts(['m3 3h4'])).toBe('M3 3h4')
  })

  it('leaves relative moves inside a part alone', () => {
    expect(joinParts(['M3 3h4m2 0h4'])).toBe('M3 3h4m2 0h4')
  })
})
