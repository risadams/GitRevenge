/** A graph row as returned by the backend (graph.Row in Go). */
export interface Seg {
  from: number
  to: number
  color: number
}

export interface Row {
  index: number
  col: number
  color: number
  up: Seg[]
  down: Seg[]
  width: number
  id: string
  parents: number
  author: string
  authorTime: number
  subject: string
}

/** A branch, tag or stash (git.Ref in Go). */
export interface Ref {
  name: string
  short: string
  kind: 'local' | 'remote' | 'tag' | 'stash'
  commit: string
  head: boolean
  upstream: string
  ahead: number
  behind: number
  gone: boolean
  subject?: string
}

export interface Page {
  rows: Row[]
  refs: Record<string, Ref[]>
  total: number
  loading: boolean
}

export const CHUNK = 200

/**
 * Caches graph rows in fixed-size chunks fetched on demand. A chunk that came back short
 * (the graph was still loading) is refetched once more rows exist.
 */
export class RowCache {
  private chunks = new Map<number, Row[]>()
  private labels = new Map<string, Ref[]>()
  private inflight = new Map<number, Promise<void>>()
  private generation = 0
  total = 0
  loading = true

  constructor(
    private fetchPage: (start: number, count: number) => Promise<Page>,
    private onUpdate: () => void,
  ) {}

  get(i: number): Row | undefined {
    return this.chunks.get(Math.floor(i / CHUNK))?.[i % CHUNK]
  }

  /** Refs pointing at the commit, as of when its row was fetched. */
  refsFor(id: string): Ref[] {
    return this.labels.get(id) ?? []
  }

  /** Ensures rows [start, end) are loaded or loading. */
  ensure(start: number, end: number) {
    const first = Math.max(0, Math.floor(start / CHUNK))
    const last = Math.floor(Math.max(start, end - 1) / CHUNK)
    for (let c = first; c <= last; c++) {
      const have = this.chunks.get(c)
      const complete = have && (have.length === CHUNK || (!this.loading && c * CHUNK + have.length >= this.total))
      if (!complete && !this.inflight.has(c)) this.load(c)
    }
  }

  /** Records a new total from a progress event; short chunks will be refetched. */
  setTotal(total: number, loading: boolean) {
    this.total = total
    this.loading = loading
  }

  /** Drops everything, e.g. after the backend swapped in a reloaded graph. */
  invalidate() {
    this.chunks.clear()
    this.labels.clear()
    this.inflight.clear()
    this.generation++
  }

  private load(c: number) {
    const gen = this.generation
    const p = this.fetchPage(c * CHUNK, CHUNK)
      .then((page) => {
        if (gen !== this.generation) return
        this.chunks.set(c, page.rows)
        for (const [id, refs] of Object.entries(page.refs ?? {})) this.labels.set(id, refs)
        this.total = page.total
        this.loading = page.loading
        this.onUpdate()
      })
      .catch(() => {
        // Leave the chunk missing; the next ensure() retries.
      })
      .finally(() => {
        if (this.inflight.get(c) === p) this.inflight.delete(c)
      })
    this.inflight.set(c, p)
  }
}
