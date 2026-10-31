import { describe, expect, it } from 'vitest'
import type { Ref } from './graph/rowCache'
import { stackLabel, stackOrder } from './stacks'

const ref = (short: string, stackParent = '', stackDepth = 0, stackSize = 0): Ref => ({
  name: 'refs/heads/' + short,
  short,
  kind: 'local',
  commit: short,
  head: false,
  upstream: '',
  ahead: 0,
  behind: 0,
  gone: false,
  stackParent,
  stackDepth,
  stackSize,
})

describe('stackOrder', () => {
  it('puts each stacked branch after its parent', () => {
    const refs = [ref('a-top', 'mid', 3, 3), ref('alpha'), ref('base', '', 1, 3), ref('mid', 'base', 2, 3), ref('zeta')]
    expect(stackOrder(refs).map((r) => r.short)).toEqual(['alpha', 'base', 'mid', 'a-top', 'zeta'])
  })

  it('keeps branches whose parent is elsewhere in place', () => {
    const refs = [ref('x', 'gone-parent', 2, 2), ref('y')]
    expect(stackOrder(refs).map((r) => r.short)).toEqual(['x', 'y'])
  })
})

describe('stackLabel', () => {
  it('names the place in the stack', () => {
    expect(stackLabel(ref('mid', 'base', 2, 3))).toBe('2 of 3')
    expect(stackLabel(ref('solo'))).toBe('')
  })
})
