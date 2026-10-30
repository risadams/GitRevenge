import { canonicalKey, keyInfo, sectionOf } from './catalog'

/** The fields of a config entry the table needs (git.ConfigEntry). */
export interface Entry {
  scope: string
  file: string
  key: string
  value: string
  implicit: boolean
}

/** A value set in another file that applies here. */
export interface Inherited {
  scope: string
  file: string
  value: string
  implicit: boolean
}

/** One key of the table. */
export interface Row {
  /** The key as written in the edited file, or as git reports it when only inherited. */
  key: string
  section: string
  /** The values in the edited file, in file order. */
  values: Entry[]
  /** Values of the same key from other files, in the order git reads them. */
  inherited: Inherited[]
  /** Git appends values of this key instead of replacing them. */
  multi: boolean
}

export interface Section {
  name: string
  rows: Row[]
}

/** Whether two paths name the same file. Config paths come from git and from the OS in
 * either slash direction, and Windows and macOS ignore case. */
export function samePath(a: string, b: string): boolean {
  const norm = (p: string) => p.replace(/\\/g, '/').replace(/\/+$/, '').toLowerCase()
  return norm(a) === norm(b)
}

/** Builds the table for one file: its own entries (from the working copy) merged with what
 * every other file git reads sets for the same keys. */
export function buildRows(own: Entry[], effective: Entry[], file: string): Section[] {
  const rows = new Map<string, Row>()
  const row = (key: string) => {
    const id = canonicalKey(key)
    let r = rows.get(id)
    if (!r) {
      r = { key, section: sectionOf(key), values: [], inherited: [], multi: false }
      rows.set(id, r)
    }
    return r
  }
  for (const e of own) row(e.key).values.push(e)
  for (const e of effective) {
    // The edited file's entries come from the working copy, not from disk.
    if (samePath(e.file, file)) continue
    row(e.key).inherited.push({ scope: e.scope, file: e.file, value: e.value, implicit: e.implicit })
  }
  const sections = new Map<string, Section>()
  for (const r of rows.values()) {
    // The same key in several files is an override; only one file setting it twice makes a list.
    const perFile = new Map<string, number>()
    for (const i of r.inherited) perFile.set(i.file, (perFile.get(i.file) ?? 0) + 1)
    r.multi = keyInfo(r.key)?.multi === true || r.values.length > 1 || [...perFile.values()].some((n) => n > 1)
    let s = sections.get(r.section)
    if (!s) sections.set(r.section, (s = { name: r.section, rows: [] }))
    s.rows.push(r)
  }
  return [...sections.values()]
}

/** What a row's own values do to the inherited ones, for the note under it. */
export function overrideNote(r: Row): string {
  if (r.values.length === 0 || r.inherited.length === 0) return ''
  if (r.multi) return `adds to ${scopesOf(r.inherited)}`
  const last = r.inherited[r.inherited.length - 1]
  return `overrides ${last.scope}: ${last.implicit ? 'true' : last.value}`
}

function scopesOf(list: Inherited[]): string {
  return [...new Set(list.map((i) => i.scope))].join(', ')
}

/** Rows whose key or any value contains the filter text, ignoring case. */
export function filterSections(sections: Section[], text: string, showInherited: boolean): Section[] {
  const q = text.trim().toLowerCase()
  return sections
    .map((s) => ({
      ...s,
      rows: s.rows.filter((r) => {
        if (!showInherited && r.values.length === 0) return false
        if (!q) return true
        const values = [...r.values.map((v) => v.value), ...(showInherited ? r.inherited.map((i) => i.value) : [])]
        return r.key.toLowerCase().includes(q) || values.some((v) => v.toLowerCase().includes(q))
      }),
    }))
    .filter((s) => s.rows.length > 0)
}

const SECRET_KEY = /\.(extraheader|password|token|secret)$/i

/** Whether a key's values are secrets that are masked until revealed. */
export function isSecretKey(key: string): boolean {
  return SECRET_KEY.test(key)
}

export const MASK = '••••'

/** Masks the credentials in any URL in text: user:password@ becomes user:••••@, and a lone
 * user@ (often a token) becomes ••••@. */
export function maskUrls(text: string): string {
  return text.replace(/\b([a-z][a-z0-9+.-]*:\/\/)([^/@\s:]+)(:[^/@\s]*)?@/gi, (_m, scheme: string, user: string, pass: string | undefined) =>
    pass ? `${scheme}${user}:${MASK}@` : `${scheme}${MASK}@`,
  )
}

/** A value as shown before it is revealed. */
export function maskValue(key: string, value: string): string {
  return isSecretKey(key) && value ? MASK : maskUrls(value)
}

/** Whether masking hides anything in this key or value, so the row needs a reveal control. */
export function hasSecret(key: string, value: string): boolean {
  return maskValue(key, value) !== value || maskUrls(key) !== key
}
