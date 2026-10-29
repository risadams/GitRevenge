import { describe, expect, it } from 'vitest'
import { formatHours, textToBase64, toCSV } from './export'

describe('toCSV', () => {
  it('quotes only fields that need it', () => {
    expect(toCSV([['name', 'commits'], ['Ada, Countess', 12], ['Say "hi"', 3], ['two\nlines', 0]])).toBe(
      'name,commits\r\n"Ada, Countess",12\r\n"Say ""hi""",3\r\n"two\nlines",0\r\n',
    )
  })
})

describe('textToBase64', () => {
  it('encodes UTF-8', () => {
    expect(atob(textToBase64('a'))).toBe('a')
    expect(textToBase64('Järvinen')).toBe(btoa(String.fromCharCode(...new TextEncoder().encode('Järvinen'))))
  })
})

describe('formatHours', () => {
  it('picks a readable unit', () => {
    expect([0.5, 5.5, 30, 100].map(formatHours)).toEqual(['30 min', '5.5 h', '30 h', '4.2 days'])
  })
})
