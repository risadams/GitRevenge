import { describe, expect, it } from 'vitest'
import { changedLines, dayLabel, splitRows, unifiedRows } from './history'

const backup = '[core]\r\n\tfilemode = true\r\n\tautocrlf = true\r\n[user]\r\n\temail = a@example.com\r\n'
const now = '[core]\n\tfilemode = true\n\tautocrlf = input\n[user]\n\temail = a@example.com\n\tname = A\n'

describe('history diff', () => {
  it('compares across line endings and keeps context', () => {
    expect(unifiedRows(backup, now).map((r) => `${r.kind} ${r.text}`)).toEqual([
      'eq [core]',
      'eq \tfilemode = true',
      'del \tautocrlf = true',
      'add \tautocrlf = input',
      'eq [user]',
      'eq \temail = a@example.com',
      'add \tname = A',
    ])
    expect(changedLines(backup, now)).toBe(3)
    expect(changedLines(now, now)).toBe(0)
  })

  it('pairs a replaced line with its replacement side by side', () => {
    const rows = splitRows(backup, now)
    expect(rows.find((r) => r.kind === 'del')).toEqual({ kind: 'del', text: '\tautocrlf = true', right: '\tautocrlf = input' })
    expect(rows.at(-1)).toEqual({ kind: 'add', text: '', right: '\tname = A' })
  })
})

describe('dayLabel', () => {
  const now = new Date(2026, 9, 30, 15, 0)
  it('names recent days and dates older ones', () => {
    expect(dayLabel(new Date(2026, 9, 30, 0, 5), now)).toBe('Today')
    expect(dayLabel(new Date(2026, 9, 29, 23, 59), now)).toBe('Yesterday')
    expect(dayLabel(new Date(2026, 9, 22), now)).not.toMatch(/Today|Yesterday|2026/)
    expect(dayLabel(new Date(2025, 9, 22), now)).toMatch(/2025/)
  })
})
