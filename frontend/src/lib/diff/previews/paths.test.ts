import { describe, expect, it } from 'vitest'
import { resolveInRepo } from './paths'

describe('resolveInRepo', () => {
  it('resolves against the document folder', () => {
    expect(resolveInRepo('docs/guide.md', 'img/a.png')).toBe('docs/img/a.png')
    expect(resolveInRepo('docs/guide.md', './b.md')).toBe('docs/b.md')
    expect(resolveInRepo('docs/guide.md', '../README.md')).toBe('README.md')
    expect(resolveInRepo('README.md', 'docs/PLAN.md#goals')).toBe('docs/PLAN.md')
  })

  it('treats a leading slash as the repository root', () => {
    expect(resolveInRepo('docs/deep/x.md', '/build/appicon.svg')).toBe('build/appicon.svg')
  })

  it('decodes escaped characters', () => {
    expect(resolveInRepo('a.md', 'my%20file.md')).toBe('my file.md')
  })

  it('refuses web links, fragments and escapes from the repository', () => {
    expect(resolveInRepo('a.md', 'https://example.com/x.png')).toBeNull()
    expect(resolveInRepo('a.md', '//cdn.example.com/x.png')).toBeNull()
    expect(resolveInRepo('a.md', '#section')).toBeNull()
    expect(resolveInRepo('docs/a.md', '../../etc/passwd')).toBeNull()
    expect(resolveInRepo('a.md', 'javascript:alert(1)')).toBeNull()
  })
})
