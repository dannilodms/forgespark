import type { OpenedSetupCode } from "../data/setup"

/**
 * The setup is one flow of named steps (State in CODING_STANDARDS.md): the
 * code, the consultancy, the administrator and the review. A code that opens
 * the registration of a new administrator leaves the consultancy out.
 *
 * To add a step: a member here, the action that reaches it in
 * `setupReducer`, and its branch in `SetupScreen`, which the type check then
 * asks for.
 */
export type SetupFlow =
  | { step: "code" }
  | { step: "consultancy"; opened: OpenedSetupCode }
  | { step: "administrator"; opened: OpenedSetupCode }

export type SetupAction = { type: "codeAccepted"; opened: OpenedSetupCode }

/** A screen that has just opened asks for the setup code. */
export const startOfTheSetup: SetupFlow = { step: "code" }

export function setupReducer(flow: SetupFlow, action: SetupAction): SetupFlow {
  switch (action.type) {
    case "codeAccepted":
      return action.opened.opens === "first_setup"
        ? { step: "consultancy", opened: action.opened }
        : { step: "administrator", opened: action.opened }
    default:
      return flow
  }
}
