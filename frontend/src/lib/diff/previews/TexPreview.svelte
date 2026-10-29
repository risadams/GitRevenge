<script lang="ts">
  // A .tex file's math. A full LaTeX document isn't typeset here, so its display math is
  // shown on its own; a file that is just math renders whole.
  import MathView from '../../markdown/MathView.svelte'
  import { texMath } from '../../markdown/math'
  import type { PreviewProps } from './registry'

  let { text }: PreviewProps = $props()

  let blocks = $derived(texMath(text))
  let isDocument = $derived(/\\documentclass|\\begin\{document\}/.test(text))
</script>

<div class="tex">
  {#if isDocument}
    <p class="meta">LaTeX document: showing its {blocks.length} display {blocks.length === 1 ? 'equation' : 'equations'}. Raw has the full source.</p>
  {/if}
  {#each blocks as tex, i (i)}<MathView {tex} display />{:else}<p class="muted">No display math found. Raw has the source.</p>{/each}
</div>

<style>
  .tex {
    padding: 20px 28px 32px;
  }

  .muted {
    color: var(--ink-mute);
    font-style: italic;
  }
</style>
