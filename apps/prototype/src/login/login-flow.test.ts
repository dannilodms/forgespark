import { describe, expect, test } from "vite-plus/test"

import {
  type LoginAction,
  type LoginState,
  codeIsOver,
  loginReducer,
  openLogin,
} from "./login-flow"

/** The state the flow reaches from a fresh screen after a list of actions. */
function after(...actions: LoginAction[]): LoginState {
  return actions.reduce(
    loginReducer,
    openLogin({ loginCode: true, sessionEnded: false })
  )
}

/** The actions that take a fresh screen to the Code step. */
const toTheCodeStep: LoginAction[] = [
  { type: "emailTyped", email: "ana@example.com" },
  { type: "codeAsked" },
  { type: "codeOnItsWay", at: 1_000 },
]

describe("a wrong login code", () => {
  test("says how many attempts are left and empties the boxes", () => {
    const state = after(
      ...toTheCodeStep,
      { type: "codeTyped", code: "123456" },
      { type: "loginAsked" },
      { type: "codeRefused", problem: { kind: "wrong", attemptsLeft: 2 } }
    )

    expect(state).toMatchObject({
      step: "code",
      code: "",
      codeProblem: { kind: "wrong", attemptsLeft: 2 },
      waiting: null,
    })
  })
})

describe("a login code that is over", () => {
  test("the third wrong attempt ends the code", () => {
    const state = after(
      ...toTheCodeStep,
      { type: "loginAsked" },
      { type: "codeRefused", problem: { kind: "ended" } }
    )

    expect(codeIsOver(state)).toBe(true)
    expect(state).toMatchObject({ step: "code", code: "", waiting: null })
  })

  test("an expired code is over too", () => {
    const state = after(
      ...toTheCodeStep,
      { type: "loginAsked" },
      { type: "codeRefused", problem: { kind: "expired" } }
    )

    expect(codeIsOver(state)).toBe(true)
  })

  test("a wrong code with attempts left is not over", () => {
    const state = after(
      ...toTheCodeStep,
      { type: "loginAsked" },
      { type: "codeRefused", problem: { kind: "wrong", attemptsLeft: 1 } }
    )

    expect(codeIsOver(state)).toBe(false)
  })

  test("a new code can be typed, and the wait starts again", () => {
    const state = after(
      ...toTheCodeStep,
      { type: "loginAsked" },
      { type: "codeRefused", problem: { kind: "ended" } },
      { type: "newCodeAsked" },
      { type: "codeOnItsWay", at: 90_000 }
    )

    expect(codeIsOver(state)).toBe(false)
    expect(state).toMatchObject({
      step: "code",
      codeProblem: null,
      newCodeAt: 150_000,
      waiting: null,
    })
  })
})

describe("an answer that refuses, on the Start step", () => {
  test("too many codes asked for: an alert, and the Start step stays", () => {
    const state = after(
      { type: "emailTyped", email: "ana@example.com" },
      { type: "codeAsked" },
      { type: "limitReached" }
    )

    expect(state).toMatchObject({
      step: "start",
      email: "ana@example.com",
      alert: { kind: "limitReached" },
      waiting: null,
    })
  })

  test("a login provider that refuses comes back with its alert", () => {
    const state = after(
      { type: "providerChosen", provider: "github" },
      {
        type: "providerRefused",
        alert: { kind: "providerFailed", provider: "github" },
      }
    )

    expect(state).toMatchObject({
      step: "start",
      alert: { kind: "providerFailed", provider: "github" },
      waiting: null,
    })
  })

  test("asking again takes the alert away while the answer is awaited", () => {
    const refused = after(
      { type: "providerChosen", provider: "google" },
      { type: "providerRefused", alert: { kind: "accountNotRegistered" } }
    )

    expect(
      loginReducer(refused, { type: "providerChosen", provider: "google" })
    ).toMatchObject({ alert: null, waiting: "google" })
    expect(loginReducer(refused, { type: "codeAsked" })).toMatchObject({
      alert: null,
      waiting: "code",
    })
  })
})

