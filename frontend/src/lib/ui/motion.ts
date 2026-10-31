import type { TransitionConfig } from 'svelte/transition'

/** The motion tokens from style.css, in milliseconds, for transitions driven from script.
 * Keep them in step with --dur-quick, --dur-base, --dur-move and --dur-enter. */
export const DUR = { quick: 120, base: 200, move: 280, enter: 420 } as const

/** How long a list's cascade waits between rows, and how many rows it staggers. */
export const STAGGER_STEP = 40
export const STAGGER_MAX = 6

/** A CSS cubic-bezier(x1, y1, x2, y2) as an easing function on [0, 1]. */
export function cubicBezier(x1: number, y1: number, x2: number, y2: number): (t: number) => number {
  const cx = 3 * x1
  const bx = 3 * (x2 - x1) - cx
  const ax = 1 - cx - bx
  const cy = 3 * y1
  const by = 3 * (y2 - y1) - cy
  const ay = 1 - cy - by
  const sampleX = (s: number) => ((ax * s + bx) * s + cx) * s
  const sampleY = (s: number) => ((ay * s + by) * s + cy) * s
  const slopeX = (s: number) => (3 * ax * s + 2 * bx) * s + cx

  // Solve x(s) = t for s: Newton's method, falling back to bisection where the slope is flat.
  const solve = (t: number) => {
    let s = t
    for (let i = 0; i < 8; i++) {
      const err = sampleX(s) - t
      if (Math.abs(err) < 1e-6) return s
      const d = slopeX(s)
      if (Math.abs(d) < 1e-6) break
      s -= err / d
    }
    let lo = 0
    let hi = 1
    s = t
    while (hi - lo > 1e-6) {
      if (sampleX(s) < t) lo = s
      else hi = s
      s = (lo + hi) / 2
    }
    return s
  }

  return (t) => (t <= 0 ? 0 : t >= 1 ? 1 : sampleY(solve(t)))
}

/** --ease-ink: things arriving and moving. */
export const easeInk = cubicBezier(0.2, 0.6, 0.2, 1)
/** --ease-exit: things leaving. */
export const easeExit = cubicBezier(0.4, 0, 0.6, 1)

/** True when the person asked their system for less motion. */
export function reducedMotion(): boolean {
  return typeof window !== 'undefined' && !!window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
}

/** A duration that collapses to 0 under reduced motion. */
function dur(ms: number): number {
  return reducedMotion() ? 0 : ms
}

/** The delay for row `index` of a list that cascades in: 40ms apart, capped at six rows so
 * a long list never keeps someone waiting. */
export function stagger(index: number): number {
  return Math.min(Math.max(index, 0), STAGGER_MAX - 1) * STAGGER_STEP
}

type Opts = { delay?: number }

/** Dialogs, toasts and the palette arriving: fade, 6px blur to sharp, 6px rise. */
export function inkIn(_node: Element, { delay = 0 }: Opts = {}): TransitionConfig {
  return {
    delay: reducedMotion() ? 0 : delay,
    duration: dur(DUR.enter),
    easing: easeInk,
    css: (t, u) => `opacity: ${t}; filter: blur(${u * 6}px); transform: translateY(${u * 6}px)`,
  }
}

/** Anything leaving: a fade and a 4px drop, faster than it came. */
export function inkOut(): TransitionConfig {
  return {
    duration: dur(DUR.base),
    easing: easeExit,
    css: (t, u) => `opacity: ${t}; transform: translateY(${u * 4}px)`,
  }
}

/** Panels and list rows arriving: a fade and a 4px rise. Pass `delay: stagger(i)` to cascade. */
export function fadeUp(_node: Element, { delay = 0 }: Opts = {}): TransitionConfig {
  return {
    delay: reducedMotion() ? 0 : delay,
    duration: dur(DUR.move),
    easing: easeInk,
    css: (t, u) => `opacity: ${t}; transform: translateY(${u * 4}px)`,
  }
}

/** A row that moved here, such as a file just staged: it slides in 6px under a rose wash
 * that fades over the rest of the time, so the eye follows it. */
export function rowLand(): TransitionConfig {
  const total = DUR.enter * 2
  return {
    duration: dur(total),
    css: (t) => {
      const arrive = easeInk(Math.min(t / 0.4, 1))
      const wash = t < 0.4 ? 1 : 1 - easeInk((t - 0.4) / 0.6)
      return (
        `opacity: ${arrive}; transform: translateX(${(1 - arrive) * -6}px); ` +
        `background-color: color-mix(in srgb, var(--selected) ${Math.round(wash * 100)}%, transparent)`
      )
    },
  }
}
