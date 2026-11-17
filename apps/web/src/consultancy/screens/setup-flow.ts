import type { OpenedSetupCode } from "../data/setup"

/** A step after the setup code, whose fields belong to the setup form. */
type FormStep = "consultancy" | "administrator" | "review"

/**
 * The setup is one flow of named steps (State in CODING_STANDARDS.md): the
 * code, the consultancy, the administrator and the review. What is typed
 * lives in the setup form, beside the flow, so it outlives every step: a
 * setup code that stops working leads back to the step of the code, which
 * remembers the step to return to with a new one.
 *
 * To add a step: a member here, the action that reaches it in
 * `setupReducer`, and its branch in `SetupScreen`, which the type check then
 * asks for.
 */
export type SetupFlow =
  | { step: "code"; resumeAt?: FormStep }
  | { step: FormStep; opened: OpenedSetupCode }
  // The registration of a new administrator, ticket 12 of the effort: its
  // steps are the administrator and the review, with no consultancy.
  | { step: "newAdministrator"; opened: OpenedSetupCode }

export type SetupAction =
  /** A setup code opened the setup. */
  | { type: "codeAccepted"; opened: OpenedSetupCode }
  /** The person moves to a step: on, back, or to correct a field. */
  | { type: "wentTo"; step: FormStep }
  /** A request answered that the setup code no longer works. */
  | { type: "codeStoppedWorking" }

/** A screen that has just opened asks for the setup code. */
export const startOfTheSetup: SetupFlow = { step: "code" }

export function setupReducer(flow: SetupFlow, action: SetupAction): SetupFlow {
  switch (action.type) {
    case "codeAccepted":
      if (action.opened.opens === "new_administrator") {
        return { step: "newAdministrator", opened: action.opened }
      }
      return {
        step: (flow.step === "code" && flow.resumeAt) || "consultancy",
        opened: action.opened,
      }
    case "wentTo":
      return flow.step === "code"
        ? flow
        : { step: action.step, opened: flow.opened }
    case "codeStoppedWorking":
      return flow.step === "code" || flow.step === "newAdministrator"
        ? flow
        : { step: "code", resumeAt: flow.step }
    default:
      return flow
  }
}
