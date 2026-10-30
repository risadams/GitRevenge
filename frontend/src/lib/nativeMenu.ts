// WebView2's own context menu (Back, Reload, Print, Inspect…) is noise in an app. The Wails
// runtime hides it only in production builds, so this applies the same rule everywhere: keep
// it for editable text and selected text, where Cut/Copy/Paste are useful, and drop it
// elsewhere. Shift+right-click still opens it, which keeps Inspect reachable while developing.

export interface MenuTarget {
  tagName: string
  isContentEditable: boolean
  readOnly?: boolean
  disabled?: boolean
}

/** Whether a right-click on `target` should get the native menu. */
export function wantsNativeMenu(target: MenuTarget, shiftKey: boolean, selectedText: string): boolean {
  if (shiftKey || target.isContentEditable || selectedText.length > 0) return true
  if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return !target.readOnly && !target.disabled
  return false
}

/** Suppresses the native menu wherever the app hasn't handled the click and it isn't wanted. */
export function suppressNativeMenu(win: Window): void {
  win.addEventListener('contextmenu', (e) => {
    if (e.defaultPrevented || !(e.target instanceof HTMLElement)) return
    const input = e.target as HTMLInputElement
    // Only a selection under the pointer counts: copying it is what the menu would be for.
    const sel = win.getSelection()
    const selected = sel?.containsNode(input, true) ? sel.toString() : ''
    if (!wantsNativeMenu(input, e.shiftKey, selected)) e.preventDefault()
  })
}
