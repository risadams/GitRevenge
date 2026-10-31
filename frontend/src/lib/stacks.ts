import type { Ref } from './graph/rowCache'

/** Orders branches so each stacked branch follows its parent, parents before children and
 * siblings in the given order. Branches whose parent isn't in the list stay where they are. */
export function stackOrder(refs: Ref[]): Ref[] {
  const byName = new Map(refs.map((r) => [r.short, r]))
  const children = new Map<string, Ref[]>()
  const roots: Ref[] = []
  for (const r of refs) {
    const p = r.stackParent
    if (p && byName.has(p) && p !== r.short) {
      const list = children.get(p) ?? []
      list.push(r)
      children.set(p, list)
    } else {
      roots.push(r)
    }
  }
  const out: Ref[] = []
  const seen = new Set<string>()
  const walk = (r: Ref) => {
    if (seen.has(r.short)) return
    seen.add(r.short)
    out.push(r)
    for (const c of children.get(r.short) ?? []) walk(c)
  }
  for (const r of roots) walk(r)
  // Anything left (a loop that slipped through) keeps its place at the end.
  for (const r of refs) if (!seen.has(r.short)) out.push(r)
  return out
}

/** "2 of 3", for a branch in a stack; '' otherwise. */
export function stackLabel(r: Ref): string {
  return r.stackDepth && r.stackSize ? `${r.stackDepth} of ${r.stackSize}` : ''
}
