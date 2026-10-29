import { GraphDensity, SetGraphDensity, SetTheme, Theme } from '../../wailsjs/go/app/App.js'
import type { Density } from './graph/draw'

type Resolved = 'light' | 'dark'
/** A saved choice, or '' to follow the operating system. */
type Mode = '' | Resolved

const media = window.matchMedia('(prefers-color-scheme: dark)')

/** Appearance: the colour theme, applied as data-theme on the document root, and the
 * graph's row density. */
class ThemeState {
  mode = $state<Mode>('')
  density = $state<Density>('comfortable')
  system = $state<Resolved>(media.matches ? 'dark' : 'light')
  resolved: Resolved = $derived(this.mode || this.system)

  constructor() {
    // Set before the first render so the page never paints in the wrong theme.
    document.documentElement.dataset.theme = this.resolved
    media.addEventListener('change', (e) => (this.system = e.matches ? 'dark' : 'light'))
    $effect.root(() => {
      $effect(() => {
        document.documentElement.dataset.theme = this.resolved
      })
    })
  }

  /** Loads the saved choices. Until they arrive, and if they can't be read, the OS picks the
   * theme and the graph uses the timeline. */
  async load() {
    try {
      const [mode, density] = await Promise.all([Theme(), GraphDensity()])
      this.mode = mode as Mode
      this.density = density === 'compact' ? 'compact' : 'comfortable'
    } catch {
      // Keep the defaults.
    }
  }

  /** Switches the graph between the timeline and compact rows, and saves the choice. */
  async toggleDensity() {
    this.density = this.density === 'compact' ? 'comfortable' : 'compact'
    await SetGraphDensity(this.density === 'compact' ? 'compact' : '')
  }

  /** Switches to the other theme and saves that as the choice. */
  async toggle() {
    this.mode = this.resolved === 'dark' ? 'light' : 'dark'
    await SetTheme(this.mode)
  }
}

export const theme = new ThemeState()
