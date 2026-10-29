<script lang="ts" module>
  import { PATHS, type IconName as Name } from './icons'

  export type IconName = Name

  /** Watercolor tones for a wash behind an icon. */
  export type Wash = 'rose' | 'teal' | 'plum' | 'sepia' | 'ochre'
</script>

<script lang="ts">
  let {
    name,
    size = 16,
    wash,
  }: {
    name: IconName
    size?: number
    /** Draws a watercolor wash behind the icon. Only for headers and empty states, never in dense lists. */
    wash?: Wash
  } = $props()

  // Optical weight: heavier at small sizes, so every size renders between 1.33 and 1.75 px.
  let stroke = $derived(size <= 20 ? 2 : 1.75)
</script>

<svg width={size} height={size} viewBox="0 0 24 24" aria-hidden="true" focusable="false" class:washed={!!wash}>
  {#if wash}
    <path class="wash" style:fill="var(--ink-{wash})" d="M6.2 4.6C9.6 2.4 16.8 3 19.6 6.6c2.6 3.4 1.4 9.6-2.4 12.2-3.6 2.4-9.4 2-12.2-1.2-2.8-3.4-2.6-10.6 1.2-13z" />
  {/if}
  <path d={PATHS[name]} fill="none" stroke="currentColor" stroke-width={stroke} stroke-linecap="round" stroke-linejoin="round" />
</svg>

<style>
  svg {
    flex: none;
  }

  svg.washed {
    overflow: visible;
  }

  .wash {
    opacity: 0.3;
    transform: translate(2px, 1px) rotate(-12deg);
    transform-origin: 12px 12px;
  }

  @media (forced-colors: active) {
    .wash {
      display: none;
    }
  }
</style>
