const systemDark = window.matchMedia("(prefers-color-scheme: dark)")

function applyTheme() {
  document.documentElement.classList.toggle("dark", systemDark.matches)
}

/**
 * Puts the theme of the system on the root element and keeps it there. Call
 * once, before the first render.
 */
export function startTheme() {
  applyTheme()
  systemDark.addEventListener("change", applyTheme)
}
