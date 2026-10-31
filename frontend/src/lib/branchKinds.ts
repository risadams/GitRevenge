import type { IconName } from './ui/icons'

/** A branch's kind (ADR 0009); '' is an ordinary branch. */
export type BranchKind = '' | 'trunk' | 'release' | 'environment' | 'docs' | 'bot'

export interface KindInfo {
  label: string
  /** The sidebar section's title. */
  section: string
  icon: IconName
  /** The CSS token the kind's chip and icon are drawn in. */
  token: string
}

export const KINDS: Record<Exclude<BranchKind, ''>, KindInfo> = {
  trunk: { label: 'Trunk', section: 'Trunk', icon: 'trunk', token: '--link' },
  release: { label: 'Release', section: 'Releases', icon: 'package', token: '--accent-b' },
  environment: { label: 'Environment', section: 'Environments', icon: 'server', token: '--warning' },
  docs: { label: 'Docs', section: 'Docs', icon: 'book', token: '--accent-a' },
  bot: { label: 'Bot', section: 'Bots', icon: 'bot', token: '--ink-faint' },
}

/** The kind's info, or undefined for an ordinary branch (or an unknown kind). */
export function kindInfo(kind: string | undefined): KindInfo | undefined {
  return kind ? KINDS[kind as Exclude<BranchKind, ''>] : undefined
}

/** Kinds that ask before a force push overwrites them (ADR 0009). */
export const CONFIRM_FORCE_PUSH: readonly string[] = ['trunk', 'release', 'environment']

/** Kinds that ask, naming the kind, before they're deleted. */
export const CONFIRM_DELETE: readonly string[] = ['trunk', 'environment']

/** "main is a Trunk branch", for confirmations; '' for kinds that don't ask. */
export function kindWarning(name: string, kind: string | undefined, kinds: readonly string[]): string {
  const info = kindInfo(kind)
  if (!info || !kind || !kinds.includes(kind)) return ''
  return `${name} is ${/^[AEIOU]/.test(info.label) ? 'an' : 'a'} ${info.label} branch.`
}
