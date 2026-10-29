import { describe, expect, it } from 'vitest'
import { asPerson, valueParts } from './trailers'

describe('asPerson', () => {
  it('reads "Name <email>"', () => {
    expect(asPerson('Ada Lovelace <ada@example.com>')).toEqual({ name: 'Ada Lovelace', email: 'ada@example.com' })
  })

  it('rejects anything else', () => {
    expect(asPerson('#12')).toBeNull()
    expect(asPerson('<ada@example.com>')).toBeNull()
    expect(asPerson('Ada <not an email>')).toBeNull()
  })
})

describe('valueParts', () => {
  const issues = 'https://github.com/o/r/issues/{n}'

  it('links issue references when the pattern is known', () => {
    expect(valueParts('#1, #22', issues)).toEqual([
      { text: '#1', href: 'https://github.com/o/r/issues/1' },
      { text: ', ' },
      { text: '#22', href: 'https://github.com/o/r/issues/22' },
    ])
  })

  it('leaves issue references as text without one', () => {
    expect(valueParts('#7', '')).toEqual([{ text: '#7' }])
  })

  it('links URLs, leaving trailing punctuation out', () => {
    expect(valueParts('see https://example.com/a?b=1.', '')).toEqual([
      { text: 'see ' },
      { text: 'https://example.com/a?b=1', href: 'https://example.com/a?b=1' },
      { text: '.' },
    ])
  })

  it('returns plain text as one part', () => {
    expect(valueParts('no links here', issues)).toEqual([{ text: 'no links here' }])
  })
})
