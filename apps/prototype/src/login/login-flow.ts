/** The login providers the screen can offer, in the order it shows them. */
export const loginProviders = [
  { id: "google", name: "Google" },
  { id: "microsoft", name: "Microsoft" },
  { id: "github", name: "GitHub" },
] as const

export type LoginProviderId = (typeof loginProviders)[number]["id"]

/** One login code can be asked for each minute (ADR-0011). */
export const secondsBetweenCodes = 60

export const loginCodeLength = 6

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
      waiting: null | "code" | LoginProviderId
    }
  | {
      step: "code"
      email: string
      code: string
      /** When a new code can be asked for, in milliseconds since the epoch. */
      newCodeAt: number
      waiting: null | "login" | "newCode"
    }

export type LoginAction =
  | { type: "emailTyped"; email: string }
  | { type: "emailRefused" }
  | { type: "codeAsked" }
  | { type: "codeOnItsWay"; at: number }
  | { type: "providerChosen"; provider: LoginProviderId }
  | { type: "codeTyped"; code: string }
  | { type: "loginAsked" }
  | { type: "newCodeAsked" }
  | { type: "otherEmailWanted" }

export const initialLoginState: LoginState = {
  step: "start",
  email: "",
  emailInvalid: false,
  waiting: null,
}

export function isEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim())
}

export function loginReducer(
  state: LoginState,
  action: LoginAction
): LoginState {
  if (state.step === "start") {
    switch (action.type) {
      case "emailTyped":
        return { ...state, email: action.email, emailInvalid: false }
      case "emailRefused":
        return { ...state, emailInvalid: true }
      case "codeAsked":
        return { ...state, waiting: "code" }
      case "providerChosen":
        return { ...state, waiting: action.provider }
      case "codeOnItsWay":
        return {
          step: "code",
          email: state.email.trim(),
          code: "",
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
    case "newCodeAsked":
      return { ...state, waiting: "newCode" }
    case "codeOnItsWay":
      return {
        ...state,
        code: "",
        newCodeAt: action.at + secondsBetweenCodes * 1000,
        waiting: null,
      }
    case "otherEmailWanted":
      return {
        step: "start",
        email: state.email,
        emailInvalid: false,
        waiting: null,
      }
    default:
      return state
  }
}
