import { BranchAction, Integrate, PushRef, Refs, StashAction } from '../../wailsjs/go/app/App.js'
import type { MenuItem } from './ContextMenu.svelte'
import type { DialogSpec } from './Dialog.svelte'
import type { Ref, Row } from './graph/rowCache'
import { errorText } from './errors'
import { session } from './session.svelte'

/** What the menus need from the view that opens them. */
export interface MenuContext {
  path: string
  /** The checked-out branch, or "" when HEAD is detached. */
  current: () => string
  openDialog: (spec: DialogSpec) => void
}

const integrate = (path: string, action: string, ...args: string[]) => () =>
  session.run(() => Integrate(path, action, args))

/** Merge and rebase items for integrating target (a ref name or commit) with the current branch. */
function integrateItems(ctx: MenuContext, target: string, label: string): MenuItem[] {
  const cur = ctx.current()
  if (!cur) return []
  return [
    { label: `Merge ${label} into ${cur}`, run: integrate(ctx.path, 'merge', target) },
    { label: `Merge ${label} into ${cur} (no fast-forward)`, run: integrate(ctx.path, 'merge', target, 'no-ff') },
    { label: `Rebase ${cur} onto ${label}`, run: integrate(ctx.path, 'rebase', target) },
  ]
}

/** Runs a backend call for a dialog: an error message keeps the dialog open. */
async function attempt(fn: () => Promise<unknown>): Promise<string | void> {
  try {
    await fn()
  } catch (e) {
    return errorText(e)
  }
}

const branch = (path: string, action: string, ...args: string[]) => BranchAction(path, action, args)

/** The remote a local branch pushes to: its upstream's remote, else "origin". */
function remoteOf(ref: Ref): { remote: string; branch: string } {
  const [remote, ...rest] = ref.upstream.split('/')
  return ref.upstream ? { remote, branch: rest.join('/') } : { remote: 'origin', branch: ref.short }
}

const push = (path: string, remote: string, kind: string, name: string) => () => session.run(() => PushRef(path, remote, kind, name))

/** Menu items for one ref (branch, remote branch, tag or stash). */
export function refItems(ref: Ref, ctx: MenuContext): MenuItem[] {
  const { path, openDialog } = ctx
  switch (ref.kind) {
    case 'local':
      return [
        ...(ref.head ? [] : [{ label: `Checkout ${ref.short}`, run: () => session.run(() => branch(path, 'checkout', ref.short)) }]),
        ...(ref.head ? [] : integrateItems(ctx, ref.short, ref.short)),
        ...(ref.head
          ? []
          : [
              {
                label: `Push ${ref.short} to ${remoteOf(ref).remote}`,
                run: push(path, remoteOf(ref).remote, 'branch', ref.short),
              },
            ]),
        {
          label: `Rename ${ref.short}…`,
          run: () =>
            openDialog({
              title: `Rename branch ${ref.short}`,
              fields: [{ key: 'name', label: 'New name', value: ref.short }],
              submit: 'Rename',
              run: (v) => attempt(() => branch(path, 'rename', ref.short, String(v.name).trim())),
            }),
        },
        {
          label: `Set upstream of ${ref.short}…`,
          run: async () => {
            const refs = ((await session.run(() => Refs(path))) ?? []) as Ref[]
            const remotes = refs.filter((r) => r.kind === 'remote').map((r) => ({ value: r.short, label: r.short }))
            openDialog({
              title: `Upstream for ${ref.short}`,
              message: ref.upstream ? `Currently tracking ${ref.upstream}.` : 'Not tracking a remote branch.',
              fields: [
                {
                  key: 'upstream',
                  label: 'Track',
                  type: 'select',
                  value: ref.upstream,
                  options: [{ value: '', label: '(none)' }, ...remotes],
                },
              ],
              submit: 'Save',
              run: (v) => attempt(() => branch(path, 'upstream', ref.short, String(v.upstream))),
            })
          },
        },
        ...(ref.head
          ? []
          : [
              {
                label: `Delete ${ref.short}…`,
                run: () =>
                  openDialog({
                    title: `Delete branch ${ref.short}?`,
                    message: 'The branch is removed locally. Its commits stay reachable from other branches if merged.',
                    fields: [{ key: 'force', label: 'Delete even if it has unmerged commits', type: 'checkbox' }],
                    submit: 'Delete',
                    danger: true,
                    run: (v) => attempt(() => branch(path, 'delete', ref.short, v.force ? 'true' : '')),
                  }),
              },
            ]),
      ]
    case 'remote':
      {
        const [remote, ...rest] = ref.short.split('/')
        return [
          { label: `Checkout ${ref.short}`, run: () => session.run(() => branch(path, 'checkout', ref.short)) },
          ...integrateItems(ctx, ref.short, ref.short),
          { label: `Delete ${ref.short} from ${remote}`, danger: true, run: push(path, remote, 'delete-branch', rest.join('/')) },
        ]
      }
    case 'tag':
      return [
        ...integrateItems(ctx, ref.name, ref.short),
        { label: `Push tag ${ref.short} to origin`, run: push(path, 'origin', 'tag', ref.short) },
        {
          label: `Delete tag ${ref.short}`,
          danger: true,
          run: () => session.run(() => branch(path, 'untag', ref.short)),
        },
        { label: `Delete tag ${ref.short} from origin`, danger: true, run: push(path, 'origin', 'delete-tag', ref.short) },
      ]
    case 'stash': {
      const act = (action: string) => () => session.run(() => StashAction(path, ref.commit, action))
      return [
        { label: 'Apply stash', run: act('apply') },
        { label: 'Pop stash', run: act('pop') },
        { label: 'Drop stash', danger: true, run: act('drop') },
      ]
    }
  }
}

