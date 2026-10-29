import { SvelteMap } from 'svelte/reactivity'

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
}

export const commands = new Commands()

/** Contributes commands while the calling component is mounted. Call during component setup. */
export function provideCommands(provider: Provider) {
  $effect(() => commands.add(provider))
}
