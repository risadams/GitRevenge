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

/**
 * Draws an <svg> element onto a canvas at twice its size and returns the PNG as base64.
 * The SVG must carry its colours as values, not CSS variables, since it's drawn standalone.
 */
export async function svgToPngBase64(svg: SVGSVGElement, background: string): Promise<string> {
  const { width, height } = svg.getBoundingClientRect()
  const clone = svg.cloneNode(true) as SVGSVGElement
  clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg')
  clone.setAttribute('width', String(width))
  clone.setAttribute('height', String(height))
  const url = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(new XMLSerializer().serializeToString(clone))}`
  const img = new Image()
  await new Promise<void>((resolve, reject) => {
    img.onload = () => resolve()
    img.onerror = () => reject(new Error('Could not draw the chart'))
    img.src = url
  })
  const canvas = document.createElement('canvas')
  canvas.width = Math.round(width * 2)
  canvas.height = Math.round(height * 2)
  const ctx = canvas.getContext('2d')!
  ctx.fillStyle = background
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  ctx.scale(2, 2)
  ctx.drawImage(img, 0, 0, width, height)
  return canvas.toDataURL('image/png').split(',')[1]
}

/** Formats a duration in hours for people: "40 min", "5.5 h", "3.2 days". */
export function formatHours(h: number): string {
  if (h < 1) return `${Math.round(h * 60)} min`
  if (h < 48) return `${h.toFixed(h < 10 ? 1 : 0)} h`
  return `${(h / 24).toFixed(1)} days`
}
