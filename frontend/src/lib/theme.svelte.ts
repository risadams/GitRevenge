import { SetTheme, Theme } from '../../wailsjs/go/app/App.js'

type Resolved = 'light' | 'dark'
/** A saved choice, or '' to follow the operating system. */
type Mode = '' | Resolved

const media = window.matchMedia('(prefers-color-scheme: dark)')

/** The colour theme, applied as data-theme on the document root. */
class ThemeState {
  mode = $state<Mode>('')
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

  /** Loads the saved choice. Until it arrives, and if it can't be read, the OS decides. */
  async load() {
    try {
      this.mode = (await Theme()) as Mode
    } catch {
      // Keep following the operating system.
    }
  }

  /** Switches to the other theme and saves that as the choice. */
  async toggle() {
    this.mode = this.resolved === 'dark' ? 'light' : 'dark'
    await SetTheme(this.mode)
  }
}

export const theme = new ThemeState()
