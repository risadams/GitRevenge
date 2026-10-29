/** A key combination such as "Mod+Shift+Z". "Mod" is Ctrl, or ⌘ on macOS. */
export interface Keys {
  mod: boolean
  shift: boolean
  alt: boolean
  /** Lower-case for single characters ("z", "1", "/"), otherwise the key name ("Tab"). */
  key: string
}

export function parseShortcut(s: string): Keys {
  const parts = s.split('+')
  const key = parts.pop() ?? ''
  return {
    mod: parts.includes('Mod'),
    shift: parts.includes('Shift'),
    alt: parts.includes('Alt'),
    key: key.length === 1 ? key.toLowerCase() : key,
  }
}

type KeyEventLike = Pick<KeyboardEvent, 'key' | 'code' | 'ctrlKey' | 'metaKey' | 'shiftKey' | 'altKey'>

/** Whether a key event is the given shortcut. mac picks ⌘ over Ctrl for "Mod". */
export function matchesShortcut(e: KeyEventLike, shortcut: string, mac: boolean): boolean {
  const k = parseShortcut(shortcut)
  const mod = mac ? e.metaKey : e.ctrlKey
  // The other platform's modifier must not be held, so ⌃Z on a Mac isn't taken as ⌘Z.
  const stray = mac ? e.ctrlKey : e.metaKey
  if (mod !== k.mod || stray || e.shiftKey !== k.shift || e.altKey !== k.alt) return false
  // Shift turns digits into symbols on most layouts; the physical key still says which.
  if (/^[0-9]$/.test(k.key)) return e.code === `Digit${k.key}` || e.key === k.key
  const key = e.key.length === 1 ? e.key.toLowerCase() : e.key
  if (key === k.key) return true
  // Option on macOS turns letters into symbols (⌥Z types Ω); the physical key still says which.
  return k.alt && /^[a-z]$/.test(k.key) && e.code === `Key${k.key.toUpperCase()}`
}

/** Whether the event comes from somewhere text is typed, where most shortcuts stand aside. */
export function isTyping(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null
  return !!el?.closest?.('input, textarea, select, [contenteditable=""], [contenteditable="true"]')
}
