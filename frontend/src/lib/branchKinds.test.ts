import { describe, expect, it } from 'vitest'
import { CONFIRM_DELETE, CONFIRM_FORCE_PUSH, kindInfo, kindWarning } from './branchKinds'

describe('kindWarning', () => {
  it('names the kind for kinds that ask', () => {
    expect(kindWarning('main', 'trunk', CONFIRM_FORCE_PUSH)).toBe('main is a Trunk branch.')
    expect(kindWarning('release/1.4', 'release', CONFIRM_FORCE_PUSH)).toBe('release/1.4 is a Release branch.')
    expect(kindWarning('production', 'environment', CONFIRM_DELETE)).toBe('production is an Environment branch.')
  })

  it('stays quiet for kinds that do not ask, and ordinary branches', () => {
    expect(kindWarning('release/1.4', 'release', CONFIRM_DELETE)).toBe('')
    expect(kindWarning('gh-pages', 'docs', CONFIRM_FORCE_PUSH)).toBe('')
    expect(kindWarning('dependabot/x', 'bot', CONFIRM_DELETE)).toBe('')
    expect(kindWarning('feature/x', '', CONFIRM_FORCE_PUSH)).toBe('')
    expect(kindWarning('feature/x', undefined, CONFIRM_DELETE)).toBe('')
  })

  it('knows every kind', () => {
    for (const k of ['trunk', 'release', 'environment', 'docs', 'bot']) expect(kindInfo(k)).toBeDefined()
    expect(kindInfo('nightly')).toBeUndefined()
  })
})
