import type { Command } from '../commands.svelte'

const KEY = 'gitrevenge.wordWrap'

function saved(): boolean {
  try {
    return localStorage.getItem(KEY) === 'on'
  } catch {
    return false
  }
}

/** Word wrap for the diff and blame views: one setting for all of them, remembered across sessions. */
export const wrap = $state({ on: saved() })

export function toggleWrap() {
  wrap.on = !wrap.on
  try {
    localStorage.setItem(KEY, wrap.on ? 'on' : 'off')
  } catch {
    // Storage unavailable; the choice lasts for this session only.
  }
}

/** The palette entry and Alt+Z, provided by each view that wraps. */
export function wrapCommand(): Command {
  return {
    id: 'view.wrap',
    group: 'View',
    title: wrap.on ? 'Turn off word wrap' : 'Turn on word wrap',
    keywords: 'wrap long lines soft',
    shortcut: 'Alt+Z',
    run: toggleWrap,
  }
}
