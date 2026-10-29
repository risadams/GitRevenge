import { describe, expect, it } from 'vitest'
import type { Ref, Row } from './graph/rowCache'
import { dropItems, type MenuContext } from './menus'

const row = (id: string): Row => ({ index: 0, col: 0, color: 0, up: [], down: [], width: 1, id, parents: 1, author: '', authorTime: 0, subject: '' })
const ref = (short: string, kind: Ref['kind'], head = false): Ref => ({
  name: kind === 'tag' ? `refs/tags/${short}` : short,
  short,
  kind,
  head,
  commit: '',
  upstream: '',
  ahead: 0,
  behind: 0,
  gone: false,
})
const ctx = (current: string): MenuContext => ({ path: '/r', current: () => current, openDialog: () => {}, openRebase: () => {} })
const labels = (items: { label: string }[]) => items.map((i) => i.label)

describe('dropItems', () => {
  const head = [ref('main', 'local', true)]

  it('dropping a branch onto the checked-out branch merges or rebases', () => {
    const items = dropItems(row('a'.repeat(40)), [ref('feature', 'local')], row('b'.repeat(40)), head, ctx('main'))
    expect(labels(items)).toEqual(['Merge feature into main', 'Merge feature into main (no fast-forward)', 'Rebase main onto feature'])
  })

  it('dropping a bare commit onto the checked-out branch also offers cherry-pick', () => {
    const items = dropItems(row('c'.repeat(40)), [], row('b'.repeat(40)), head, ctx('main'))
    expect(labels(items)).toContain('Cherry-pick ccccccc onto main')
  })

  it('dragging the checked-out branch onto a tag offers rebasing onto it', () => {
    const items = dropItems(row('b'.repeat(40)), head, row('d'.repeat(40)), [ref('v1.0', 'tag')], ctx('main'))
    expect(labels(items)).toContain('Rebase main onto v1.0')
  })

  it('offers nothing when neither side is checked out, or HEAD is detached', () => {
    expect(dropItems(row('a'.repeat(40)), [ref('x', 'local')], row('b'.repeat(40)), [ref('y', 'local')], ctx('main'))).toEqual([])
    expect(dropItems(row('a'.repeat(40)), [ref('x', 'local')], row('b'.repeat(40)), head, ctx(''))).toEqual([])
  })
})
