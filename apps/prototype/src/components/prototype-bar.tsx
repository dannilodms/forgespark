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

import { useLanguage, useLanguageStore } from "#stores/language"
import {
  type ScenarioGroup,
  useChosenScenario,
  useScenarioStore,
} from "#stores/scenarios"
import { useTheme, useThemeStore } from "#stores/theme"

import { supportedLanguages } from "../locales/languages"

/**
 * One choice of the bar: its name, and what holds now. The list opens upwards,
 * since the bar sits at the bottom of the screen.
 */
function BarSelect<Value extends string>({
  label,
  items,
  value,
  onChoose,
}: {
  label: string
  items: readonly { value: Value; label: string }[]
  value: Value
  onChoose: (value: Value) => void
}) {
  return (
    <Select
      items={items}
      value={value}
      onValueChange={(chosen) => {
        if (chosen) {
          onChoose(chosen)
        }
      }}
    >
      <SelectTrigger
        size="sm"
        aria-label={label}
        className="border-transparent text-foreground hover:bg-accent"
      >
        <span aria-hidden="true" className="text-muted-foreground">
          {label}
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

const languages = supportedLanguages.map(({ code, name }) => ({
  value: code,
  label: name,
}))

/** The language of every route. Each language is listed under its own name. */
function LanguageSelect() {
  const language = useLanguage()
  const choose = useLanguageStore((state) => state.choose)

  return (
    <BarSelect
      label="Language"
      items={languages}
      value={language}
      onChoose={choose}
    />
  )
}

/** One group of scenarios: its name, and the scenario that holds now. */
function ScenarioSelect({ group }: { group: ScenarioGroup }) {
  const chosen = useChosenScenario(group)
  const choose = useScenarioStore((state) => state.choose)
  const items = group.scenarios.map(({ id, label }) => ({ value: id, label }))

  return (
    <BarSelect
      label={group.label}
      items={items}
      value={chosen}
      onChoose={(scenario) => choose(group.id, scenario)}
    />
  )
}

/**
 * The tool of the prototype, outside every page: it holds what a person
 * changes to look at a screen, never something the screen itself offers.
 * Its labels stay in English, whatever the language of the screen.
 *
 * Its edge is the one `shadow-floating` draws: it adds no border of its own.
 *
 * It holds the theme and the language, which apply to every route. Then it
 * lists the scenarios the current route declares in its `staticData`, one
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
      className="flex max-w-full flex-wrap items-center justify-center gap-x-2 rounded-control bg-card py-1 pr-1 pl-3 type-caption text-muted-foreground shadow-floating"
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
      <LanguageSelect />
      {scenarioGroups.map((group) => (
        <ScenarioSelect key={group.id} group={group} />
      ))}
    </aside>
  )
}
