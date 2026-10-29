const NUM = String.raw`-?(?:\d+\.?\d*|\.\d+)`
const LEADING_MOVE = new RegExp(String.raw`^\s*m\s*(${NUM})[\s,]*(${NUM})[\s,]*`)

/**
 * Joins an icon's parts into one path. Each part is written as if it were its own path,
 * where a leading relative "m" means an absolute position; in a joined path it would be
 * relative to the previous part's end. So the move is made absolute, and any coordinate
 * pairs after it, which are relative line-tos, get an explicit "l" to stay relative.
 */
export function joinParts(parts: readonly string[]): string {
  return parts
    .map((p) =>
      p.replace(LEADING_MOVE, (whole, x: string, y: string, offset: number, all: string) => {
        const next = all.slice(offset + whole.length).trimStart()
        return `M${x} ${y}` + (/^[-.\d]/.test(next) ? 'l' : '')
      }),
    )
    .join(' ')
}
