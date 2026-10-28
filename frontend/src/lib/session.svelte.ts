import {
  ChooseAndOpenRepo,
  CloseRepo,
  OpenRepo,
  RecentRepos,
  RemoveRecentRepo,
  RepoInfo,
  RestoreSession,
  SetActiveRepo,
} from '../../wailsjs/go/app/App.js'
import type { app } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime.js'
import { errorText } from './errors'

/** Payload of the "repo:changed" event (RepoChange in Go). Events aren't bindings, so
 * Wails doesn't generate this type. */
interface RepoChange {
  path: string
  kinds: Array<'head' | 'refs' | 'index' | 'state' | 'config' | 'worktree'>
}

/** Open tabs, the active tab and recent repos, mirrored from the Go session manager. */
class Session {
  tabs = $state<app.RepoTab[]>([])
  /** Path of the active tab, or '' for the start page. */
  active = $state('')
  recent = $state<app.RecentRepo[]>([])
  /** Last error to show in the banner, or ''. */
  error = $state('')
  /** Per-repo counter bumped on every change on disk; views re-fetch when it moves. */
  revision = $state<Record<string, number>>({})

  activeTab = $derived(this.tabs.find((t) => t.path === this.active))

  async restore() {
    EventsOn('repo:changed', (c: RepoChange) => void this.onChange(c))
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

  /** Handles a change detected on disk by the backend watcher. */
  private async onChange(c: RepoChange) {
    this.revision[c.path] = (this.revision[c.path] ?? 0) + 1
    if (c.kinds.includes('head') || c.kinds.includes('refs')) {
      const i = this.tabs.findIndex((t) => t.path === c.path)
      if (i < 0) return
      const tab = await this.run(() => RepoInfo(c.path))
      if (tab) this.tabs[i] = tab
    }
  }

  private add(tab: app.RepoTab) {
    if (!this.tabs.some((t) => t.path === tab.path)) this.tabs.push(tab)
    this.active = tab.path
    this.error = ''
    void this.refreshRecent()
  }

  /** Runs a backend call, turning a rejection into the error banner. */
  async run<T>(fn: () => Promise<T>): Promise<T | undefined> {
    try {
      return await fn()
    } catch (e) {
      this.error = errorText(e)
      return undefined
    }
  }
}

export const session = new Session()
