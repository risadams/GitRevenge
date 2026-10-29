const LIST = /^\s{0,3}([-*+•]|\d+[.)])\s/
const CODE = /^(\t| {4})/
/** Quotes, headings and trailers (Co-authored-by: …) stand on their own line. */
const OWN_LINE = /^(>|#|[A-Z][\w-]*: )/

/**
 * Reflows a commit message body for proportional type. Git bodies are usually hard-wrapped
 * near 72 columns, which reads as ragged mid-sentence breaks outside a terminal, so the
 * wrapped lines of a paragraph or list item are joined. Blank lines, new list items,
 * quotes, headings, indented code and trailers keep their line breaks.
 */
export function reflow(body: string): string {
  return body
    .replace(/\r\n/g, '\n')
    .split(/\n{2,}/)
    .map((para) => {
      const joined: string[] = []
      for (const line of para.split('\n')) {
        const prev = joined[joined.length - 1]
        const continues =
          prev !== undefined &&
          prev.trim() !== '' &&
          line.trim() !== '' &&
          !LIST.test(line) &&
          !CODE.test(line) &&
          !OWN_LINE.test(line) &&
          !CODE.test(prev) &&
          !OWN_LINE.test(prev)
        if (continues) joined[joined.length - 1] = `${prev.trimEnd()} ${line.trim()}`
        else joined.push(line)
      }
      return joined.join('\n')
    })
    .join('\n\n')
}
