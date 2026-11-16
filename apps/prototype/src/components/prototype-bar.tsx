import { Button } from "@workspace/ui/components/button"
import { MoonIcon, SunIcon } from "lucide-react"

import { useTheme, useThemeStore } from "#stores/theme"

/**
 * The tool of the prototype, outside every page: it holds what a person
 * changes to look at a screen, never something the screen itself offers.
 * Its labels stay in English.
 */
export function PrototypeBar() {
  const theme = useTheme()
  const choose = useThemeStore((state) => state.choose)
  const other = theme === "dark" ? "light" : "dark"

  return (
    <aside
      aria-label="Prototype bar"
      className="flex items-center gap-2 rounded-control border bg-card py-1 pr-1 pl-3 text-xs text-muted-foreground shadow-floating"
    >
      <span className="font-mono">Prototype</span>
      <Button
        variant="ghost"
        size="sm"
        aria-label={`Theme: ${theme}. Switch to ${other}.`}
        onClick={() => choose(other)}
      >
        {theme === "dark" ? (
          <MoonIcon data-icon="inline-start" />
        ) : (
          <SunIcon data-icon="inline-start" />
        )}
        {theme === "dark" ? "Dark" : "Light"}
      </Button>
    </aside>
  )
}
