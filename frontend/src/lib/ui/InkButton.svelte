<script lang="ts">
  import type { Snippet } from 'svelte'

  /** The brand's call to action: a serif label on a torn ink blob. One primary per view;
   * ghost (an ink outline) for the actions beside it. */
  let {
    variant = 'primary',
    arrow = false,
    type = 'button',
    disabled = false,
    title,
    onclick,
    children,
  }: {
    variant?: 'primary' | 'ghost'
    arrow?: boolean
    type?: 'button' | 'submit'
    disabled?: boolean
    title?: string
    onclick?: (e: MouseEvent) => void
    children: Snippet
  } = $props()
</script>

<button class="ink {variant}" {type} {disabled} {title} {onclick}>
  {@render children()}
  {#if arrow}<span class="arrow" aria-hidden="true">→</span>{/if}
</button>

<style>
  .ink {
    position: relative;
    z-index: 0;
    display: inline-flex;
    align-items: center;
    gap: 10px;
    padding: 10px 24px;
    background: none;
    border: none;
    border-radius: 0;
    font-family: var(--font-serif);
    font-size: 15px;
    font-weight: 500;
    letter-spacing: 0.01em;
    cursor: pointer;
    transition: transform 0.3s var(--ease-ink);
  }

  .ink::before {
    content: '';
    position: absolute;
    inset: 0;
    z-index: -1;
    filter: url(#wc-edge);
    border-radius: 45% 55% 50% 50% / 55% 45%;
    transition:
      opacity 0.3s,
      transform 0.3s var(--ease-ink),
      background-color 0.3s;
  }

  .ink:hover:not(:disabled) {
    background: none;
    transform: translateY(-1px);
  }

  .ink:hover:not(:disabled)::before {
    transform: scale(1.03);
  }

  .primary {
    color: var(--on-accent);
  }

  .primary::before {
    background: var(--accent);
    opacity: 0.92;
  }

  .primary:hover:not(:disabled)::before {
    opacity: 1;
  }

  .ghost {
    color: var(--ink);
  }

  .ghost::before {
    border: 1.5px solid var(--ink);
  }

  .ghost:hover:not(:disabled)::before {
    background: var(--wash-hover);
  }

  .arrow {
    font-family: var(--font-hand);
    font-size: 22px;
    line-height: 1;
    transition: transform 0.3s var(--ease-ink);
  }

  .ink:hover:not(:disabled) .arrow {
    transform: translateX(4px);
  }

  @media (prefers-reduced-motion: reduce) {
    .ink,
    .ink::before,
    .arrow {
      transition: none;
    }

    .ink:hover:not(:disabled),
    .ink:hover:not(:disabled)::before,
    .ink:hover:not(:disabled) .arrow {
      transform: none;
    }
  }
</style>
