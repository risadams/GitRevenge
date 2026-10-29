<script lang="ts">
  // A PDF drawn page by page with pdf.js, onto a canvas. pdf.js never runs a document's
  // JavaScript when rendering pages (and since v6 never uses eval); XFA forms are off, and no
  // annotation layer is drawn, so links and form fields in the PDF are inert. pdf.js loads on
  // first use.
  import workerSrc from 'pdfjs-dist/build/pdf.worker.min.mjs?url'
  import type { PDFDocumentLoadingTask, PDFDocumentProxy, RenderTask } from 'pdfjs-dist'
  import Icon from '../../ui/Icon.svelte'
  import type { PreviewProps } from './registry'

  let { loadData }: PreviewProps = $props()

  let doc = $state<PDFDocumentProxy | null>(null)
  let error = $state('')
  let page = $state(1)
  /** "fit" scales to the view's width; a number is a zoom factor. */
  let zoom = $state<'fit' | number>('fit')
  let canvas = $state<HTMLCanvasElement>()
  let width = $state(800)

  $effect(() => {
    let cancelled = false
    let loading: PDFDocumentLoadingTask | undefined
    doc = null
    error = ''
    page = 1
    void (async () => {
      try {
        const b64 = loadData ? await loadData() : ''
        if (!b64) throw new Error('The file is empty or missing.')
        const bytes = Uint8Array.from(atob(b64), (c) => c.charCodeAt(0))
        const pdfjs = await import('pdfjs-dist')
        pdfjs.GlobalWorkerOptions.workerSrc = workerSrc
        loading = pdfjs.getDocument({ data: bytes, enableXfa: false })
        const loaded = await loading.promise
        if (!cancelled) doc = loaded
      } catch (e) {
        if (!cancelled) error = e instanceof Error ? e.message : String(e)
      }
    })()
    return () => {
      cancelled = true
      void loading?.destroy()
    }
  })

  // Draw the current page whenever the page, zoom or width changes.
  $effect(() => {
    const d = doc
    const n = page
    const z = zoom
    const w = width
    const c = canvas
    if (!d || !c) return
    let task: RenderTask | undefined
    let stale = false
    void d.getPage(n).then((p) => {
      if (stale) return
      const base = p.getViewport({ scale: 1 })
      const scale = z === 'fit' ? Math.max(0.25, (w - 32) / base.width) : z
      const viewport = p.getViewport({ scale })
      const dpr = window.devicePixelRatio || 1
      c.width = Math.floor(viewport.width * dpr)
      c.height = Math.floor(viewport.height * dpr)
      c.style.width = `${Math.floor(viewport.width)}px`
      c.style.height = `${Math.floor(viewport.height)}px`
      task = p.render({ canvas: c, viewport, transform: dpr === 1 ? undefined : [dpr, 0, 0, dpr, 0, 0] })
      task.promise.catch(() => {}) // cancelled by a newer render
    })
    return () => {
      stale = true
      task?.cancel()
    }
  })

  const ZOOMS = [0.5, 0.75, 1, 1.25, 1.5, 2, 3, 4]
  function step(dir: 1 | -1) {
    const cur = zoom === 'fit' ? 1 : zoom
    const next = dir === 1 ? ZOOMS.find((z) => z > cur + 0.01) : ZOOMS.findLast((z) => z < cur - 0.01)
    if (next) zoom = next
  }
</script>

<div class="pdf" bind:clientWidth={width}>
  {#if error}
    <p class="error" role="alert">Couldn't open this PDF: {error}</p>
  {:else if !doc}
    <p class="muted">Opening…</p>
  {:else}
    <div class="bar" role="toolbar" aria-label="PDF pages and zoom">
      <button class="tool" disabled={page <= 1} onclick={() => page--} aria-label="Previous page"><Icon name="back" size={14} /></button>
      <label class="pages">
        Page <input type="number" min="1" max={doc.numPages} bind:value={page} aria-label="Page number" /> of {doc.numPages}
      </label>
      <button class="tool" disabled={page >= doc.numPages} onclick={() => page++} aria-label="Next page"><Icon name="right" size={14} /></button>
      <span class="sep" aria-hidden="true"></span>
      <button class="tool" onclick={() => step(-1)} aria-label="Zoom out"><Icon name="minus" size={14} /></button>
      <span class="meta">{zoom === 'fit' ? 'Fit width' : `${Math.round(zoom * 100)}%`}</span>
      <button class="tool" onclick={() => step(1)} aria-label="Zoom in"><Icon name="plus" size={14} /></button>
      <button class="tool" aria-pressed={zoom === 'fit'} onclick={() => (zoom = 'fit')}>Fit</button>
    </div>
    <div class="page"><canvas bind:this={canvas} aria-label="Page {page} of {doc.numPages}"></canvas></div>
  {/if}
</div>

<style>
  .pdf {
    padding: 12px 16px 24px;
  }

  .bar {
    position: sticky;
    top: 0;
    z-index: 1;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px 6px;
    padding: 4px 0 8px;
    background: var(--paper);
  }

  .pages {
    font-size: 13px;
  }

  .pages input {
    width: 56px;
  }

  .sep {
    width: 1px;
    height: 18px;
    margin: 0 6px;
    background: var(--paper-edge);
  }

  .page {
    overflow: auto;
  }

  canvas {
    display: block;
    margin: 0 auto;
    background: #fff;
    box-shadow: var(--shadow);
  }

  .error {
    color: var(--danger);
  }

  .muted {
    color: var(--ink-mute);
    font-style: italic;
  }
</style>
