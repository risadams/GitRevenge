import { LanguageSupport, StreamLanguage, type StringStream } from '@codemirror/language'

// Highlighting for git config files (#226): [section "subsection"] headers, key = value lines,
// booleans and numbers in values, and # or ; comments. The INI mode only knows headers.

interface State {
  /** Inside a [section] header that continues past the current token. */
  header: boolean
  /** Past the key on a key = value line. */
  value: boolean
}

const BOOL = /^(?:true|false|yes|no|on|off)\b/i
const NUMBER = /^-?\d+[kmg]?\b/i

function token(stream: StringStream, state: State): string | null {
  if (stream.sol()) {
    state.header = false
    state.value = false
  }
  if (stream.eatSpace()) return null
  const ch = stream.peek()
  if (ch === '#' || ch === ';') {
    stream.skipToEnd()
    return 'comment'
  }
  if (state.header) {
    if (ch === '"') {
      stream.next()
      while (!stream.eol()) {
        const c = stream.next()
        if (c === '\\') stream.next()
        else if (c === '"') break
      }
      return 'string'
    }
    if (ch === ']') {
      stream.next()
      state.header = false
      return 'bracket'
    }
    stream.eatWhile(/[^\s"\]]/)
    return 'heading'
  }
  if (!state.value) {
    if (ch === '[') {
      stream.next()
      state.header = true
      return 'bracket'
    }
    if (stream.match(/^[A-Za-z][\w-]*/)) {
      state.value = true
      return 'propertyName'
    }
  }
  if (ch === '=') {
    stream.next()
    state.value = true
    return 'operator'
  }
  if (ch === '"') {
    stream.next()
    while (!stream.eol()) {
      const c = stream.next()
      if (c === '\\') stream.next()
      else if (c === '"') break
    }
    return 'string'
  }
  if (stream.match(BOOL) || stream.match(NUMBER)) return stream.current().match(BOOL) ? 'bool' : 'number'
  // A plain value runs to the next space, quote or comment.
  stream.eatWhile(/[^\s"#;]/)
  if (stream.current() === '') stream.next()
  return null
}

export const gitConfigLanguage = StreamLanguage.define<State>({
  name: 'gitconfig',
  startState: () => ({ header: false, value: false }),
  copyState: (s) => ({ ...s }),
  token,
  languageData: { commentTokens: { line: '#' } },
})

/** The language as an editor extension. */
export function gitConfigSyntax(): LanguageSupport {
  return new LanguageSupport(gitConfigLanguage)
}
