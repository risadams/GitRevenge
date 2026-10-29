<script lang="ts">
  // A markdown file rendered as a document. Relative images load from the repository at the
  // same version; relative links open that file. Remote images aren't fetched.
  import Markdown from '../../Markdown.svelte'
  import type { PreviewProps } from './registry'
  import { resolveInRepo } from './paths'

  let { path, text, loadImage, openFile }: PreviewProps = $props()

  async function resolveImage(src: string): Promise<string> {
    const file = resolveInRepo(path, src)
    return file ? loadImage(file) : ''
  }

  function onLink(href: string) {
    const file = resolveInRepo(path, href)
    if (file) openFile(file)
  }
</script>

<div class="preview"><Markdown {text} document {resolveImage} {onLink} /></div>

<style>
  .preview {
    padding: 24px 32px 40px;
    background: var(--paper);
  }
</style>
