import { CheckForUpdate, SetUpdateChecks, UpdateChecks } from '../../wailsjs/go/app/App.js'
import type { app } from '../../wailsjs/go/models'

const DISMISSED_KEY = 'gitrevenge.updateDismissed'

function dismissedVersion(): string {
  try {
    return localStorage.getItem(DISMISSED_KEY) ?? ''
  } catch {
    return ''
  }
}

/** The check for a newer release on GitHub, and what the banner shows about it. */
class UpdateState {
  /** A newer release to announce, or null. */
  available = $state<app.UpdateInfo | null>(null)
  /** "You're up to date", after a check the user asked for. */
  upToDate = $state('')
  checksOn = $state(true)

  /** The automatic check at startup. Quiet on failure, and quiet about a version the user
   * already dismissed. */
  async startup() {
    try {
      this.checksOn = await UpdateChecks()
      const info = await CheckForUpdate(true)
      if (info.available && info.latest !== dismissedVersion()) this.available = info
    } catch {
      // Offline or rate limited: try again next launch.
    }
  }

  /** A check the user asked for: always runs, and always answers. */
  async checkNow() {
    this.upToDate = ''
    const info = await CheckForUpdate(false)
    if (info.available) this.available = info
    else this.upToDate = `You're on the newest version, ${info.current}.`
  }

  /** Hides the banner; a dismissed version isn't announced again at startup. */
  dismiss() {
    if (this.available) {
      try {
        localStorage.setItem(DISMISSED_KEY, this.available.latest)
      } catch {
        // Storage unavailable; it's announced again next launch.
      }
    }
    this.available = null
    this.upToDate = ''
  }

  async toggleChecks() {
    await SetUpdateChecks(!this.checksOn)
    this.checksOn = !this.checksOn
  }
}

export const update = new UpdateState()
