import { useForm } from "@tanstack/react-form"
import { useState } from "react"

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
