import { describe, expect, it } from 'vitest'
import { canonicalKey, keyInfo, parseBool, sectionOf } from './catalog'
import { buildRows, filterSections, hasSecret, maskUrls, maskValue, overrideNote, samePath, type Entry } from './rows'

const e = (scope: string, file: string, key: string, value: string, implicit = false): Entry => ({ scope, file, key, value, implicit })

const LOCAL = 'C:\\repo\\.git\\config'
const GLOBAL = 'C:\\Users\\me\\.gitconfig'

describe('catalog', () => {
  it('matches keys regardless of section and name case, and fills <name>', () => {
    expect(keyInfo('PULL.REBASE')?.type).toBe('enum')
    expect(keyInfo('remote.upstream.fetch')?.multi).toBe(true)
    expect(keyInfo('remote.fetch')).toBeUndefined()
    expect(keyInfo('no.such')).toBeUndefined()
  })

  it('keeps subsection case and handles dotted subsections', () => {
    expect(canonicalKey('Remote.Origin.URL')).toBe('remote.Origin.url')
    expect(sectionOf('url.https://a.example/.insteadOf')).toBe('url "https://a.example/"')
    expect(sectionOf('core.bare')).toBe('core')
  })

  it('reads booleans the way git does', () => {
    expect(parseBool('Yes')).toBe(true)
    expect(parseBool('off')).toBe(false)
    expect(parseBool('', true)).toBe(true)
    expect(parseBool('maybe')).toBeUndefined()
  })
})

describe('buildRows', () => {
  const effective = [
    e('global', GLOBAL, 'user.name', 'Global Me'),
    e('global', GLOBAL, 'pull.rebase', 'false'),
    e('global', GLOBAL, 'credential.helper', 'manager'),
    e('local', 'C:/repo/.git/config', 'user.name', 'stale on disk'),
  ]
  const own = [e('', LOCAL, 'user.name', 'Local Me'), e('', LOCAL, 'credential.helper', 'store')]
  const sections = buildRows(own, effective, LOCAL)
  const find = (key: string) =>
    sections.flatMap((s) => s.rows).find((r) => canonicalKey(r.key) === canonicalKey(key))!

  it('uses the working copy for the edited file and marks the rest inherited', () => {
    const name = find('user.name')
    expect(name.values.map((v) => v.value)).toEqual(['Local Me'])
    expect(name.inherited.map((i) => i.value)).toEqual(['Global Me'])
    expect(overrideNote(name)).toBe('overrides global: Global Me')
  })

  it('shows inherited-only keys and notes multi-valued keys as adding', () => {
    expect(find('pull.rebase').values).toEqual([])
    expect(overrideNote(find('pull.rebase'))).toBe('')
    expect(find('credential.helper').multi).toBe(true)
    expect(overrideNote(find('credential.helper'))).toBe('adds to global')
  })

  it('treats a key set in several files as an override, not a list', () => {
    const rows = buildRows(
      [e('', LOCAL, 'core.autocrlf', 'false'), e('', LOCAL, 'x.multi', 'a'), e('', LOCAL, 'x.multi', 'b')],
      [e('system', 'C:/git/etc/gitconfig', 'core.autocrlf', 'true'), e('global', GLOBAL, 'core.autocrlf', 'input'), e('global', GLOBAL, 'y.list', '1'), e('global', GLOBAL, 'y.list', '2')],
      LOCAL,
    ).flatMap((s) => s.rows)
    const by = (k: string) => rows.find((r) => r.key === k)!
    expect(by('core.autocrlf').multi).toBe(false)
    expect(overrideNote(by('core.autocrlf'))).toBe('overrides global: input')
    expect(by('x.multi').multi).toBe(true)
    expect(by('y.list').multi).toBe(true)
  })

  it('groups by section', () => {
    expect(sections.map((s) => s.name)).toEqual(['user', 'credential', 'pull'])
  })

  it('filters by key or value and can hide inherited rows', () => {
    expect(filterSections(sections, 'rebase', true).flatMap((s) => s.rows.map((r) => r.key))).toEqual(['pull.rebase'])
    expect(filterSections(sections, 'global me', true).flatMap((s) => s.rows.map((r) => r.key))).toEqual(['user.name'])
    expect(filterSections(sections, 'global me', false)).toEqual([])
    expect(filterSections(sections, '', false).flatMap((s) => s.rows.map((r) => r.key))).toEqual(['user.name', 'credential.helper'])
  })
})

describe('masking', () => {
  it('masks URL credentials', () => {
    expect(maskUrls('https://me:hunter2@example.com/r.git')).toBe('https://me:••••@example.com/r.git')
    expect(maskUrls('https://ghp_abc123@github.com/')).toBe('https://••••@github.com/')
    expect(maskUrls('git@github.com:me/r.git')).toBe('git@github.com:me/r.git')
    expect(maskUrls('https://example.com/a@b')).toBe('https://example.com/a@b')
  })

  it('masks secret keys entirely', () => {
    expect(maskValue('http.https://x.example/.extraHeader', 'Authorization: Bearer t')).toBe('••••')
    expect(maskValue('user.name', 'Me')).toBe('Me')
    expect(hasSecret('url.https://me:tok@host/.insteadOf', 'https://host/')).toBe(true)
    expect(hasSecret('user.name', 'Me')).toBe(false)
  })
})

describe('samePath', () => {
  it('ignores slash direction, case and trailing slashes', () => {
    expect(samePath('C:\\Repo\\.git\\config', 'c:/repo/.git/config')).toBe(true)
    expect(samePath('/a/b', '/a/c')).toBe(false)
  })
})
