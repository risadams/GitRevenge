import { describe, expect, it } from 'vitest'
import { hasMarkers, parseConflicts, resolveRegion } from './markers'

const merged = [
  'header\n',
  '<<<<<<< HEAD\n',
  'ours line\n',
  '=======\n',
  'theirs line\n',
  '>>>>>>> feature\n',
  'middle\n',
  '<<<<<<< HEAD\n',
  'A\n',
  '||||||| base\n',
  'B\n',
  '=======\n',
  'C\n',
  '>>>>>>> feature\n',
  'footer\n',
].join('')

describe('parseConflicts', () => {
  it('finds blocks, labels and diff3 base sections', () => {
    const r = parseConflicts(merged)
    expect(r).toHaveLength(2)
    expect(r[0]).toMatchObject({ ours: 'ours line\n', theirs: 'theirs line\n', base: null, oursLabel: 'HEAD', theirsLabel: 'feature' })
    expect(r[1]).toMatchObject({ ours: 'A\n', base: 'B\n', theirs: 'C\n' })
    expect(merged.slice(r[0].from, r[0].to).startsWith('<<<<<<<')).toBe(true)
    expect(merged.slice(r[0].from, r[0].to).endsWith('>>>>>>> feature\n')).toBe(true)
  })

  it('handles CRLF files', () => {
    const r = parseConflicts('<<<<<<< a\r\nx\r\n=======\r\ny\r\n>>>>>>> b\r\n')
    expect(r).toHaveLength(1)
    expect(r[0].ours).toBe('x\r\n')
  })

  it('ignores incomplete blocks and look-alike lines', () => {
    expect(parseConflicts('<<<<<<< a\nno end\n')).toEqual([])
    expect(parseConflicts('a ======= b\n')).toEqual([])
  })
})

describe('resolveRegion', () => {
  it('resolves one block at a time, leaving the others', () => {
    let text = merged
    text = resolveRegion(text, parseConflicts(text)[0], 'theirs')
    expect(text.startsWith('header\ntheirs line\nmiddle\n')).toBe(true)
    expect(parseConflicts(text)).toHaveLength(1)
    text = resolveRegion(text, parseConflicts(text)[0], 'both')
    expect(text).toBe('header\ntheirs line\nmiddle\nA\nC\nfooter\n')
    expect(hasMarkers(text)).toBe(false)
  })

  it('can take the base side of a diff3 block', () => {
    const r = parseConflicts(merged)[1]
    expect(resolveRegion(merged, r, 'base')).toContain('middle\nB\nfooter\n')
  })
})

describe('hasMarkers', () => {
  it('detects start or end markers only at line starts', () => {
    expect(hasMarkers(merged)).toBe(true)
    expect(hasMarkers('code // <<<<<<< not a marker\n')).toBe(false)
  })
})
