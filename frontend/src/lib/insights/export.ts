/** A table as CSV: fields with commas, quotes or line breaks are quoted, quotes doubled. */
export function toCSV(rows: (string | number)[][]): string {
  const field = (v: string | number) => {
    const s = String(v)
    return /[",\r\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s
  }
  return rows.map((r) => r.map(field).join(',')).join('\r\n') + '\r\n'
}

/** UTF-8 text as base64, for SaveFile. */
export function textToBase64(text: string): string {
  const bytes = new TextEncoder().encode(text)
  let bin = ''
  for (const b of bytes) bin += String.fromCharCode(b)
  return btoa(bin)
}

/** Formats a duration in hours for people: "40 min", "5.5 h", "3.2 days". */
export function formatHours(h: number): string {
  if (h < 1) return `${Math.round(h * 60)} min`
  if (h < 48) return `${h.toFixed(h < 10 ? 1 : 0)} h`
  return `${(h / 24).toFixed(1)} days`
}
