import { diffLines, splitLines, toHunks } from '../diff/lineDiff'

// The History view's model (#228): backups grouped by day, and a backup compared with the
// file as it is now.

export interface DiffRow {
  kind: 'eq' | 'del' | 'add' | 'gap'
  /** The old (backup) line, for eq and del; the new line for add. */
  text: string
  /** For split rows: the line on the right, or undefined when the left stands alone. */
  right?: string
}

const lines = (text: string) => splitLines(text.replace(/\r\n/g, '\n')).map((l) => l.replace(/\n$/, ''))

/** A unified diff from a backup to the current text: changed lines with three lines of
 * context, and a gap row between hunks. */
export function unifiedRows(backup: string, now: string): DiffRow[] {
  const a = lines(backup)
  const b = lines(now)
  const ops = diffLines(a, b)
  const rows: DiffRow[] = []
  toHunks(ops).forEach((h, i) => {
    if (i > 0 || h.start > 0) rows.push({ kind: 'gap', text: '' })
    for (const op of ops.slice(h.start, h.end)) {
      rows.push({ kind: op.kind, text: op.kind === 'add' ? b[op.newIndex] : a[op.oldIndex] })
    }
  })
  return rows
}

/** The same changes side by side: a removed line pairs with the added line that replaced it. */
export function splitRows(backup: string, now: string): DiffRow[] {
  const out: DiffRow[] = []
  const rows = unifiedRows(backup, now)
  for (let i = 0; i < rows.length; i++) {
    const r = rows[i]
    if (r.kind !== 'del') {
      out.push(r.kind === 'eq' ? { ...r, right: r.text } : r.kind === 'add' ? { kind: 'add', text: '', right: r.text } : r)
      continue
    }
    // Pair a run of removals with the run of additions after it.
    const dels: string[] = []
    while (i < rows.length && rows[i].kind === 'del') dels.push(rows[i++].text)
    const adds: string[] = []
    while (i < rows.length && rows[i].kind === 'add') adds.push(rows[i++].text)
    i--
    for (let k = 0; k < Math.max(dels.length, adds.length); k++) {
      out.push({ kind: k < dels.length ? 'del' : 'add', text: dels[k] ?? '', right: adds[k] })
    }
  }
  return out
}

/** How many lines a backup and the current text differ by. */
export function changedLines(backup: string, now: string): number {
  return unifiedRows(backup, now).filter((r) => r.kind === 'del' || r.kind === 'add').length
}

/** A heading for the day a backup was made: Today, Yesterday, or a date. */
export function dayLabel(when: Date, now = new Date()): string {
  const day = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
  const diff = Math.round((day(now) - day(when)) / 86_400_000)
  if (diff === 0) return 'Today'
  if (diff === 1) return 'Yesterday'
  return when.toLocaleDateString(undefined, { month: 'short', day: 'numeric', ...(when.getFullYear() !== now.getFullYear() && { year: 'numeric' }) })
}
