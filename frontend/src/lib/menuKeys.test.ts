import { describe, expect, it } from 'vitest'
import { edge, step, typeahead } from './menuKeys'

const items = [{ label: 'Checkout' }, { label: 'Cherry-pick', disabled: true }, { label: 'Create branch' }, { label: 'Delete' }]

describe('step', () => {
  it('skips disabled items and wraps', () => {
    expect(step(items, 0, 1)).toBe(2)
    expect(step(items, 3, 1)).toBe(0)
    expect(step(items, 2, -1)).toBe(0)
    expect(step(items, 0, -1)).toBe(3)
  })

  it('starts from nothing selected', () => {
    expect(step(items, -1, 1)).toBe(0)
  })

  it('returns -1 when every item is disabled', () => {
    expect(step([{ label: 'a', disabled: true }], 0, 1)).toBe(-1)
  })
})

describe('edge', () => {
  it('finds the first and last enabled items', () => {
    expect(edge(items, false)).toBe(0)
    expect(edge(items, true)).toBe(3)
    expect(edge([{ label: 'a', disabled: true }, { label: 'b' }], false)).toBe(1)
  })
})

describe('typeahead', () => {
  it('cycles through items starting with a repeated letter, skipping disabled ones', () => {
    expect(typeahead(items, 0, 'c')).toBe(2)
    expect(typeahead(items, 2, 'c')).toBe(0)
  })

  it('matches a longer prefix from the current item, ignoring case', () => {
    expect(typeahead(items, 0, 'cr')).toBe(2)
    expect(typeahead(items, 2, 'CR')).toBe(2)
  })

  it('returns -1 without a match', () => {
    expect(typeahead(items, 0, 'x')).toBe(-1)
  })
})
