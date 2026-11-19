import { useSyncExternalStore } from "react"

// Under 960px the sidebar of the App shell is a sheet (DESIGN.md). The same
// width is `--breakpoint-shell` in globals.css, which the `shell:` classes
// of the sidebar read: change both together.
const SHEET_BREAKPOINT = 960
const query = `(max-width: ${SHEET_BREAKPOINT - 1}px)`

function subscribe(onChange: () => void) {
  const mql = window.matchMedia(query)
  mql.addEventListener("change", onChange)
  return () => mql.removeEventListener("change", onChange)
}

/** Whether the screen is narrower than the width at which a sidebar becomes a sheet. */
export function useIsMobile() {
  return useSyncExternalStore(
    subscribe,
    () => window.matchMedia(query).matches,
    () => false
  )
}
