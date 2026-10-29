import { describe, expect, it } from 'vitest'
import { reflow } from './text'

describe('reflow', () => {
  it('joins the hard-wrapped lines of a paragraph', () => {
    expect(reflow('The merge flag: a flag on a mast with\ntwo branches merging into one\ncommit.')).toBe(
      'The merge flag: a flag on a mast with two branches merging into one commit.',
    )
  })

  it('keeps paragraphs apart', () => {
    expect(reflow('First line\nwrapped.\n\nSecond\nparagraph.')).toBe('First line wrapped.\n\nSecond paragraph.')
  })

  it('keeps list items, joining their wrapped continuations', () => {
    expect(reflow('- one item that\n  wraps\n- two\n1. numbered')).toBe('- one item that wraps\n- two\n1. numbered')
  })

  it('leaves quotes, headings, indented code and trailers alone', () => {
    const body = '## What\n> quoted\n    code line\n    more code\nCloses #12\nCo-Authored-By: Someone <a@b.c>'
    expect(reflow(body)).toBe(body)
  })

  it('normalizes Windows line endings', () => {
    expect(reflow('a\r\nb')).toBe('a b')
  })
})
