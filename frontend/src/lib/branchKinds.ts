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
