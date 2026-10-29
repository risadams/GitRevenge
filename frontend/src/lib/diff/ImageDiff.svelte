<script lang="ts">
  import type { git } from '../../../wailsjs/go/models'

  let { contents }: { contents: git.FileContents } = $props()

  // One component for comparing and for viewing a single version (the rendered file view).
  let both = $derived(contents.old.exists && contents.new.exists)
  let mode = $state<'side' | 'swipe' | 'file'>('side')
  let zoom = $state<'fit' | 'actual'>('fit')
  let swipe = $state(50)
  /** Pixel size of each loaded image, by its data URL. */
  let dims = $state<Record<string, string>>({})

  // A file only on one side has nothing to compare: show that version.
  $effect(() => {
    if (!both && mode !== 'file') mode = 'file'
  })

  let shown = $derived(contents.new.exists ? contents.new : contents.old)

  function measured(e: Event) {
    const img = e.currentTarget as HTMLImageElement
    dims[img.src] = `${img.naturalWidth} × ${img.naturalHeight}`
  }

  const src = (s: git.Side) => `data:${contents.mime};base64,${s.image}`
  const size = (n: number) => (n < 1024 ? `${n} B` : n < 1 << 20 ? `${(n / 1024).toFixed(1)} KB` : `${(n / (1 << 20)).toFixed(1)} MB`)
</script>

<div class="images" class:actual={zoom === 'actual'}>
  <div class="modes">
    <span class="segmented" role="group" aria-label="Image view">
      {#if both}
        <button aria-pressed={mode === 'side'} onclick={() => (mode = 'side')}>Side by side</button>
        <button aria-pressed={mode === 'swipe'} onclick={() => (mode = 'swipe')}>Swipe</button>
      {/if}
      <button aria-pressed={mode === 'file'} onclick={() => (mode = 'file')}>This version</button>
    </span>
    <span class="segmented" role="group" aria-label="Zoom">
      <button aria-pressed={zoom === 'fit'} onclick={() => (zoom = 'fit')}>Fit</button>
      <button aria-pressed={zoom === 'actual'} onclick={() => (zoom = 'actual')}>Actual size</button>
    </span>
    {#if mode === 'swipe'}<input type="range" min="0" max="100" bind:value={swipe} aria-label="Swipe position" />{/if}
  </div>

  {#if mode === 'file'}
    <figure class="single">
      <figcaption class="meta">
        {contents.new.exists ? 'This version' : 'Deleted · last version'} · {size(shown.size)}{dims[src(shown)] ? ` · ${dims[src(shown)]}` : ''}
      </figcaption>
      <img src={src(shown)} alt={contents.path} onload={measured} />
    </figure>
  {:else if mode === 'swipe' && both}
    <div class="swipe">
      <img src={src(contents.new)} alt="New version" />
      <img class="over" src={src(contents.old)} alt="Old version" style:clip-path="inset(0 {100 - swipe}% 0 0)" />
    </div>
  {:else}
    <div class="side">
      {#each [['Before', contents.old], ['After', contents.new]] as [label, s] (label)}
        {@const side = s as git.Side}
        <figure>
          <figcaption class="meta">{label}{side.exists ? ` · ${size(side.size)}` : ''}{side.exists && dims[src(side)] ? ` · ${dims[src(side)]}` : ''}</figcaption>
          {#if side.exists}<img src={src(side)} alt="{label} version" onload={measured} />{:else}<p class="muted">Not present</p>{/if}
        </figure>
      {/each}
    </div>
  {/if}
</div>

<style>
  .images {
    padding: 16px;
  }

  .modes {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 12px;
    align-items: center;
    margin-bottom: 12px;
  }

  input[type='range'] {
    accent-color: var(--accent);
  }

  .side {
    display: flex;
    gap: 16px;
  }

  figure {
    flex: 1;
    margin: 0;
    min-width: 0;
  }

  figcaption {
    color: var(--ink-mute);
    margin-bottom: 6px;
  }

  img {
    max-width: 100%;
    /* A checkerboard shows transparency. */
    background: repeating-conic-gradient(var(--paper-deep) 0% 25%, var(--paper-warm) 0% 50%) 50% / 16px 16px;
  }

  .swipe {
    position: relative;
    display: inline-block;
  }

  .swipe .over {
    position: absolute;
    inset: 0;
  }

  .muted {
    color: var(--ink-mute);
  }

  .single {
    margin: 0;
  }

  .single img {
    max-width: 100%;
    height: auto;
  }

  /* Actual size: pixels at 1:1, scrolling when larger than the view. */
  .actual img {
    max-width: none !important;
    width: auto;
  }
</style>
