/**
 * Parsing and resolving git conflict markers in a merged file:
 *
 *   <<<<<<< ours-label
 *   ours
 *   ||||||| base-label        (only with merge.conflictStyle=diff3 / zdiff3)
 *   base
 *   =======
 *   theirs
 *   >>>>>>> theirs-label
 */

export interface ConflictRegion {
  /** Character offsets of the whole block, markers included. */
  from: number
  to: number
  ours: string
  base: string | null
  theirs: string
  oursLabel: string
  theirsLabel: string
}

export type Choice = 'ours' | 'theirs' | 'both' | 'base'

const START = /^<{7}(?: (.*))?$/
const BASE = /^\|{7}(?: .*)?$/
const MID = /^={7}$/
const END = /^>{7}(?: (.*))?$/

/** Finds every complete conflict block; malformed fragments are left alone. */
export function parseConflicts(text: string): ConflictRegion[] {
  const regions: ConflictRegion[] = []
  // Walk line by line, tracking offsets so blocks can be replaced in place.
  const lines = text.match(/[^\n]*\n|[^\n]+$/g) ?? []
  let offset = 0
  let cur: { from: number; oursLabel: string; part: 'ours' | 'base' | 'theirs'; ours: string; base: string | null; theirs: string } | null = null
  for (const raw of lines) {
    const line = raw.replace(/\r?\n$/, '')
    const start = START.exec(line)
    if (start) {
      cur = { from: offset, oursLabel: start[1] ?? '', part: 'ours', ours: '', base: null, theirs: '' }
    } else if (cur && cur.part === 'ours' && BASE.test(line)) {
      cur.part = 'base'
      cur.base = ''
    } else if (cur && cur.part !== 'theirs' && MID.test(line)) {
      cur.part = 'theirs'
    } else if (cur && cur.part === 'theirs' && END.test(line)) {
      regions.push({
        from: cur.from,
        to: offset + raw.length,
        ours: cur.ours,
        base: cur.base,
        theirs: cur.theirs,
        oursLabel: cur.oursLabel,
        theirsLabel: END.exec(line)?.[1] ?? '',
      })
      cur = null
    } else if (cur) {
      if (cur.part === 'ours') cur.ours += raw
      else if (cur.part === 'base') cur.base += raw
      else cur.theirs += raw
    }
    offset += raw.length
  }
  return regions
}

/** The text a region resolves to for a choice. "both" keeps ours then theirs. */
export function resolution(region: ConflictRegion, choice: Choice): string {
  switch (choice) {
    case 'ours':
      return region.ours
    case 'theirs':
      return region.theirs
    case 'both':
      return region.ours + region.theirs
    case 'base':
      return region.base ?? ''
  }
}

/** Replaces one region in text with the chosen side. */
export function resolveRegion(text: string, region: ConflictRegion, choice: Choice): string {
  return text.slice(0, region.from) + resolution(region, choice) + text.slice(region.to)
}

/** True if text still contains a conflict marker line. */
export function hasMarkers(text: string): boolean {
  return /^(<{7}|>{7})( |$)/m.test(text)
}