/** Menu items for a commit row, before the items for refs pointing at it. */
export function commitItems(row: Row, refs: Ref[], ctx: MenuContext): MenuItem[] {
  const { path, openDialog } = ctx
  if (refs.some((r) => r.kind === 'stash')) return []
  const short = row.id.slice(0, 7)
  return [
    {
      label: 'Create branch here…',
      run: () =>
        openDialog({
          title: `New branch at ${short}`,
          fields: [
            { key: 'name', label: 'Branch name', placeholder: 'feature/my-change' },
            { key: 'checkout', label: 'Check it out', type: 'checkbox', value: true },
          ],
          submit: 'Create branch',
          run: (v) => attempt(() => branch(path, 'create', String(v.name).trim(), row.id, v.checkout ? 'true' : '')),
        }),
    },
    {
      label: 'Create tag here…',
      run: () =>
        openDialog({
          title: `New tag at ${short}`,
          fields: [
            { key: 'name', label: 'Tag name', placeholder: 'v1.0.0' },
            { key: 'message', label: 'Message (makes an annotated tag)', type: 'textarea' },
          ],
          submit: 'Create tag',
          run: (v) => attempt(() => branch(path, 'tag', String(v.name).trim(), row.id, String(v.message))),
        }),
    },
    {
      label: `Checkout ${short} (detached)`,
      run: () => session.run(() => branch(path, 'checkout', row.id)),
    },
    ...(ctx.current() && !refs.some((r) => r.head)
      ? [{ label: `Cherry-pick ${short} onto ${ctx.current()}`, run: integrate(path, 'cherry-pick', row.id) }]
      : []),
  ]
}

/**
 * Items for dropping one graph row onto another. One side must be the checked-out branch:
 * dropping something onto it brings that thing in; dragging it onto something moves it.
 */
export function dropItems(source: Row, sourceRefs: Ref[], target: Row, targetRefs: Ref[], ctx: MenuContext): MenuItem[] {
  const cur = ctx.current()
  if (!cur || source.id === target.id) return []
  const named = (refs: Ref[]) => refs.filter((r) => r.kind !== 'stash' && !r.head)
  const describe = (row: Row, refs: Ref[]) => {
    const r = named(refs)[0]
    return r ? { target: r.kind === 'tag' ? r.name : r.short, label: r.short } : { target: row.id, label: row.id.slice(0, 7) }
  }
  if (targetRefs.some((r) => r.head)) {
    const s = describe(source, sourceRefs)
    return [
      ...integrateItems(ctx, s.target, s.label),
      ...(named(sourceRefs).length === 0
        ? [{ label: `Cherry-pick ${s.label} onto ${cur}`, run: integrate(ctx.path, 'cherry-pick', source.id) }]
        : []),
    ]
  }
  if (sourceRefs.some((r) => r.head)) {
    const t = describe(target, targetRefs)
    return integrateItems(ctx, t.target, t.label)
  }
  return []
}

/** The dialog for "new branch from HEAD". */
export function newBranchDialog(ctx: MenuContext): DialogSpec {
  return {
    title: 'New branch',
    message: 'Starts at the current commit (HEAD).',
    fields: [
      { key: 'name', label: 'Branch name', placeholder: 'feature/my-change' },
      { key: 'checkout', label: 'Check it out', type: 'checkbox', value: true },
    ],
    submit: 'Create branch',
    run: (v) => attempt(() => branch(ctx.path, 'create', String(v.name).trim(), '', v.checkout ? 'true' : '')),
  }
}

/** Double-click on a ref in the sidebar checks it out, as in GitKraken. */
export function checkoutRef(path: string, ref: Ref) {
  if (ref.kind === 'stash' || ref.head) return
  void session.run(() => branch(path, 'checkout', ref.kind === 'tag' ? ref.name : ref.short))
}
