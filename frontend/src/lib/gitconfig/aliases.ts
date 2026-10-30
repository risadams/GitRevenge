import { samePath, type Entry } from './rows'

// The Aliases view's model (#227): every alias git sees from the selected file's point of
// view, and what would stop each one working as intended.

/** One alias, set in the edited file or inherited from another one. */
export interface Alias {
  name: string
  /** What git runs: a git command line, or a shell command after a leading "!". */
  value: string
  /** Set in the edited file, rather than inherited. */
  own: boolean
  /** Where an inherited alias comes from: its scope and file. */
  scope: string
  file: string
  shell: boolean
  warnings: string[]
}

/** A valid alias name: git's rule for variable names. */
export const NAME = /^[A-Za-z][A-Za-z0-9-]*$/

const PREFIX = 'alias.'

function aliasName(key: string): string | undefined {
  // Section and variable names are case-insensitive; git reports them lower-cased.
  return key.toLowerCase().startsWith(PREFIX) && !key.slice(PREFIX.length).includes('.') ? key.slice(PREFIX.length).toLowerCase() : undefined
}

/** Builds the list for one file: its own aliases (from the working copy) first, then those
 * other files set that it doesn't, with warnings for each. builtins are git's own commands. */
export function buildAliases(own: Entry[], effective: Entry[], file: string, builtins: string[]): Alias[] {
  const out: Alias[] = []
  const mine = new Set<string>()
  for (const e of own) {
    const name = aliasName(e.key)
    if (!name) continue
    mine.add(name)
    out.push({ name, value: e.value, own: true, scope: '', file, shell: e.value.startsWith('!'), warnings: [] })
  }
  // The last value git reads for a name wins, so keep each inherited name's last one.
  const inherited = new Map<string, Alias>()
  for (const e of effective) {
    const name = aliasName(e.key)
    if (!name || samePath(e.file, file)) continue
    inherited.set(name, { name, value: e.value, own: false, scope: e.scope, file: e.file, shell: e.value.startsWith('!'), warnings: [] })
  }
  const builtin = new Set(builtins)
  for (const a of out) {
    const hidden = inherited.get(a.name)
    if (hidden) a.warnings.push(`Replaces the ${hidden.scope} alias ${a.name}, which runs ${hidden.value}.`)
  }
  for (const [name, a] of inherited) if (!mine.has(name)) out.push(a)
  for (const a of out) {
    if (builtin.has(a.name)) a.warnings.push(`Never runs: ${a.name} is a built-in command, and git ignores aliases with the same name.`)
    else if (callsItself(a)) a.warnings.push('Calls itself, so it would loop until git gives up.')
  }
  return out
}

/** Whether an alias runs itself: a git alias whose command is its own name, or a shell
 * alias that runs `git <name>`. */
export function callsItself(a: Pick<Alias, 'name' | 'value'>): boolean {
  if (a.value.startsWith('!')) return new RegExp(`(^|[\\s;&|(])git\\s+${a.name}(\\s|$|[;&|)])`).test(a.value.slice(1))
  return a.value.trim().split(/\s+/)[0]?.toLowerCase() === a.name
}
