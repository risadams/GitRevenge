export interface Match {
  score: number
  /** Indexes into the text of the characters that matched, for highlighting. */
  positions: number[]
}

const BOUNDARY = /[\s\-_./\\:·()[\]{}<>"']/

function isWordStart(text: string, i: number): boolean {
  if (i === 0) return true
  const prev = text[i - 1]
  if (BOUNDARY.test(prev)) return true
  // camelCase and PascalCase humps.
  return prev === prev.toLowerCase() && text[i] !== text[i].toLowerCase()
}

/**
 * Fuzzy-matches query against text: every query character must appear in order. Word
 * starts, runs of consecutive characters and an early first match score higher, so "pf"
 * ranks "Pull (fast-forward only)" above "Push force". Spaces in the query are ignored.
 * Returns null when the text doesn't match.
 */
export function fuzzy(query: string, text: string): Match | null {
  const q = query.toLowerCase().replace(/\s+/g, '')
  if (!q) return { score: 0, positions: [] }
  const t = text.toLowerCase()

  const positions: number[] = []
  let from = 0
  for (const c of q) {
    // Prefer the next occurrence that starts a word, if one comes before a plain one would
    // waste it; otherwise take the nearest.
    const near = t.indexOf(c, from)
    if (near < 0) return null
    let pick = near
    if (!isWordStart(text, near)) {
      for (let i = near + 1; i < t.length; i++) {
        if (t[i] === c && isWordStart(text, i)) {
          pick = i
          break
        }
      }
      // Don't jump ahead if that would make the rest of the query unmatchable.
      if (pick !== near && !matchesFrom(t, q.slice(positions.length + 1), pick + 1)) pick = near
    }
    positions.push(pick)
    from = pick + 1
  }

  let score = 0
  for (let k = 0; k < positions.length; k++) {
    const i = positions[k]
    score += 1
    if (isWordStart(text, i)) score += 3
    if (k > 0 && positions[k - 1] === i - 1) score += 2
    if (k > 0) score -= Math.min(3, (i - positions[k - 1] - 1) * 0.1)
  }
  score -= positions[0] * 0.05
  if (t.includes(q)) score += 4
  if (t.startsWith(q)) score += 4
  // Among equals, shorter texts are closer matches.
  score -= text.length * 0.01
  return { score, positions }
}

function matchesFrom(t: string, rest: string, from: number): boolean {
  for (const c of rest) {
    const i = t.indexOf(c, from)
    if (i < 0) return false
    from = i + 1
  }
  return true
}
