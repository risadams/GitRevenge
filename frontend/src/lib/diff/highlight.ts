import { LanguageDescription, type LanguageSupport } from '@codemirror/language'
import { languages } from '@codemirror/language-data'
import { classHighlighter, highlightCode } from '@lezer/highlight'

/** One styled piece of a line. cls is a `tok-*` class, or '' for plain text. */
export interface Token {
  text: string
  cls: string
}

const cache = new Map<string, Promise<LanguageSupport | null>>()

/** Loads the language for a file name, or null if none matches. Languages load lazily. */
export function languageFor(filename: string): Promise<LanguageSupport | null> {
  const desc = LanguageDescription.matchFilename(languages, filename)
  if (!desc) return Promise.resolve(null)
  let p = cache.get(desc.name)
  if (!p) {
    p = desc.load().catch(() => null)
    cache.set(desc.name, p)
  }
  return p
}

/** Above this size, highlighting is skipped; parsing would stall the UI. */
const MAX_HIGHLIGHT = 1 << 20

/**
 * Highlights text and returns tokens per line. Lines match splitLines(text) by index;
 * terminators are not included in the tokens.
 */
export function highlightLines(text: string, lang: LanguageSupport | null): Token[][] {
  if (!lang || text.length > MAX_HIGHLIGHT) {
    return text.split('\n').map((line) => [{ text: line.replace(/\r$/, ''), cls: '' }])
  }
  const lines: Token[][] = [[]]
  const tree = lang.language.parser.parse(text)
  highlightCode(
    text,
    tree,
    classHighlighter,
    (t, cls) => lines[lines.length - 1].push({ text: t.replace(/\r$/, ''), cls }),
    () => lines.push([]),
  )
  return lines
}
