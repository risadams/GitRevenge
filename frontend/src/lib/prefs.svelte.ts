import { Preferences } from '../../wailsjs/go/app/App.js'

/** Settings that act outside the settings panel: the interface zoom and the clone folder. */
class Prefs {
  zoom = $state(100)
  cloneDir = $state('')

  /** Loads the saved settings and applies the zoom. Until then, and if they can't be read,
   * the defaults stand. */
  async load() {
    try {
      const p = await Preferences()
      this.zoom = p.zoom
      this.cloneDir = p.cloneDir
    } catch {
      // Keep the defaults.
    }
    this.apply()
  }

  /** Scales the whole interface. The UI is sized in pixels, so a root font size wouldn't. */
  apply() {
    document.documentElement.style.zoom = this.zoom === 100 ? '' : `${this.zoom}%`
  }
}

export const prefs = new Prefs()
