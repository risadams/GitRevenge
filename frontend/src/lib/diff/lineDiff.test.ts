import { describe, expect, it } from 'vitest'
import { applyChanges, diffLines, splitLines, toHunks, type Op } from './lineDiff'

const L = (...lines: string[]) => lines.map((l) => l + '\n')
const kinds = (ops: Op[]) => ops.map((o) => o.kind[0]).join('')

describe('splitLines', () => {
  it('keeps terminators so text round-trips', () => {
    const text = 'a\r\nb\nc'
    expect(splitLines(text)).toEqual(['a\r\n', 'b\n', 'c'])
    expect(splitLines(text).join('')).toBe(text)
    expect(splitLines('')).toEqual([])
  })
})

describe('diffLines', () => {
  it('finds a minimal edit', () => {
    const ops = diffLines(L('a', 'b', 'c', 'd'), L('a', 'x', 'c', 'd', 'e'))
    expect(kinds(ops)).toBe('edaeea')
  })

  it('handles empty sides', () => {
    expect(kinds(diffLines([], L('a', 'b')))).toBe('aa')
    expect(kinds(diffLines(L('a'), []))).toBe('d')
    expect(diffLines([], [])).toEqual([])
  })

  it('puts deletions before additions within a change', () => {
    expect(kinds(diffLines(L('a', 'old1', 'old2', 'z'), L('a', 'new1', 'z')))).toBe('eddae')
  })

  it('always reproduces both sides', () => {
    // Pseudo-random edits on a pseudo-random file.
    let seed = 7
    const rnd = (n: number) => (seed = (seed * 1103515245 + 12345) % 2 ** 31) % n
    for (let trial = 0; trial < 50; trial++) {
      const a = Array.from({ length: rnd(40) }, () => `l${rnd(8)}\n`)
      const b = a.filter(() => rnd(4) !== 0).flatMap((l) => (rnd(5) === 0 ? [l, `n${rnd(8)}\n`] : [l]))
      const ops = diffLines(a, b)
      expect(applyChanges(ops, a, b, () => true)).toBe(b.join(''))
      expect(applyChanges(ops, a, b, () => false)).toBe(a.join(''))
    }
  })
})

describe('toHunks', () => {
  it('groups nearby changes and reports git-style headers', () => {
    const a = L(...'abcdefghijklmnopqrst'.split(''))
    const b = [...a]
    b[2] = 'C\n' // change near the top
    b[4] = 'E\n' // close enough to merge with the first
    b[17] = 'R\n' // far away: separate hunk
    const hunks = toHunks(diffLines(a, b))
    expect(hunks).toHaveLength(2)
    expect(hunks[0]).toMatchObject({ oldStart: 1, oldCount: 8, newStart: 1, newCount: 8 })
    expect(hunks[1]).toMatchObject({ oldStart: 15, oldCount: 6, newStart: 15, newCount: 6 })
  })

  it('numbers a pure insertion into an empty file like git', () => {
    const hunks = toHunks(diffLines([], L('a', 'b')))
    expect(hunks[0]).toMatchObject({ oldStart: 0, oldCount: 0, newStart: 1, newCount: 2 })
  })
})

describe('applyChanges for partial staging', () => {
  const index = L('one', 'two', 'three')
  const work = L('ONE', 'two', 'THREE', 'four')
  const ops = diffLines(index, work)

  it('stages only selected lines', () => {
    // Apply the change to line one (its deletion and addition) only.
    const sel = new Set(ops.flatMap((o, i) => ((o.kind === 'del' && index[o.oldIndex] === 'one\n') || (o.kind === 'add' && work[o.newIndex] === 'ONE\n') ? [i] : [])))
    expect(applyChanges(ops, index, work, (i) => sel.has(i))).toBe('ONE\ntwo\nthree\n')
  })

  it('discards selected lines by applying everything else', () => {
    const four = ops.findIndex((o) => o.kind === 'add' && work[o.newIndex] === 'four\n')
    expect(applyChanges(ops, index, work, (i) => i !== four)).toBe('ONE\ntwo\nTHREE\n')
  })
})
