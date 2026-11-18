import { useForm } from "@tanstack/react-form"
import { type RefObject, useEffect, useEffectEvent, useState } from "react"

import { focusTheFirstProblem } from "../../shared/components/form-fields"
import type { FieldProblems } from "../../shared/errors/problems"
import { currencyOf, timeZoneOfTheBrowser } from "../../shared/places/places"
import { blankSetup } from "../data/setup"

/**
 * The form of the setup: everything typed after the setup code, in one
 * TanStack Form that outlives the steps, so that moving between them, or
 * back to the step of the code, loses nothing. Each step is a group of its
 * fields (`form.FormGroup`), which checks only those fields when the person
 * moves on. The time zone starts at the browser's, when the API accepts it.
 */
export function useSetupForm() {
  const [defaultValues] = useState(() => blankSetup(timeZoneOfTheBrowser()))
  return useForm({ defaultValues })
}

export type SetupForm = ReturnType<typeof useSetupForm>

/** The fields of the form that have a problem now, by their names. */
export function fieldsWithProblems(form: SetupForm) {
  return Object.entries(form.state.fieldMeta)
    .filter(([, meta]) => (meta?.errors.length ?? 0) > 0)
    .map(([name]) => name)
}

/**
 * Suggests the currency of the country chosen for the consultancy, unless the
 * person chose a currency by hand. The suggestion does not count as a change
 * by hand, so the next country suggests again.
 */
export function suggestTheCurrency(form: SetupForm, country: string) {
  if (form.getFieldMeta("consultancy.currency")?.isDirty) {
    return
  }
  form.setFieldValue("consultancy.currency", currencyOf(country), {
    dontUpdateMeta: true,
  })
}

/** A step whose fields are typed: a group of the form, by its name. */
export type TypedStep = "consultancy" | "administrator"

/** Whether a field, by its place in the body, belongs to a step. */
export function holds(step: TypedStep, field: string) {
  return field.startsWith(`${step}.`)
}

/** The refused fields a step holds, or nothing when it holds none. */
export function problemsOf(
  refused: FieldProblems | undefined,
  step: TypedStep
): FieldProblems | undefined {
  const own = Object.entries(refused ?? {}).filter(([field]) =>
    holds(step, field)
  )
  return own.length > 0 ? Object.fromEntries(own) : undefined
}

/**
 * When a step opens, it shows the problems the API named on its fields, and
 * the focus goes to the first of them, or else to the title of the step.
 *
 * The problems are put on the form here, once the fields of the step are
 * mounted, and not before: a field of TanStack Form forgets its problems
 * when it unmounts, and React in development mounts each new component,
 * unmounts it and mounts it again. This effect runs again after that, with
 * the same problems.
 */
export function useStepOpening({
  form,
  step,
  refused,
  onProblemsShown,
  element,
  heading,
}: {
  form: SetupForm
  step: TypedStep
  /** The fields the API refused, of every step. */
  refused: FieldProblems | undefined
  /** The step showed the refused fields it holds. */
  onProblemsShown: () => void
  element: RefObject<HTMLElement | null>
  heading: RefObject<HTMLElement | null>
}) {
  const open = useEffectEvent(() => {
    const own = problemsOf(refused, step)
    if (own) {
      form.setErrorMap({ onSubmit: { fields: own } })
      onProblemsShown()
    }
    if (!focusTheFirstProblem(element.current, fieldsWithProblems(form))) {
      heading.current?.focus()
    }
  })

  // When the step opens, and not again: the flow forgets the problems
  // shown, and a field changed afterwards keeps no problem.
  useEffect(() => {
    open()
  }, [])
}
