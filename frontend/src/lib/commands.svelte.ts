import { SvelteMap } from 'svelte/reactivity'
import { isTyping, matchesShortcut } from './shortcuts'

export interface Command {
  /** Stable id, unique across providers; also the list key. */
  id: string
  title: string
  /** The palette section it's listed under. */
  group: string
  /** Extra words to match on that the title doesn't show. */
  keywords?: string
  /** Shortcut in Mod+Shift+K form ("Mod" is Ctrl, or ⌘ on macOS), shown beside it. */
  shortcut?: string
  /** Other key combinations that run it, not shown. */
  aliases?: string[]
  /** Bound to its keys but not listed in the palette. */
  hidden?: boolean
  /** Its keys work while typing in a field (most shortcuts leave text editing alone). */
  whenTyping?: boolean
  /** Its keys work even while a dialog or menu is open. */
  global?: boolean
  run: () => unknown
}

type Provider = () => Command[]

export const isMac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform)

/** "Mod+Shift+Z" → "Ctrl+Shift+Z", or "⌘⇧Z" on macOS. */
export function formatShortcut(s: string): string {
  if (!isMac) return s.replace(/\bMod\b/g, 'Ctrl')
  const glyph: Record<string, string> = { Mod: '⌘', Shift: '⇧', Alt: '⌥', Ctrl: '⌃', Enter: '↩' }
  return s
    .split('+')
    .map((k) => glyph[k] ?? k)
    .join('')
}

/** Commands contributed by whatever is on screen, and the palette's open state. */
class Commands {
  private providers = new SvelteMap<symbol, Provider>()
  paletteOpen = $state(false)
  shortcutsOpen = $state(false)
  /** The clone dialog lives at the top level so the palette can open it from anywhere. */
  cloneOpen = $state(false)

  /** Adds a provider; returns the function that removes it. */
  add(provider: Provider): () => void {
    const key = Symbol()
    this.providers.set(key, provider)
    return () => this.providers.delete(key)
  }

  /** Every command available right now. Providers are read on demand, so titles reflect
   * current state (the undo label, the checked-out branch). */
  all(): Command[] {
    const seen: Record<string, true> = {}
    const out: Command[] = []
    for (const p of this.providers.values()) {
      for (const c of p()) {
        if (seen[c.id]) continue
        seen[c.id] = true
        out.push(c)
      }
    }
    return out
  }

  /** Runs the command bound to a key event, if any. Returns whether one ran. */
  dispatch(e: KeyboardEvent): boolean {
    if (e.defaultPrevented) return false
    const overlay = !!document.querySelector('[aria-modal="true"], [role="menu"]')
    const typing = isTyping(e.target)
    for (const c of this.all()) {
      if (!c.shortcut && !c.aliases) continue
      if ((overlay && !c.global) || (typing && !c.whenTyping && !c.global)) continue
      const keys = c.shortcut ? [c.shortcut, ...(c.aliases ?? [])] : (c.aliases ?? [])
      if (!keys.some((k) => matchesShortcut(e, k, isMac))) continue
      e.preventDefault()
      void c.run()
      return true
    }
    return false
  }
}

export const commands = new Commands()

/** Contributes commands while the calling component is mounted. Call during component setup. */
export function provideCommands(provider: Provider) {
  $effect(() => commands.add(provider))
}
