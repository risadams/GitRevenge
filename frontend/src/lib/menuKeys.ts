/** Keyboard movement through a menu's items, skipping disabled ones. */

interface Navigable {
  label: string
  disabled?: boolean
}

/** The next enabled item after from in direction dir (1 or -1), wrapping; -1 if none. */
export function step(items: Navigable[], from: number, dir: 1 | -1): number {
  const n = items.length
  for (let k = 1; k <= n; k++) {
    const j = (((from + dir * k) % n) + n) % n
    if (!items[j].disabled) return j
  }
  return -1
}

/** The first or last enabled item; -1 if none. */
export function edge(items: Navigable[], last: boolean): number {
  return last ? step(items, 0, -1) : step(items, -1, 1)
}

/**
 * Type-ahead: the next enabled item whose label starts with query, searching after from for a
 * single letter (so repeating it cycles) and from from itself for a longer query; -1 if none.
 */
export function typeahead(items: Navigable[], from: number, query: string): number {
  const n = items.length
  const q = query.toLowerCase()
  const start = q.length > 1 ? Math.max(from, 0) : from + 1
  for (let k = 0; k < n; k++) {
    const j = (start + k + n) % n
    if (!items[j].disabled && items[j].label.toLowerCase().startsWith(q)) return j
  }
  return -1
}
