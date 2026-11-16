import { create } from "zustand"

/**
 * A scenario is a named condition a screen can be put in. A screen declares
 * its scenarios in groups, one group for each thing that can vary on its own,
 * and exactly one scenario of a group holds at a time. The first scenario of
 * a group is its default.
 *
 * The route of the screen carries the groups in its `staticData`, the
 * prototype bar lists them, and the fake data of the screen answers according
 * to the scenario chosen in each. Labels are in English, as the bar is.
 */
export type ScenarioGroup<ScenarioId extends string = string> = {
  /** Unique in the prototype: prefix it with the name of the screen. */
  id: string
  label: string
  scenarios: readonly [Scenario<ScenarioId>, ...Scenario<ScenarioId>[]]
}

type Scenario<ScenarioId extends string> = { id: ScenarioId; label: string }

declare module "@tanstack/react-router" {
  interface StaticDataRouteOption {
    /** The scenarios of the screen this route opens. */
    scenarioGroups?: readonly ScenarioGroup[]
  }
}

type ScenarioState = {
  /** The scenario chosen in each group, by the id of the group. */
  chosen: Record<string, string>
  choose: (groupId: string, scenarioId: string) => void
}

/** Not persisted: a choice lasts until the page is reloaded. */
export const useScenarioStore = create<ScenarioState>()((set) => ({
  chosen: {},
  choose: (groupId, scenarioId) =>
    set((state) => ({ chosen: { ...state.chosen, [groupId]: scenarioId } })),
}))

function scenarioIn<ScenarioId extends string>(
  group: ScenarioGroup<ScenarioId>,
  chosen: Record<string, string>
): ScenarioId {
  const scenario = group.scenarios.find(({ id }) => id === chosen[group.id])
  return (scenario ?? group.scenarios[0]).id
}

/** The scenario chosen in a group, read once. For the fake data of a screen. */
export function chosenScenario<ScenarioId extends string>(
  group: ScenarioGroup<ScenarioId>
): ScenarioId {
  return scenarioIn(group, useScenarioStore.getState().chosen)
}

/** The scenario chosen in a group, kept current. */
export function useChosenScenario<ScenarioId extends string>(
  group: ScenarioGroup<ScenarioId>
): ScenarioId {
  return useScenarioStore((state) => scenarioIn(group, state.chosen))
}
