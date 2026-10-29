import { parse as parseToml } from 'smol-toml'
import { parseAllDocuments } from 'yaml'

export type Parsed = { ok: true; value: unknown } | { ok: false; error: string }

/** The structured format a path is in, or null. */
export function formatOf(path: string): 'json' | 'yaml' | 'toml' | null {
  if (/\.(json|jsonc|webmanifest)$/i.test(path)) return 'json'
  if (/\.(ya?ml)$/i.test(path)) return 'yaml'
  if (/\.toml$/i.test(path)) return 'toml'
  return null
}

/** Parses a JSON, YAML or TOML file. A YAML file with several documents is a list of them. */
export function parseStructured(path: string, text: string): Parsed {
  try {
    switch (formatOf(path)) {
      case 'json':
        return { ok: true, value: JSON.parse(text) }
      case 'toml':
        return { ok: true, value: parseToml(text) }
      case 'yaml': {
        const docs = parseAllDocuments(text)
        const list = Array.isArray(docs) ? docs : [docs]
        const bad = list.find((d) => d.errors.length)
        if (bad) return { ok: false, error: bad.errors[0].message }
        const values = list.map((d) => d.toJS())
        return { ok: true, value: values.length === 1 ? values[0] : values }
      }
    }
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : String(e) }
  }
  return { ok: false, error: 'Not a structured data file.' }
}

/** One line of the tree: a key and either a scalar or the size of a container. */
export interface Node {
  /** Unique path, e.g. "$.dependencies.yaml" or "$.items[2]". */
  id: string
  parent: string | null
  depth: number
  key: string
  container: boolean
  /** "{3}" or "[12]" for containers, else the scalar as text. */
  summary: string
  kind: 'object' | 'array' | 'string' | 'number' | 'boolean' | 'null' | 'other'
}

function kindOf(v: unknown): Node['kind'] {
  if (v === null || v === undefined) return 'null'
  if (Array.isArray(v)) return 'array'
  if (v instanceof Date) return 'string'
  const t = typeof v
  if (t === 'object') return 'object'
  if (t === 'string' || t === 'number' || t === 'boolean') return t
  if (t === 'bigint') return 'number'
  return 'other'
}

function scalar(v: unknown): string {
  if (v === null || v === undefined) return 'null'
  if (v instanceof Date) return v.toISOString()
  if (typeof v === 'string') return JSON.stringify(v)
  return String(v)
}

/** Flattens a value into tree lines in document order, up to max lines. */
export function flatten(value: unknown, max = 50_000): Node[] {
  const out: Node[] = []
  const walk = (v: unknown, id: string, parent: string | null, depth: number, key: string) => {
    if (out.length >= max) return
    const kind = kindOf(v)
    if (kind === 'object' || kind === 'array') {
      const entries: [string, unknown][] =
        kind === 'array' ? (v as unknown[]).map((x, i) => [String(i), x]) : Object.entries(v as Record<string, unknown>)
      out.push({ id, parent, depth, key, container: true, summary: kind === 'array' ? `[${entries.length}]` : `{${entries.length}}`, kind })
      for (const [k, child] of entries) walk(child, kind === 'array' ? `${id}[${k}]` : `${id}.${k}`, id, depth + 1, k)
    } else {
      out.push({ id, parent, depth, key, container: false, summary: scalar(v), kind })
    }
  }
  walk(value, '$', null, 0, '')
  return out
}

/**
 * The lines to show: with a query, the lines whose key or value contains it plus their
 * ancestors (collapsed ones open); without one, everything not inside a collapsed node.
 */
export function visibleNodes(nodes: Node[], collapsed: ReadonlySet<string>, query: string): Node[] {
  const q = query.trim().toLowerCase()
  if (q) {
    const byId = new Map(nodes.map((n) => [n.id, n]))
    const keep = new Set<string>()
    for (const n of nodes) {
      if (n.key.toLowerCase().includes(q) || (!n.container && n.summary.toLowerCase().includes(q))) {
        for (let p: Node | undefined = n; p && !keep.has(p.id); p = p.parent ? byId.get(p.parent) : undefined) keep.add(p.id)
      }
    }
    return nodes.filter((n) => keep.has(n.id))
  }
  const hidden = new Set<string>()
  return nodes.filter((n) => {
    if (n.parent && (hidden.has(n.parent) || collapsed.has(n.parent))) {
      hidden.add(n.id)
      return false
    }
    return true
  })
}
