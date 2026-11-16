import { useSyncExternalStore } from "react"
import { create } from "zustand"
import { persist } from "zustand/middleware"

export type Theme = "light" | "dark"

type ThemeState = {
  /** The theme chosen in the prototype bar. `null` follows the system. */
  chosen: Theme | null
  choose: (theme: Theme) => void
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      chosen: null,
      choose: (theme) => set({ chosen: theme }),
    }),
    { name: "prototype-theme" }
  )
)

const systemDark = window.matchMedia("(prefers-color-scheme: dark)")

function systemTheme(): Theme {
  return systemDark.matches ? "dark" : "light"
}

function subscribeToSystem(onChange: () => void) {
  systemDark.addEventListener("change", onChange)
  return () => systemDark.removeEventListener("change", onChange)
}

function applyTheme() {
  const theme = useThemeStore.getState().chosen ?? systemTheme()
  document.documentElement.classList.toggle("dark", theme === "dark")
}

/** Puts the theme on the root element and keeps it there. Call once, before the first render. */
export function startTheme() {
  applyTheme()
  useThemeStore.subscribe(applyTheme)
  systemDark.addEventListener("change", applyTheme)
}

/** The theme in force: the chosen one, or the system's when none was chosen. */
export function useTheme(): Theme {
  const chosen = useThemeStore((state) => state.chosen)
  const system = useSyncExternalStore(subscribeToSystem, systemTheme)
  return chosen ?? system
}