describe("asking for a new code", () => {
  test("too many codes asked for: an alert, and the code typed so far stays", () => {
    const state = after(
      ...toTheCodeStep,
      { type: "codeTyped", code: "12" },
      { type: "newCodeAsked" },
      { type: "limitReached" }
    )

    expect(state).toMatchObject({
      step: "code",
      code: "12",
      alert: { kind: "limitReached" },
      newCodeAt: 61_000,
      waiting: null,
    })
  })

  test("a new code on its way takes the alert away", () => {
    const state = after(
      ...toTheCodeStep,
      { type: "newCodeAsked" },
      { type: "limitReached" },
      { type: "newCodeAsked" },
      { type: "codeOnItsWay", at: 200_000 }
    )

    expect(state).toMatchObject({ step: "code", alert: null })
  })
})

describe("an answer nobody is waiting for", () => {
  test("is left out on the Start step", () => {
    const idle = after({ type: "emailTyped", email: "ana@example.com" })

    expect(loginReducer(idle, { type: "codeOnItsWay", at: 1_000 })).toBe(idle)
    expect(loginReducer(idle, { type: "limitReached" })).toBe(idle)
    expect(
      loginReducer(idle, {
        type: "providerRefused",
        alert: { kind: "accountNotRegistered" },
      })
    ).toBe(idle)
  })

  test("is left out on the Code step", () => {
    const idle = after(...toTheCodeStep)

    expect(
      loginReducer(idle, { type: "codeRefused", problem: { kind: "expired" } })
    ).toBe(idle)
    expect(loginReducer(idle, { type: "limitReached" })).toBe(idle)
    expect(loginReducer(idle, { type: "codeOnItsWay", at: 5_000 })).toBe(idle)
  })
})

describe("a session that ended", () => {
  test("opens the Start step with its alert", () => {
    expect(openLogin({ loginCode: true, sessionEnded: true })).toMatchObject({
      step: "start",
      alert: { kind: "sessionEnded" },
    })
  })

  test("without it, the Start step opens with no alert", () => {
    expect(openLogin({ loginCode: true, sessionEnded: false })).toMatchObject({
      step: "start",
      alert: null,
    })
  })
})

describe("what the screen reads changes while it is open", () => {
  const onTheCodeStep = after(
    ...toTheCodeStep,
    { type: "codeTyped", code: "123" },
    { type: "loginAsked" },
    { type: "codeRefused", problem: { kind: "ended" } }
  )

  test("the login code turned off: back to the Start step, e-mail kept", () => {
    const state = loginReducer(onTheCodeStep, {
      type: "reopened",
      opening: { loginCode: false, sessionEnded: false },
    })

    expect(state).toEqual({
      step: "start",
      email: "ana@example.com",
      emailInvalid: false,
      alert: null,
      waiting: null,
    })
  })

  test("the session ended: back to the Start step, with its alert", () => {
    const state = loginReducer(onTheCodeStep, {
      type: "reopened",
      opening: { loginCode: true, sessionEnded: true },
    })

    expect(state).toMatchObject({
      step: "start",
      email: "ana@example.com",
      alert: { kind: "sessionEnded" },
    })
  })

  test("the Code step stays when it is still allowed, with a code to type", () => {
    const state = loginReducer(onTheCodeStep, {
      type: "reopened",
      opening: { loginCode: true, sessionEnded: false },
    })

    expect(state).toMatchObject({
      step: "code",
      email: "ana@example.com",
      code: "",
      codeProblem: null,
      alert: null,
      newCodeAt: 61_000,
      waiting: null,
    })
  })

  test("the Start step keeps the e-mail and stops waiting", () => {
    const waiting = after(
      { type: "emailTyped", email: "ana@example.com" },
      { type: "codeAsked" }
    )
    const state = loginReducer(waiting, {
      type: "reopened",
      opening: { loginCode: false, sessionEnded: false },
    })

    expect(state).toMatchObject({
      step: "start",
      email: "ana@example.com",
      alert: null,
      waiting: null,
    })
    expect(loginReducer(state, { type: "codeOnItsWay", at: 1_000 })).toBe(state)
  })

  test("an alert of an earlier answer goes away", () => {
    const refused = after(
      { type: "providerChosen", provider: "google" },
      { type: "providerRefused", alert: { kind: "accountNotRegistered" } }
    )

    expect(
      loginReducer(refused, {
        type: "reopened",
        opening: { loginCode: true, sessionEnded: false },
      })
    ).toMatchObject({ step: "start", alert: null })
  })
})
