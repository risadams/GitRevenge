<script lang="ts">
  /** The Git Revenge mark: a flag on a mast with two branches merging into one commit, cut
   * out of the cloth. Theme colours, and a real cut-out (a mask) so it sits on any ground.
   * Below 24 px it switches to the heavier small drawing from build/appicon-small.svg. */
  let { size = 20, label }: { size?: number; label?: string } = $props()

  const id = $props.id()
  let small = $derived(size < 24)
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 64 64"
  role={label ? 'img' : undefined}
  aria-label={label}
  aria-hidden={label ? undefined : 'true'}
  focusable="false"
>
  <defs>
    <mask id="{id}-cut" maskUnits="userSpaceOnUse" x="0" y="0" width="64" height="64">
      <rect width="64" height="64" fill="#fff" />
      {#if small}
        <path d="M24 19 C32 19 34 25.5 45 25.5 M24 32 C32 32 34 25.5 45 25.5" fill="none" stroke="#000" stroke-width="5" stroke-linecap="round" />
        <circle cx="46" cy="25.5" r="5.5" fill="#000" />
      {:else}
        <path d="M24 19 C32 19 34 25 44 25 M24 31 C32 31 34 25 44 25" fill="none" stroke="#000" stroke-width="3.5" stroke-linecap="round" />
        <circle cx="45" cy="25" r="4.5" fill="#000" />
      {/if}
    </mask>
  </defs>
  {#if small}
    <path class="mast" d="M11 8 H18 V58 H11 Z" />
    <path class="flag" d="M18 11 C28 6 39 16 56 10 V39 C39 45 28 35 18 41 Z" mask="url(#{id}-cut)" />
  {:else}
    <path class="mast" d="M13 9 H18 V58 H13 Z" />
    <circle class="mast" cx="15.5" cy="7.5" r="4" />
    <path class="flag" d="M18 12 C28 7 38 17 54 11 V37 C38 43 28 33 18 38 Z" mask="url(#{id}-cut)" />
  {/if}
</svg>

<style>
  svg {
    flex: none;
  }

  .mast {
    fill: var(--accent-a);
  }

  .flag {
    fill: var(--accent);
  }
</style>
