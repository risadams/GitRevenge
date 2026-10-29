import { describe, expect, it } from 'vitest'
import { isTyping, matchesShortcut, parseShortcut } from './shortcuts'

const ev = (key: string, mods: Partial<Record<'ctrl' | 'meta' | 'shift' | 'alt', boolean>> = {}, code = '') => ({
  key,
  code,
  ctrlKey: !!mods.ctrl,
  metaKey: !!mods.meta,
  shiftKey: !!mods.shift,
  altKey: !!mods.alt,
})

describe('parseShortcut', () => {
  it('splits modifiers from the key', () => {
    expect(parseShortcut('Mod+Shift+Z')).toEqual({ mod: true, shift: true, alt: false, key: 'z' })
    expect(parseShortcut('Escape')).toEqual({ mod: false, shift: false, alt: false, key: 'Escape' })
  })
})

describe('matchesShortcut', () => {
  it('maps Mod to Ctrl, or to ⌘ on macOS', () => {
    expect(matchesShortcut(ev('z', { ctrl: true }), 'Mod+Z', false)).toBe(true)
    expect(matchesShortcut(ev('z', { meta: true }), 'Mod+Z', true)).toBe(true)
    expect(matchesShortcut(ev('z', { meta: true }), 'Mod+Z', false)).toBe(false)
    expect(matchesShortcut(ev('z', { ctrl: true }), 'Mod+Z', true)).toBe(false)
  })

  it('needs the exact modifiers', () => {
    expect(matchesShortcut(ev('Z', { ctrl: true, shift: true }), 'Mod+Z', false)).toBe(false)
    expect(matchesShortcut(ev('Z', { ctrl: true, shift: true }), 'Mod+Shift+Z', false)).toBe(true)
    expect(matchesShortcut(ev('z', { ctrl: true, alt: true }), 'Mod+Z', false)).toBe(false)
  })

  it('matches Alt+letter by the physical key, since Option types a symbol on macOS', () => {
    expect(matchesShortcut(ev('Ω', { alt: true }, 'KeyZ'), 'Alt+Z', true)).toBe(true)
    expect(matchesShortcut(ev('z', { alt: true }, 'KeyZ'), 'Alt+Z', false)).toBe(true)
    expect(matchesShortcut(ev('Ω', { alt: true }, 'KeyX'), 'Alt+Z', true)).toBe(false)
  })

  it('matches digits by the physical key', () => {
    expect(matchesShortcut(ev('!', { ctrl: true, shift: true }, 'Digit1'), 'Mod+Shift+1', false)).toBe(true)
    expect(matchesShortcut(ev('1', { ctrl: true }, 'Digit1'), 'Mod+1', false)).toBe(true)
    expect(matchesShortcut(ev('2', { ctrl: true }, 'Digit2'), 'Mod+1', false)).toBe(false)
  })

  it('matches named keys', () => {
    expect(matchesShortcut(ev('Tab', { ctrl: true, shift: true }), 'Mod+Shift+Tab', false)).toBe(true)
    expect(matchesShortcut(ev('Escape'), 'Escape', false)).toBe(true)
    expect(matchesShortcut(ev('/', { ctrl: true }), 'Mod+/', false)).toBe(true)
  })
})

describe('isTyping', () => {
  it('spots text fields and editable content', () => {
    expect(isTyping({ closest: (s: string) => (s.includes('input') ? {} : null) } as unknown as EventTarget)).toBe(true)
    expect(isTyping({ closest: () => null } as unknown as EventTarget)).toBe(false)
    expect(isTyping(null)).toBe(false)
  })
})
