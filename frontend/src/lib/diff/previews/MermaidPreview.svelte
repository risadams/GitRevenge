<script lang="ts">
  // A .mmd or .mermaid file, drawn as its diagram.
  import MermaidDiagram from '../../markdown/MermaidDiagram.svelte'
  import type { PreviewProps } from './registry'

  let { text }: PreviewProps = $props()
</script>

<div class="preview"><MermaidDiagram code={text} /></div>

<style>
  .preview {
    padding: 20px 28px 32px;
  }
</style>
