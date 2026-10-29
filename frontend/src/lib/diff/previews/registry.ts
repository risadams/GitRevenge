import type { Component } from 'svelte'
import MarkdownPreview from './MarkdownPreview.svelte'

/** What every previewer is given: the file, its text and ways to reach the rest of the repo. */
export interface PreviewProps {
  path: string
  text: string
  /** Loads an image in the repository, at the same version, as a data URL ("" if none). */
  loadImage: (file: string) => Promise<string>
  /** Opens another file in the repository, at the same version. */
  openFile: (file: string) => void
}

export interface Previewer {
  id: string
  matches: (path: string) => boolean
  component: Component<PreviewProps>
}

/**
 * Rendered previews for File mode, by file type. Adding a type is one entry here and one
 * component. SVG isn't listed: it's classified as an image and already shows through an
 * <img> element, where its scripts never run.
 */
export const PREVIEWERS: Previewer[] = [
  { id: 'markdown', matches: (p) => /\.(md|markdown|mdown|mkd|mkdn)$/i.test(p), component: MarkdownPreview },
]

export function previewerFor(path: string): Previewer | undefined {
  return PREVIEWERS.find((p) => p.matches(path))
}
