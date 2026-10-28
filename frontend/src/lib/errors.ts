/** A backend rejection as text, without JavaScript's "Error: " prefix. */
export function errorText(e: unknown): string {
  return e instanceof Error ? e.message : String(e)
}
