// The Insights tab's pure logic (#249): combining per-repository results and ordering them.
// Every order has a full tie-break, so the same results always display the same way
// (ADR-0008).

/** The fields of a shipped item the model reads (git.ShippedItem). */
export interface Item {
  sha: string
  title: string
  author: string
  email: string
  committed: number
  refs: string[]
  revert: boolean
  fixup: boolean
}

/** The fields of a branch signal the model reads (app.BranchSignal). */
export interface Branch {
  name: string
  sha: string
  author: string
  local: boolean
  lastCommit: number
  flags: string[]
}

/** A dry-run merge result (the insights:conflict event). */
export interface ConflictInfo {
  repo: string
  targetSha: string
  branch: string
  sha: string
  conflicts: boolean
  files: string[]
}

/** Something from one repository of the scope. */
export interface Tagged<T> {
  repo: string
  value: T
}

/** A branch with its conflict check, once known. */
export interface Ranked extends Tagged<Branch> {
  conflict?: ConflictInfo
}

const DAY = 24 * 60 * 60

/** Branches that need attention: a flag, or a conflict. Conflicts first, then the most
 * flags, then the most recent tip, then repository and name. */
export function attention(branches: Ranked[]): Ranked[] {
  return branches
    .filter((b) => b.value.flags.length > 0 || b.conflict?.conflicts)
    .sort((a, b) => {
      const ca = a.conflict?.conflicts ? 1 : 0
      const cb = b.conflict?.conflicts ? 1 : 0
      return (
        cb - ca ||
        b.value.flags.length - a.value.flags.length ||
        b.value.lastCommit - a.value.lastCommit ||
        a.repo.localeCompare(b.repo, 'en') ||
        a.value.name.localeCompare(b.value.name, 'en')
      )
    })
}

/** The newest shipped items first, ties by repository and SHA. */
export function latest(items: Tagged<Item>[]): Tagged<Item>[] {
  return [...items].sort((a, b) => b.value.committed - a.value.committed || a.repo.localeCompare(b.repo, 'en') || a.value.sha.localeCompare(b.value.sha, 'en'))
}

export interface Tiles {
  shipped: number
  people: number
  attention: number
  conflicts: number
  rework: number
}

export function tiles(items: Tagged<Item>[], branches: Ranked[]): Tiles {
  const flagged = attention(branches)
  return {
    shipped: items.length,
    people: new Set(items.map((i) => i.value.email.toLowerCase())).size,
    attention: flagged.length,
    conflicts: flagged.filter((b) => b.conflict?.conflicts).length,
    rework: items.filter((i) => i.value.revert || i.value.fixup).length,
  }
}

export interface DayCount {
  /** The UTC day's start, Unix seconds. */
  day: number
  count: number
}

/** Items per UTC day, for every day from since to until, empty days included. */
export function perDay(items: Tagged<Item>[], since: number, until: number): DayCount[] {
  const start = Math.floor(since / DAY) * DAY
  const days: DayCount[] = []
  for (let d = start; d <= until; d += DAY) days.push({ day: d, count: 0 })
  for (const i of items) {
    const at = Math.floor((i.value.committed - start) / DAY)
    if (at >= 0 && at < days.length) days[at].count++
  }
  return days
}

/** A UTC date as "Oct 16", with the year when it differs from yearOf's. */
export function utcDay(seconds: number, yearOf = seconds): string {
  const d = new Date(seconds * 1000)
  const sameYear = d.getUTCFullYear() === new Date(yearOf * 1000).getUTCFullYear()
  return d.toLocaleDateString('en', { month: 'short', day: 'numeric', ...(sameYear ? {} : { year: 'numeric' }), timeZone: 'UTC' })
}

/** A UTC time as "Oct 29, 17:42". */
export function utcMoment(seconds: number): string {
  const d = new Date(seconds * 1000)
  const date = d.toLocaleDateString('en', { month: 'short', day: 'numeric', timeZone: 'UTC' })
  const time = d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })
  return `${date}, ${time}`
}

/** How long ago a moment was, roughly, for fetch freshness: "2 hours ago". Display only. */
export function ago(seconds: number, now = Date.now() / 1000): string {
  const s = Math.max(0, now - seconds)
  if (s < 90) return 'just now'
  if (s < 90 * 60) return `${Math.round(s / 60)} minutes ago`
  if (s < 36 * 3600) return `${Math.round(s / 3600)} hours ago`
  return `${Math.round(s / DAY)} days ago`
}

/** The group scope's shared as-of point: the latest of its repositories' own. */
export function sharedAsOf(asOfs: number[]): number {
  return asOfs.reduce((m, v) => Math.max(m, v), 0)
}
