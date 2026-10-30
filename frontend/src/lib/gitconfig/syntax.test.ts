import { describe, expect, it } from 'vitest'
import { highlightLines } from '../diff/highlight'
import { gitConfigSyntax } from './syntax'

/** Each non-blank token of each line as "text:class". */
function tokens(text: string): string[][] {
  return highlightLines(text, gitConfigSyntax()).map((line) => line.filter((t) => t.text.trim()).map((t) => `${t.text.trim()}:${t.cls}`))
}

describe('git config highlighting', () => {
  it('tells headers, keys, values and comments apart', () => {
    const [header, key, bool, comment, num] = tokens('[remote "origin"]\n\turl = https://example.com/r.git\n\tprune = yes\n# note\n\tdepth = 50 ; trailing')
    expect(header).toEqual(['[:tok-punctuation', 'remote:tok-heading', '"origin":tok-string', ']:tok-punctuation'])
    expect(key).toEqual(['url:tok-propertyName', '=:tok-operator', 'https://example.com/r.git:'])
    expect(bool).toEqual(['prune:tok-propertyName', '=:tok-operator', 'yes:tok-bool'])
    expect(comment).toEqual(['# note:tok-comment'])
    expect(num).toEqual(['depth:tok-propertyName', '=:tok-operator', '50:tok-number', '; trailing:tok-comment'])
  })

  it('treats a key without a value as a key', () => {
    expect(tokens('[core]\n\tbare')[1]).toEqual(['bare:tok-propertyName'])
  })
})
