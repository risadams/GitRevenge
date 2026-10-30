import { describe, expect, it } from 'vitest'
import { wantsNativeMenu } from './nativeMenu'

const div = { tagName: 'DIV', isContentEditable: false }

describe('wantsNativeMenu', () => {
  it('drops the menu on plain elements', () => {
    expect(wantsNativeMenu(div, false, '')).toBe(false)
  })

  it('keeps it for editable fields but not read-only or disabled ones', () => {
    expect(wantsNativeMenu({ tagName: 'INPUT', isContentEditable: false }, false, '')).toBe(true)
    expect(wantsNativeMenu({ tagName: 'TEXTAREA', isContentEditable: false }, false, '')).toBe(true)
    expect(wantsNativeMenu({ tagName: 'INPUT', isContentEditable: false, readOnly: true }, false, '')).toBe(false)
    expect(wantsNativeMenu({ tagName: 'INPUT', isContentEditable: false, disabled: true }, false, '')).toBe(false)
    expect(wantsNativeMenu({ tagName: 'DIV', isContentEditable: true }, false, '')).toBe(true)
  })

  it('keeps it when text is selected or Shift is held', () => {
    expect(wantsNativeMenu(div, false, 'abc')).toBe(true)
    expect(wantsNativeMenu(div, true, '')).toBe(true)
  })
})
