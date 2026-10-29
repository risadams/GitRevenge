import { describe, expect, it } from 'vitest'
import { fuzzy } from './fuzzy'

const rank = (query: string, texts: string[]) =>
  texts
    .map((t) => ({ t, m: fuzzy(query, t) }))
    .filter((x) => x.m)
    .sort((a, b) => b.m!.score - a.m!.score)
    .map((x) => x.t)

describe('fuzzy', () => {
  it('needs every character, in order', () => {
    expect(fuzzy('fth', 'Fetch')).not.toBeNull()
    expect(fuzzy('hcf', 'Fetch')).toBeNull()
    expect(fuzzy('zz', 'Fetch')).toBeNull()
  })

  it('matches everything with an empty query', () => {
    expect(fuzzy('  ', 'anything')).toEqual({ score: 0, positions: [] })
  })

  it('reports the matched positions, preferring word starts', () => {
    expect(fuzzy('co', 'Check out main')?.positions).toEqual([0, 6])
    expect(fuzzy('nb', 'New branch…')?.positions).toEqual([0, 4])
  })

  it('ranks prefixes and word starts first', () => {
    expect(rank('pu', ['Push', 'Pull', 'Stash push', 'Open recent: puzzle'])[0]).toMatch(/^Pu/)
    expect(rank('pr', ['Pull (rebase)', 'Open repository…', 'Prompt'])).toEqual(['Prompt', 'Pull (rebase)', 'Open repository…'])
    expect(rank('dark', ['Switch to the dark theme', 'Discard all'])[0]).toBe('Switch to the dark theme')
  })

  it('ignores spaces and case in the query', () => {
    expect(fuzzy('CHECK OUT', 'Check out main')).not.toBeNull()
  })

  it('falls back to a plain occurrence when jumping to a word start would strand the rest', () => {
    expect(fuzzy('ab', 'xab Ax')).not.toBeNull()
  })
})
