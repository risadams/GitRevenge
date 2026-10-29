/** A trailer value read as a person: "Name <email>", as in Co-authored-by or Signed-off-by. */
export interface Person {
  name: string
  email: string
}

const PERSON = /^\s*(.+?)\s*<([^<>\s]+@[^<>\s]+)>\s*$/

export function asPerson(value: string): Person | null {
  const m = PERSON.exec(value)
  return m ? { name: m[1], email: m[2] } : null
}

/** A run of a trailer value: plain text, or a link (an issue reference or a URL). */
export interface Part {
  text: string
  href?: string
}

const TOKEN = /(https?:\/\/[^\s<>]+[^\s<>.,;:!?)\]'"])|(#(\d+))/g

/**
 * Splits a trailer value into text and links. URLs link to themselves; "#123" links to the
 * issue when issueURL (a pattern with "{n}") is known, and stays text otherwise.
 */
export function valueParts(value: string, issueURL: string): Part[] {
  const parts: Part[] = []
  let at = 0
  for (const m of value.matchAll(TOKEN)) {
    if (m.index > at) parts.push({ text: value.slice(at, m.index) })
    if (m[1]) parts.push({ text: m[1], href: m[1] })
    else parts.push(issueURL ? { text: m[2], href: issueURL.replace('{n}', m[3]) } : { text: m[2] })
    at = m.index + m[0].length
  }
  if (at < value.length) parts.push({ text: value.slice(at) })
  return parts
}
