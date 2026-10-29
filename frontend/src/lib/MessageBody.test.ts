import { render } from 'svelte/server'
import { describe, expect, it, vi } from 'vitest'

// The Wails runtime only exists in the app; links don't need it to render.
vi.mock('../../wailsjs/runtime/runtime.js', () => ({ BrowserOpenURL: vi.fn() }))

const { default: MessageBody } = await import('./MessageBody.svelte')

/** Rendered HTML without hydration comments or scoped class names. */
const html = (text: string, issueURL = '') =>
  render(MessageBody, { props: { text, issueURL } })
    .body.replace(/<!--.*?-->/g, '')
    .replace(/ class="[^"]*"/g, '')

describe('MessageBody', () => {
  it('renders emphasis, code, lists and links', () => {
    const out = html('Some **bold**, *em* and `x < y`.\n\n- one\n- two\n\nSee [docs](https://example.com).')
    expect(out).toContain('<strong>bold</strong>')
    expect(out).toContain('<em>em</em>')
    expect(out).toContain('<code>x &lt; y</code>')
    expect(out).toMatch(/<ul[^>]*>.*<li>.*one.*<\/li>.*<li>.*two.*<\/li>.*<\/ul>/s)
    expect(out).toContain('href="https://example.com"')
  })

  it('shows raw HTML as text, never as markup', () => {
    const out = html('<script>alert(1)</script>\n\nand <b>inline</b> too')
    expect(out).not.toContain('<script>')
    expect(out).not.toContain('<b>')
    expect(out).toContain('&lt;script>alert(1)&lt;/script>')
  })

  it('links issue references when the remote is known', () => {
    expect(html('Fixes the bug in #12.', 'https://github.com/o/r/issues/{n}')).toContain('href="https://github.com/o/r/issues/12"')
    expect(html('Fixes the bug in #12.')).not.toContain('href=')
  })

  it('keeps an indented block as code', () => {
    expect(html('Trace:\n\n    at foo (bar.go:12)')).toMatch(/<pre><code>at foo \(bar.go:12\)<\/code><\/pre>/)
  })
})
