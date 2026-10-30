import { describe, expect, it } from 'vitest'
import { buildAliases, callsItself, NAME } from './aliases'
import type { Entry } from './rows'

const e = (scope: string, file: string, key: string, value: string): Entry => ({ scope, file, key, value, implicit: false })
const LOCAL = 'C:\\repo\\.git\\config'
const GLOBAL = 'C:/Users/me/.gitconfig'

describe('buildAliases', () => {
  const list = buildAliases(
    [e('', LOCAL, 'alias.co', 'checkout'), e('', LOCAL, 'alias.log', 'log --oneline -20'), e('', LOCAL, 'alias.wip', '!git add -A && git commit -m wip'), e('', LOCAL, 'user.name', 'x')],
    [e('global', GLOBAL, 'alias.co', 'checkout -q'), e('global', GLOBAL, 'alias.st', 'status -sb'), e('local', 'c:/repo/.git/config', 'alias.co', 'stale')],
    LOCAL,
    ['log', 'status', 'checkout'],
  )
  const by = (n: string) => list.find((a) => a.name === n)!

  it('lists own aliases first, then inherited ones this file does not set', () => {
    expect(list.map((a) => `${a.name}:${a.own}`)).toEqual(['co:true', 'log:true', 'wip:true', 'st:false'])
    expect(by('st')).toMatchObject({ scope: 'global', value: 'status -sb' })
  })

  it('flags shell aliases and what goes wrong', () => {
    expect(by('wip').shell).toBe(true)
    expect(by('co').warnings).toEqual(['Replaces the global alias co, which runs checkout -q.'])
    expect(by('log').warnings).toHaveLength(1)
    expect(by('log').warnings[0]).toMatch(/built-in command/)
    expect(by('st').warnings).toEqual([])
  })
})

describe('callsItself', () => {
  it('spots a git or shell alias running itself', () => {
    expect(callsItself({ name: 'redo', value: '!git redo --hard' })).toBe(true)
    expect(callsItself({ name: 'up', value: '!git fetch && git up' })).toBe(true)
    expect(callsItself({ name: 'up', value: '!git fetch && git upstream' })).toBe(false)
    expect(callsItself({ name: 'lg', value: 'lg --all' })).toBe(true)
    expect(callsItself({ name: 'lg', value: 'log --graph' })).toBe(false)
  })
})

describe('NAME', () => {
  it('follows git variable naming', () => {
    expect(NAME.test('push-f')).toBe(true)
    expect(NAME.test('1st')).toBe(false)
    expect(NAME.test('a.b')).toBe(false)
  })
})
