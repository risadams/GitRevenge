/**
 * Parses delimited text (RFC 4180): fields may be quoted, quotes inside are doubled, and a
 * quoted field may span lines. Stops after maxRows rows; truncated says whether it did.
 */
export function parseDelimited(text: string, delimiter: string, maxRows = Infinity): { rows: string[][]; truncated: boolean } {
  const rows: string[][] = []
  let row: string[] = []
  let field = ''
  let quoted = false
  let i = 0
  const n = text.length
  const endRow = () => {
    row.push(field)
    field = ''
    // A trailing empty line isn't a row.
    if (!(row.length === 1 && row[0] === '')) rows.push(row)
    row = []
  }
  while (i < n) {
    const c = text[i]
    if (quoted) {
      if (c === '"') {
        if (text[i + 1] === '"') {
          field += '"'
          i += 2
          continue
        }
        quoted = false
      } else field += c
      i++
      continue
    }
    if (c === '"' && field === '') quoted = true
    else if (c === delimiter) {
      row.push(field)
      field = ''
    } else if (c === '\n' || c === '\r') {
      if (c === '\r' && text[i + 1] === '\n') i++
      endRow()
      if (rows.length >= maxRows) return { rows, truncated: i + 1 < n }
    } else field += c
    i++
  }
  if (field !== '' || row.length) endRow()
  return { rows, truncated: false }
}

/** Compares two cells: numerically when both are numbers, else as text, ignoring case. */
export function compareCells(a: string, b: string): number {
  const x = Number(a.trim())
  const y = Number(b.trim())
  if (a.trim() !== '' && b.trim() !== '' && !Number.isNaN(x) && !Number.isNaN(y)) return x - y
  return a.localeCompare(b, undefined, { sensitivity: 'base', numeric: true })
}
