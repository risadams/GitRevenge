<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLButtonAttributes } from 'svelte/elements'

  /** A button from the family in style.css: one shape in three sizes and five variants.
   * Primary is the one call to action per view (a clean ink body with a small bleed behind
   * it); secondary sits beside it; quiet is for toolbars and rows; text is an inline italic
   * link; danger confirms something destructive. Size alone sets height, padding, type and
   * radius: md by default, sm in dense rows, lg for a view's main action. The arrow belongs
   * on an lg primary that leads somewhere. Busy keeps the label and width. */
  let {
    variant = 'primary',
    size = 'md',
    arrow = false,
    icon = false,
    busy = false,
    type = 'button',
    disabled = false,
    onclick,
    children,
    ...rest
  }: Omit<HTMLButtonAttributes, 'children'> & {
    variant?: 'primary' | 'secondary' | 'quiet' | 'text' | 'danger' | 'ghost'
    size?: 'sm' | 'md' | 'lg'
    arrow?: boolean
    /** Icon-only: a square button. Give it an aria-label. */
    icon?: boolean
    busy?: boolean
    onclick?: (e: MouseEvent) => void
    children: Snippet
  } = $props()

  // ghost was the 0.7 name for the outline button.
  const kind = $derived(variant === 'ghost' ? 'secondary' : variant)
</script>

<button
  {...rest}
  class={['btn', kind, size !== 'md' && size, icon && 'icon']}
  {type}
  {disabled}
  aria-busy={busy || undefined}
  onclick={(e) => {
    if (!busy) onclick?.(e)
  }}
>
  {#if busy}<span class="busy" aria-hidden="true"><i></i><i></i><i></i></span>{/if}
  {@render children()}
  {#if arrow}<span class="arrow" aria-hidden="true">→</span>{/if}
</button>
