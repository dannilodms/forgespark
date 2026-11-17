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
 * What the setup code opened decides the steps. The first setup has all
 * four. The registration of a new administrator, the way back in, leaves out
 * the step of the consultancy: the code, the administrator and the review.
 *
 * To add a step: a member here, the action that reaches it in
 * `setupReducer`, and its branch in `SetupScreen`, which the type check then
 * asks for.
 */
export type SetupFlow =
  | { step: "code"; resumeAt?: FormStep }
  | { step: FormStep; opened: OpenedSetupCode }

export type SetupAction =
  /** A setup code opened the setup. */
  | { type: "codeAccepted"; opened: OpenedSetupCode }
  /** The person moves to a step: on, back, or to correct a field. */
  | { type: "wentTo"; step: FormStep }
  /** A request answered that the setup code no longer works. */
  | { type: "codeStoppedWorking" }

/** A screen that has just opened asks for the setup code. */
export const startOfTheSetup: SetupFlow = { step: "code" }

/**
 * Where a setup code that opened the setup leads: the step the person was
 * on when the code before it stopped working, or else the first step of what
 * it opens. The way back in has no step of the consultancy.
 */
function stepOpenedBy(opened: OpenedSetupCode, resumeAt?: FormStep): FormStep {
  if (opened.opens === "new_administrator") {
    return resumeAt === "review" ? "review" : "administrator"
  }
  return resumeAt ?? "consultancy"
}

export function setupReducer(flow: SetupFlow, action: SetupAction): SetupFlow {
  switch (action.type) {
    case "codeAccepted":
      return {
        step: stepOpenedBy(
          action.opened,
          flow.step === "code" ? flow.resumeAt : undefined
        ),
        opened: action.opened,
      }
    case "wentTo":
      return flow.step === "code"
        ? flow
        : { step: stepOpenedBy(flow.opened, action.step), opened: flow.opened }
    case "codeStoppedWorking":
      return flow.step === "code" ? flow : { step: "code", resumeAt: flow.step }
    default:
      return flow
  }
}
