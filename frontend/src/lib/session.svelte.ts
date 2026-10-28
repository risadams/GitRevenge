import {
  ChooseAndOpenRepo,
  CloseRepo,
  OpenRepo,
  RecentRepos,
  RemoveRecentRepo,
  RestoreSession,
  SetActiveRepo,
} from '../../wailsjs/go/app/App.js'
import type { app } from '../../wailsjs/go/models'

/** Open tabs, the active tab and recent repos, mirrored from the Go session manager. */
class Session {
  tabs = $state<app.RepoTab[]>([])
  /** Path of the active tab, or '' for the start page. */
  active = $state('')
  recent = $state<app.RecentRepo[]>([])
  /** Last error to show in the banner, or ''. */
  error = $state('')

  activeTab = $derived(this.tabs.find((t) => t.path === this.active))

  async restore() {
    const st = await this.run(() => RestoreSession())
    if (!st) return
    this.tabs = st.tabs
    this.active = st.active
    if (st.failed.length > 0) {
      this.error = `Couldn't reopen: ${st.failed.map((f) => f.path).join(', ')}`
    }
    await this.refreshRecent()
  }

  async choose() {
    const tab = await this.run(() => ChooseAndOpenRepo())
    if (tab) this.add(tab)
  }

  async open(path: string) {
    const tab = await this.run(() => OpenRepo(path))
    if (tab) this.add(tab)
  }

  async close(path: string) {
    const i = this.tabs.findIndex((t) => t.path === path)
    if (i < 0) return
    await this.run(() => CloseRepo(path))
    this.tabs = this.tabs.filter((t) => t.path !== path)
    if (this.active === path) {
      // Same rule as the backend: select the tab that slid into this position.
      this.active = this.tabs[Math.min(i, this.tabs.length - 1)]?.path ?? ''
    }
  }

  async select(path: string) {
    this.active = path
    await this.run(() => SetActiveRepo(path))
  }

  async forget(path: string) {
    await this.run(() => RemoveRecentRepo(path))
    await this.refreshRecent()
  }

  async refreshRecent() {
    this.recent = (await this.run(() => RecentRepos())) ?? []
  }

  private add(tab: app.RepoTab) {
    if (!this.tabs.some((t) => t.path === tab.path)) this.tabs.push(tab)
    this.active = tab.path
    this.error = ''
    void this.refreshRecent()
  }

  /** Runs a backend call, turning a rejection into the error banner. */
  private async run<T>(fn: () => Promise<T>): Promise<T | undefined> {
    try {
      return await fn()
    } catch (e) {
      this.error = String(e)
      return undefined
    }
  }
}

export const session = new Session()
