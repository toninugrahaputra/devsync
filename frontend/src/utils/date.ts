// due_date/start_date are date-only concepts (no meaningful time-of-day), but
// the backend returns them as full RFC3339 timestamps at UTC midnight (e.g.
// "2026-12-31T00:00:00Z"). Formatting that through the browser's local
// timezone shifts the calendar day backward for anyone west of UTC — so
// these always format in UTC to show the date that was actually picked.
export function formatDate(isoDate: string, options?: Intl.DateTimeFormatOptions): string {
  return new Date(isoDate).toLocaleDateString(undefined, { ...options, timeZone: 'UTC' })
}
