/** A config file the Git Config tab can edit. */
export interface Scope {
  kind: 'global' | 'local' | 'file'
  path: string
  /** The repository whose view of config the file is shown in, or '' for global. */
  repo: string
}

/** Names a scope for the switcher: "Global", "Local · repo", or "File · ~/path". */
export function scopeLabel(s: Scope, tabs: { path: string; name: string }[]): string {
  switch (s.kind) {
    case 'global':
      return 'Global'
    case 'local':
      return `Local · ${tabs.find((t) => t.path === s.repo)?.name ?? baseName(s.repo)}`
    case 'file':
      return `File · ${baseName(s.path)}`
  }
}

export function baseName(p: string): string {
  return p.split(/[\\/]/).filter(Boolean).pop() ?? p
}
