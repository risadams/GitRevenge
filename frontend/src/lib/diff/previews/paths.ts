/**
 * Resolves a link or image in a document against the document's own path in the
 * repository: "img/a.png" from "docs/guide.md" is "docs/img/a.png", and a leading "/" means
 * the repository root. Returns null for a web link, a bare fragment, or a path that would
 * leave the repository.
 */
export function resolveInRepo(docPath: string, href: string): string | null {
  if (/^[a-z][a-z0-9+.-]*:/i.test(href) || href.startsWith('//') || href.startsWith('#')) return null
  let target = href.split(/[?#]/)[0]
  try {
    target = decodeURIComponent(target)
  } catch {
    // Keep it as written.
  }
  if (!target) return null
  const parts = target.startsWith('/') ? [] : docPath.split('/').slice(0, -1)
  for (const seg of target.split('/')) {
    if (seg === '' || seg === '.') continue
    if (seg === '..') {
      if (parts.length === 0) return null
      parts.pop()
    } else parts.push(seg)
  }
  return parts.length ? parts.join('/') : null
}
