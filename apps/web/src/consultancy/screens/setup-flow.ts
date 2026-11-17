import type { FieldProblems } from "../../shared/errors/problems"
import { type TypedStep, holds, problemsOf } from "../components/setup-form"
import type { OpenedSetupCode } from "../data/setup"

/** A step after the setup code, whose fields belong to the setup form. */
type FormStep = TypedStep | "review"

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
 * The fields the API refused are held here until the step that holds each
 * one opens and shows it (`useStepOpening`): a field forgets its problems
 * when it unmounts, so they cannot wait in the form.
 *
 * To add a step: a member here, the action that reaches it in
 * `setupReducer`, and its branch in `SetupScreen`, which the type check then
 * asks for.
 */
export type SetupFlow =
  | { step: "code"; resumeAt?: FormStep }
  | { step: FormStep; opened: OpenedSetupCode; refused?: FieldProblems }

export type SetupAction =
  /** A setup code opened the setup. */
  | { type: "codeAccepted"; opened: OpenedSetupCode }
  /** The person moves to a step: on, back, or to correct a field. */
  | { type: "wentTo"; step: FormStep }
  /** A request answered that the setup code no longer works. */
  | { type: "codeStoppedWorking" }
  /** The API refused fields: the first step that holds one opens. */
  | { type: "fieldsRefused"; fields: FieldProblems }
  /** A step that opened shows the refused fields it holds. */
  | { type: "problemsShown"; step: TypedStep }

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

/** The first step, in their order, that holds one of the fields refused. */
function stepHolding(fields: FieldProblems, opened: OpenedSetupCode) {
  const steps: TypedStep[] = ["consultancy", "administrator"]
  return steps
    .filter((step) => stepOpenedBy(opened, step) === step)
    .find((step) => problemsOf(fields, step) !== undefined)
}

export function setupReducer(flow: SetupFlow, action: SetupAction): SetupFlow {
  if (flow.step === "code") {
    return action.type === "codeAccepted"
      ? {
          step: stepOpenedBy(action.opened, flow.resumeAt),
          opened: action.opened,
        }
      : flow
  }
  switch (action.type) {
    case "wentTo":
      return { ...flow, step: stepOpenedBy(flow.opened, action.step) }
    case "codeStoppedWorking":
      return { step: "code", resumeAt: flow.step }
    case "fieldsRefused": {
      const step = stepHolding(action.fields, flow.opened)
      return step ? { ...flow, step, refused: action.fields } : flow
    }
    case "problemsShown":
      return {
        ...flow,
        refused: Object.fromEntries(
          Object.entries(flow.refused ?? {}).filter(
            ([field]) => !holds(action.step, field)
          )
        ),
      }
    default:
      return flow
  }
}
