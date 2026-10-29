import { describe, expect, it } from 'vitest'
import { compareCells, parseDelimited } from './csv'

describe('parseDelimited', () => {
  it('reads plain rows', () => {
    expect(parseDelimited('a,b\n1,2\n', ',').rows).toEqual([
      ['a', 'b'],
      ['1', '2'],
    ])
  })

  it('reads quoted fields with delimiters, doubled quotes and line breaks', () => {
    expect(parseDelimited('"x, y","say ""hi""","two\nlines"\r\nz,,end', ',').rows).toEqual([
      ['x, y', 'say "hi"', 'two\nlines'],
      ['z', '', 'end'],
    ])
  })

  it('reads tabs', () => {
    expect(parseDelimited('a\tb\n1\t2', '\t').rows).toEqual([
      ['a', 'b'],
      ['1', '2'],
    ])
  })

  it('stops at the row cap and says so', () => {
    const res = parseDelimited('1\n2\n3\n4\n', ',', 2)
    expect(res).toEqual({ rows: [['1'], ['2']], truncated: true })
    expect(parseDelimited('1\n2\n', ',', 2).truncated).toBe(false)
  })
})

describe('compareCells', () => {
  it('sorts numbers numerically and text without case', () => {
    expect(['10', '9', '100'].sort(compareCells)).toEqual(['9', '10', '100'])
    expect(['b', 'A', 'c'].sort(compareCells)).toEqual(['A', 'b', 'c'])
  })
})
