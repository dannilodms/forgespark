import {
  afterEach,
  beforeEach,
  describe,
  expect,
  test,
  vi,
} from "vite-plus/test"

import { useScenarioStore } from "#stores/scenarios"

import {
  askForLoginCode,
  checkLoginCode,
  logInThrough,
  loginOpening,
  loginScenarioGroups,
} from "./fake-login"

const [waysIn, outcome] = loginScenarioGroups

/** Chooses a scenario as the prototype bar does, by the label it shows. */
function choose(group: (typeof loginScenarioGroups)[number], label: string) {
  const scenario = group.scenarios.find((each) => each.label === label)
  if (!scenario) {
    throw new Error(`${group.label} has no scenario "${label}"`)
  }
  useScenarioStore.getState().choose(group.id, scenario.id)
}

/** The answer of the fake data, without waiting for its delay. */
async function answerOf<Answer>(asked: Promise<Answer>) {
  await vi.runAllTimersAsync()
  return asked
}

beforeEach(() => {
  vi.useFakeTimers()
  useScenarioStore.setState({ chosen: {} })
})

afterEach(() => {
  vi.useRealTimers()
})

describe("the scenarios the login screen declares", () => {
  test("are two groups, with the scenarios of the spec in its order", () => {
    expect(
      loginScenarioGroups.map((group) => [
        group.label,
        group.scenarios.map((scenario) => scenario.label),
      ])
    ).toEqual([
      [
        "Ways in",
        [
          "Everything",
          "One provider",
          "Login code only",
          "Providers only",
          "Local provider",
        ],
      ],
      [
        "Outcome",
        [
          "Success",
          "Wrong code",
          "Expired code",
          "Limit reached",
          "Account not registered",
          "Provider failed",
          "Session ended",
        ],
      ],
    ])
  })
})

describe("Ways in", () => {
  const ways = () => {
    const { providers, loginCode } = loginOpening()
    return { providers: providers.map((provider) => provider.id), loginCode }
  }

  test("Everything is the default: the login code and the three providers", () => {
    expect(ways()).toEqual({
      providers: ["google", "microsoft", "github"],
      loginCode: true,
    })
  })

  test("One provider: the login code and GitHub", () => {
    choose(waysIn, "One provider")
    expect(ways()).toEqual({ providers: ["github"], loginCode: true })
  })

  test("Login code only: no provider", () => {
    choose(waysIn, "Login code only")
    expect(ways()).toEqual({ providers: [], loginCode: true })
  })

  test("Providers only: the three providers and no login code", () => {
    choose(waysIn, "Providers only")
    expect(ways()).toEqual({
      providers: ["google", "microsoft", "github"],
      loginCode: false,
    })
  })

  test("Local provider: everything, with the local provider last", () => {
    choose(waysIn, "Local provider")
    expect(ways()).toEqual({
      providers: ["google", "microsoft", "github", "local"],
      loginCode: true,
    })
  })

  test("only the local provider is marked as for development", () => {
    choose(waysIn, "Local provider")
    expect(
      loginOpening()
        .providers.filter((provider) => provider.forDevelopment)
        .map((provider) => provider.id)
    ).toEqual(["local"])
  })
})

describe("Outcome", () => {
  test("Success is the default: any code and any provider log in", async () => {
    expect(await answerOf(askForLoginCode("ana@example.com"))).toEqual({
      outcome: "onItsWay",
    })
    expect(await answerOf(checkLoginCode("000000"))).toEqual({
      outcome: "loggedIn",
    })
    expect(await answerOf(logInThrough("microsoft"))).toEqual({
      outcome: "loggedIn",
    })
    expect(loginOpening().sessionEnded).toBe(false)
  })

  test("Wrong code: two attempts say how many are left, the third ends the code", async () => {
    choose(outcome, "Wrong code")
    await answerOf(askForLoginCode("ana@example.com"))

    expect(await answerOf(checkLoginCode("111111"))).toEqual({
      outcome: "refused",
      problem: { kind: "wrong", attemptsLeft: 2 },
    })
    expect(await answerOf(checkLoginCode("222222"))).toEqual({
      outcome: "refused",
      problem: { kind: "wrong", attemptsLeft: 1 },
    })
    expect(await answerOf(checkLoginCode("333333"))).toEqual({
      outcome: "refused",
      problem: { kind: "ended" },
    })
  })

  test("Wrong code: a new code has its own three attempts", async () => {
    choose(outcome, "Wrong code")
    await answerOf(askForLoginCode("ana@example.com"))
    await answerOf(checkLoginCode("111111"))
    await answerOf(checkLoginCode("222222"))
    await answerOf(askForLoginCode("ana@example.com"))

    expect(await answerOf(checkLoginCode("333333"))).toEqual({
      outcome: "refused",
      problem: { kind: "wrong", attemptsLeft: 2 },
    })
  })

  test("Wrong code: choosing the scenario again starts the count again", async () => {
    choose(outcome, "Wrong code")
    await answerOf(checkLoginCode("111111"))
    choose(outcome, "Success")
    choose(outcome, "Wrong code")

    expect(await answerOf(checkLoginCode("222222"))).toEqual({
      outcome: "refused",
      problem: { kind: "wrong", attemptsLeft: 2 },
    })
  })

  test("Expired code: logging in answers that the code expired", async () => {
    choose(outcome, "Expired code")

    expect(await answerOf(askForLoginCode("ana@example.com"))).toEqual({
      outcome: "onItsWay",
    })
    expect(await answerOf(checkLoginCode("123456"))).toEqual({
      outcome: "refused",
      problem: { kind: "expired" },
    })
  })

  test("Limit reached: asking for a code is refused, and a provider still logs in", async () => {
    choose(outcome, "Limit reached")

    expect(await answerOf(askForLoginCode("ana@example.com"))).toEqual({
      outcome: "limitReached",
    })
    expect(await answerOf(logInThrough("google"))).toEqual({
      outcome: "loggedIn",
    })
  })

  test("Account not registered: a provider comes back with no access", async () => {
    choose(outcome, "Account not registered")

    expect(await answerOf(logInThrough("google"))).toEqual({
      outcome: "refused",
      alert: { kind: "accountNotRegistered" },
    })
    expect(await answerOf(checkLoginCode("123456"))).toEqual({
      outcome: "loggedIn",
    })
  })

  test("Provider failed: the answer names the provider", async () => {
    choose(outcome, "Provider failed")

    expect(await answerOf(logInThrough("github"))).toEqual({
      outcome: "refused",
      alert: { kind: "providerFailed", provider: "github" },
    })
  })

  test("Session ended: the screen opens knowing it, and logging in works", async () => {
    choose(outcome, "Session ended")

    expect(loginOpening().sessionEnded).toBe(true)
    expect(await answerOf(logInThrough("github"))).toEqual({
      outcome: "loggedIn",
    })
  })
})

describe("what the screen reads", () => {
  test("is the same object until a scenario changes, and another one after", () => {
    const before = loginOpening()
    expect(loginOpening()).toBe(before)

    choose(outcome, "Wrong code")
    const after = loginOpening()
    expect(after).not.toBe(before)
    expect(after).toEqual(before)
  })
})
