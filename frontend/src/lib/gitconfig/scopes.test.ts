import { describe, expect, it } from 'vitest'
import { baseName, scopeLabel } from './scopes'

describe('scopeLabel', () => {
  const tabs = [{ path: 'C:\\code\\app', name: 'app' }]

  it('names each kind of scope', () => {
    expect(scopeLabel({ kind: 'global', path: 'C:\\Users\\me\\.gitconfig', repo: '' }, tabs)).toBe('Global')
    expect(scopeLabel({ kind: 'local', path: 'C:\\code\\app\\.git\\config', repo: 'C:\\code\\app' }, tabs)).toBe('Local · app')
    expect(scopeLabel({ kind: 'file', path: '/home/me/work/.gitconfig', repo: '' }, tabs)).toBe('File · .gitconfig')
  })

  it('falls back to the folder name for a repository without a tab', () => {
    expect(scopeLabel({ kind: 'local', path: '/src/lib/.git/config', repo: '/src/lib' }, tabs)).toBe('Local · lib')
  })
})

describe('baseName', () => {
  it('handles both separators and trailing ones', () => {
    expect(baseName('C:\\a\\b\\')).toBe('b')
    expect(baseName('/a/b')).toBe('b')
  })
})
