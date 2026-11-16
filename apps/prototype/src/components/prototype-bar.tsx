import { useMatches } from "@tanstack/react-router"
import { Button } from "@workspace/ui/components/button"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import { MoonIcon, SunIcon } from "lucide-react"

import {
  type ScenarioGroup,
  useChosenScenario,
  useScenarioStore,
} from "#stores/scenarios"
import { useTheme, useThemeStore } from "#stores/theme"

/** One group of scenarios: its name, and the scenario that holds now. */
function ScenarioSelect({ group }: { group: ScenarioGroup }) {
  const chosen = useChosenScenario(group)
  const choose = useScenarioStore((state) => state.choose)
  const items = group.scenarios.map(({ id, label }) => ({ value: id, label }))

  return (
    <Select
      items={items}
      value={chosen}
      onValueChange={(scenario) => {
        if (scenario) {
          choose(group.id, scenario)
        }
      }}
    >
      <SelectTrigger
        size="sm"
        aria-label={group.label}
        className="border-transparent text-foreground hover:bg-accent"
      >
        <span aria-hidden="true" className="text-muted-foreground">
          {group.label}
        </span>
        <SelectValue />
      </SelectTrigger>
      <SelectContent
        side="top"
        align="start"
        alignItemWithTrigger={false}
        className="w-auto min-w-(--anchor-width)"
      >
        <SelectGroup>
          {items.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  )
}

/**
 * The tool of the prototype, outside every page: it holds what a person
 * changes to look at a screen, never something the screen itself offers.
 * Its labels stay in English.
 *
 * It lists the scenarios the current route declares in its `staticData`, one
 * control for each group, and none on a route that declares no scenario.
 */
export function PrototypeBar() {
  const theme = useTheme()
  const choose = useThemeStore((state) => state.choose)
  const other = theme === "dark" ? "light" : "dark"
  const scenarioGroups = useMatches().flatMap(
    (match) => match.staticData.scenarioGroups ?? []
  )

  return (
    <aside
      aria-label="Prototype bar"
      className="flex max-w-full flex-wrap items-center justify-center gap-x-2 rounded-control border bg-card py-1 pr-1 pl-3 type-caption text-muted-foreground shadow-floating"
    >
      <span className="type-code">Prototype</span>
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
      {scenarioGroups.map((group) => (
        <ScenarioSelect key={group.id} group={group} />
      ))}
    </aside>
  )
}
