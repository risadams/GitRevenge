<!-- The watercolour edge filters from Ink and Agency. Mounted once; referenced in CSS as
     filter: url(#wc-edge) for torn, organic edges on blobs, bars and buttons. -->
<svg class="filters" aria-hidden="true" focusable="false" width="0" height="0">
  <defs>
    <filter id="wc-edge" x="-20%" y="-20%" width="140%" height="140%">
      <feTurbulence type="fractalNoise" baseFrequency="0.012" numOctaves="3" seed="4" result="noise" />
      <feDisplacementMap in="SourceGraphic" in2="noise" scale="14" />
      <feGaussianBlur stdDeviation="0.6" />
    </filter>
    <filter id="wc-bleed" x="-30%" y="-30%" width="160%" height="160%">
      <feTurbulence type="fractalNoise" baseFrequency="0.008" numOctaves="2" seed="7" result="noise" />
      <feDisplacementMap in="SourceGraphic" in2="noise" scale="40" />
    </filter>
    <filter id="paper-grain">
      <feTurbulence type="fractalNoise" baseFrequency="0.85" numOctaves="3" stitchTiles="stitch" />
      <feColorMatrix type="saturate" values="0" />
    </filter>
  </defs>
</svg>

<style>
  .filters {
    position: absolute;
    overflow: hidden;
  }
</style>
