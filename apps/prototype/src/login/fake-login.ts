/*
 * Fake data. There is no backend behind the prototype: each function waits a
 * moment and answers as the real operation would, according to the scenarios
 * chosen in the prototype bar. This is the only file of the screen that reads
 * a scenario by name.
 */
import {
  type ScenarioGroup,
  chosenScenario,
  useChosenScenario,
  useScenarioStore,
} from "#stores/scenarios"

import type {
  CodeProblem,
  LoginAlert,
  LoginOpening,
  LoginProvider,
  LoginProviderId,
} from "./login-flow"

/** What the configuration turns on. */
const waysIn = {
  id: "login.waysIn",
  label: "Ways in",
  scenarios: [
    { id: "everything", label: "Everything" },
    { id: "oneProvider", label: "One provider" },
    { id: "loginCodeOnly", label: "Login code only" },
    { id: "providersOnly", label: "Providers only" },
    { id: "localProvider", label: "Local provider" },
  ],
} as const satisfies ScenarioGroup

/** How the fake data answers. */
const outcome = {
  id: "login.outcome",
  label: "Outcome",
  scenarios: [
    { id: "success", label: "Success" },
    { id: "wrongCode", label: "Wrong code" },
    { id: "expiredCode", label: "Expired code" },
    { id: "limitReached", label: "Limit reached" },
    { id: "accountNotRegistered", label: "Account not registered" },
    { id: "providerFailed", label: "Provider failed" },
    { id: "sessionEnded", label: "Session ended" },
  ],
} as const satisfies ScenarioGroup

/** The scenarios of the login screen, for its route to declare. */
export const loginScenarioGroups = [waysIn, outcome] as const

type WaysIn = (typeof waysIn)["scenarios"][number]["id"]
type Outcome = (typeof outcome)["scenarios"][number]["id"]

const google: LoginProvider = { id: "google", forDevelopment: false }
const microsoft: LoginProvider = { id: "microsoft", forDevelopment: false }
const github: LoginProvider = { id: "github", forDevelopment: false }
const local: LoginProvider = { id: "local", forDevelopment: true }

function waysInOf(
  scenario: WaysIn
): Pick<LoginOpening, "providers" | "loginCode"> {
  switch (scenario) {
    case "everything":
      return { providers: [google, microsoft, github], loginCode: true }
    case "oneProvider":
      return { providers: [github], loginCode: true }
    case "loginCodeOnly":
      return { providers: [], loginCode: true }
    case "providersOnly":
      return { providers: [google, microsoft, github], loginCode: false }
    case "localProvider":
      return { providers: [google, microsoft, github, local], loginCode: true }
  }
}

const openings = new Map<string, LoginOpening>()

/**
 * One object for each pair of scenarios, so the screen can tell that what it
 * reads changed by comparing two readings.
 */
function openingOf(ways: WaysIn, answers: Outcome): LoginOpening {
  const key = `${ways}/${answers}`
  let opening = openings.get(key)
  if (!opening) {
    opening = { ...waysInOf(ways), sessionEnded: answers === "sessionEnded" }
    openings.set(key, opening)
  }
  return opening
}

/**
 * What the screen reads before anyone is logged in: the ways in that are
 * turned on, and whether it opens because a session ended.
 */
export function loginOpening(): LoginOpening {
  return openingOf(chosenScenario(waysIn), chosenScenario(outcome))
}

/** The same, kept current: a new object when a scenario changes. */
export function useLoginOpening(): LoginOpening {
  return openingOf(useChosenScenario(waysIn), useChosenScenario(outcome))
}

/** The third wrong attempt ends a login code (ADR-0011). */
const attemptsOfACode = 3

let wrongAttempts = 0

// A scenario chosen again starts from the beginning.
useScenarioStore.subscribe(() => {
  wrongAttempts = 0
})

function answerAfter(milliseconds: number) {
  return new Promise<void>((resolve) => setTimeout(resolve, milliseconds))
}

/**
 * Asks for a login code. The answer is the same whoever the e-mail belongs
 * to, and a new code ends the one before it.
 */
export async function askForLoginCode(
  _email: string
): Promise<{ outcome: "onItsWay" } | { outcome: "limitReached" }> {
  const scenario = chosenScenario(outcome)
  await answerAfter(700)
  if (scenario === "limitReached") {
    return { outcome: "limitReached" }
  }
  wrongAttempts = 0
  return { outcome: "onItsWay" }
}

/** Checks a login code. */
export async function checkLoginCode(
  _code: string
): Promise<
  { outcome: "loggedIn" } | { outcome: "refused"; problem: CodeProblem }
> {
  const scenario = chosenScenario(outcome)
  await answerAfter(700)
  switch (scenario) {
    case "wrongCode": {
      wrongAttempts += 1
      const attemptsLeft = attemptsOfACode - wrongAttempts
      return {
        outcome: "refused",
        problem:
          attemptsLeft > 0
            ? { kind: "wrong", attemptsLeft }
            : { kind: "ended" },
      }
    }
    case "expiredCode":
      return { outcome: "refused", problem: { kind: "expired" } }
    default:
      return { outcome: "loggedIn" }
  }
}

/** Goes to a login provider and comes back. The prototype shows no page of the provider. */
export async function logInThrough(
  provider: LoginProviderId
): Promise<
  { outcome: "loggedIn" } | { outcome: "refused"; alert: LoginAlert }
> {
  const scenario = chosenScenario(outcome)
  await answerAfter(900)
  switch (scenario) {
    case "accountNotRegistered":
      return { outcome: "refused", alert: { kind: "accountNotRegistered" } }
    case "providerFailed":
      return { outcome: "refused", alert: { kind: "providerFailed", provider } }
    default:
      return { outcome: "loggedIn" }
  }
}
