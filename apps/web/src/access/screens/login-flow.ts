/**
 * The login with a login code is one flow of named steps (State in
 * CODING_STANDARDS.md): the e-mail, the sending of the code, and the code.
 * Nothing of it outlives the screen: a page loaded again starts at the
 * e-mail, with nothing typed.
 *
 * - `email`: the person types their e-mail. It holds the e-mail typed
 *   before, when they come back to it, and whether they come back because
 *   the code ended.
 * - `sending`: the API is asked for a code for that e-mail.
 * - `code`: the person types the six digits. It holds the e-mail the code
 *   was asked for, which the step shows, and when the API answered, from
 *   which the step counts the wait for a new code. The attempts left are
 *   not held here: each wrong code is answered with them, and the step shows
 *   that answer under the field.
 *
 * To add a step: a member here, the actions that reach it in `loginReducer`,
 * and its branch in `LoginScreen`, which the type check then asks for.
 */
export type LoginFlow =
  | { step: "email"; email: string; codeEnded?: boolean }
  | { step: "sending"; email: string }
  | { step: "code"; email: string; askedAt: number }

export type LoginAction =
  /** The e-mail step asks for a code. */
  | { type: "codeAskedFor"; email: string }
  /** The API took the ask, at a moment of the clock of the browser. */
  | { type: "codeSent"; at: number }
  /** The API refused the ask: the e-mail step shows why. */
  | { type: "askRefused" }
  /** The API answered that the code ended: a new one has to be asked for. */
  | { type: "codeEnded" }
  /** The person goes back to type another e-mail. */
  | { type: "wentBack" }

/** A screen that has just opened asks for the e-mail. */
export const startOfTheLogin: LoginFlow = { step: "email", email: "" }

export function loginReducer(flow: LoginFlow, action: LoginAction): LoginFlow {
  switch (flow.step) {
    case "email":
      return action.type === "codeAskedFor"
        ? { step: "sending", email: action.email }
        : flow
    case "sending":
      if (action.type === "codeSent") {
        return { step: "code", email: flow.email, askedAt: action.at }
      }
      return action.type === "askRefused"
        ? { step: "email", email: flow.email }
        : flow
    case "code":
      return codeReducer(flow, action)
  }
}

/** The step of the code: a new code, the code ended, or back to the e-mail. */
function codeReducer(
  flow: Extract<LoginFlow, { step: "code" }>,
  action: LoginAction
): LoginFlow {
  switch (action.type) {
    case "codeSent":
      return { ...flow, askedAt: action.at }
    case "codeEnded":
      return { step: "email", email: flow.email, codeEnded: true }
    case "wentBack":
      return { step: "email", email: flow.email }
    default:
      return flow
  }
}
