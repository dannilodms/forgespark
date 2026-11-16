/**
 * The login providers the screen can offer. `local` is the local provider,
 * which exists for development and tests only (ADR-0011).
 */
export type LoginProviderId = "google" | "microsoft" | "github" | "local"

export type LoginProvider = {
  id: LoginProviderId
  /** Whether the provider exists for development only, which its button says. */
  forDevelopment: boolean
}

/** One login code can be asked for each minute (ADR-0011). */
export const secondsBetweenCodes = 60

export const loginCodeLength = 6

/** Why a login code was refused. An ended or expired code is over. */
export type CodeProblem =
  | { kind: "wrong"; attemptsLeft: number }
  | { kind: "ended" }
  | { kind: "expired" }

/** What an Alert above the form reports. */
export type LoginAlert =
  | { kind: "sessionEnded" }
  | { kind: "limitReached" }
  | { kind: "accountNotRegistered" }
  | { kind: "providerFailed"; provider: LoginProviderId }

/**
 * The login screen is one flow of named steps. `waiting` names the answer the
 * screen is waiting for, so one button shows the spinner and none can be
 * pressed twice.
 */
export type LoginState =
  | {
      step: "start"
      email: string
      emailInvalid: boolean
      alert: LoginAlert | null
      waiting: null | "code" | LoginProviderId
    }
  | {
      step: "code"
      email: string
      code: string
      codeProblem: CodeProblem | null
      alert: LoginAlert | null
      /** When a new code can be asked for, in milliseconds since the epoch. */
      newCodeAt: number
      waiting: null | "login" | "newCode"
    }

/** The part of what the screen reads that decides where the flow can be. */
type FlowOpening = Pick<LoginOpening, "loginCode" | "sessionEnded">

export type LoginAction =
  | { type: "reopened"; opening: FlowOpening }
  | { type: "emailTyped"; email: string }
  | { type: "emailRefused" }
  | { type: "codeAsked" }
  | { type: "codeOnItsWay"; at: number }
  | { type: "limitReached" }
  | { type: "providerChosen"; provider: LoginProviderId }
  | { type: "providerRefused"; alert: LoginAlert }
  | { type: "codeTyped"; code: string }
  | { type: "loginAsked" }
  | { type: "codeRefused"; problem: CodeProblem }
  | { type: "newCodeAsked" }
  | { type: "otherEmailWanted" }

/** What the screen reads before anyone is logged in. */
export type LoginOpening = {
  /** The login providers that are turned on, in the order the screen shows them. */
  providers: readonly LoginProvider[]
  /** Whether the login code is turned on. */
  loginCode: boolean
  /** Whether the screen opens because a session ended. */
  sessionEnded: boolean
}

/** The Start step as it opens: nothing awaited, and no alert but the one of an ended session. */
function startStep(email: string, opening: FlowOpening): LoginState {
  return {
    step: "start",
    email,
    emailInvalid: false,
    alert: opening.sessionEnded ? { kind: "sessionEnded" } : null,
    waiting: null,
  }
}

/** The state of a screen that has just opened. */
export function openLogin(opening: FlowOpening): LoginState {
  return startStep("", opening)
}

/** Whether the login code can no longer be typed, so a new one has to be asked for. */
export function codeIsOver(state: LoginState) {
  return (
    state.step === "code" &&
    (state.codeProblem?.kind === "ended" ||
      state.codeProblem?.kind === "expired")
  )
}

export function isEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim())
}

export function loginReducer(
  state: LoginState,
  action: LoginAction
): LoginState {
  /*
   * What the screen reads changed while it was open. The step stays when it
   * is still allowed, and whatever an earlier answer left behind goes away:
   * an alert, an awaited answer, a code that was over.
   */
  if (action.type === "reopened") {
    if (
      state.step === "code" &&
      action.opening.loginCode &&
      !action.opening.sessionEnded
    ) {
      return {
        ...state,
        code: "",
        codeProblem: null,
        alert: null,
        waiting: null,
      }
    }
    return startStep(state.email, action.opening)
  }

  if (state.step === "start") {
    switch (action.type) {
      case "emailTyped":
        return { ...state, email: action.email, emailInvalid: false }
      case "emailRefused":
        return { ...state, emailInvalid: true }
      case "codeAsked":
        return { ...state, alert: null, waiting: "code" }
      case "providerChosen":
        return { ...state, alert: null, waiting: action.provider }
      case "providerRefused":
        if (state.waiting === null || state.waiting === "code") {
          return state
        }
        return { ...state, alert: action.alert, waiting: null }
      case "limitReached":
        if (state.waiting !== "code") {
          return state
        }
        return { ...state, alert: { kind: "limitReached" }, waiting: null }
      case "codeOnItsWay":
        if (state.waiting !== "code") {
          return state
        }
        return {
          step: "code",
          email: state.email.trim(),
          code: "",
          codeProblem: null,
          alert: null,
          newCodeAt: action.at + secondsBetweenCodes * 1000,
          waiting: null,
        }
      default:
        return state
    }
  }

  switch (action.type) {
    case "codeTyped":
      return { ...state, code: action.code }
    case "loginAsked":
      return { ...state, waiting: "login" }
    case "codeRefused":
      if (state.waiting !== "login") {
        return state
      }
      return { ...state, code: "", codeProblem: action.problem, waiting: null }
    case "newCodeAsked":
      return { ...state, alert: null, waiting: "newCode" }
    case "limitReached":
      if (state.waiting !== "newCode") {
        return state
      }
      return { ...state, alert: { kind: "limitReached" }, waiting: null }
    case "codeOnItsWay":
      if (state.waiting !== "newCode") {
        return state
      }
      return {
        ...state,
        code: "",
        codeProblem: null,
        alert: null,
        newCodeAt: action.at + secondsBetweenCodes * 1000,
        waiting: null,
      }
    case "otherEmailWanted":
      return {
        step: "start",
        email: state.email,
        emailInvalid: false,
        alert: null,
        waiting: null,
      }
    default:
      return state
  }
}
