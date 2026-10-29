<script lang="ts">
  import type { git } from '../../../wailsjs/go/models'

  let { contents }: { contents: git.FileContents } = $props()

  let mode = $state<'side' | 'swipe'>('side')
  let swipe = $state(50)

  const src = (s: git.Side) => `data:${contents.mime};base64,${s.image}`
  const size = (n: number) => (n < 1024 ? `${n} B` : n < 1 << 20 ? `${(n / 1024).toFixed(1)} KB` : `${(n / (1 << 20)).toFixed(1)} MB`)
</script>

<div class="images">
  {#if contents.old.exists && contents.new.exists}
    <div class="modes">
      <span class="segmented" role="group" aria-label="Compare images">
        <button aria-pressed={mode === 'side'} onclick={() => (mode = 'side')}>Side by side</button>
        <button aria-pressed={mode === 'swipe'} onclick={() => (mode = 'swipe')}>Swipe</button>
      </span>
      {#if mode === 'swipe'}<input type="range" min="0" max="100" bind:value={swipe} aria-label="Swipe position" />{/if}
    </div>
  {/if}

  {#if mode === 'swipe' && contents.old.exists && contents.new.exists}
    <div class="swipe">
      <img src={src(contents.new)} alt="New version" />
      <img class="over" src={src(contents.old)} alt="Old version" style:clip-path="inset(0 {100 - swipe}% 0 0)" />
    </div>
  {:else}
    <div class="side">
      {#each [['Before', contents.old], ['After', contents.new]] as [label, s] (label)}
        {@const side = s as git.Side}
        <figure>
          <figcaption class="meta">{label}{side.exists ? ` · ${size(side.size)}` : ''}</figcaption>
          {#if side.exists}<img src={src(side)} alt="{label} version" />{:else}<p class="muted">Not present</p>{/if}
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
    gap: 6px;
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
</style>
