import { useForm } from "@tanstack/react-form"
import { LoaderCircleIcon } from "lucide-react"
import { useRef } from "react"
import { Trans, useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"
import { Field, FieldGroup, FieldLabel } from "#ui/components/field"
import { Input } from "#ui/components/input"

import { ErrorAlert } from "../../shared/errors/error-alert"
import { FieldProblemMessage } from "../../shared/errors/field-problem"
import {
  type FieldProblem,
  answerCodeOf,
  checkTheContract,
  refusedFields,
} from "../../shared/errors/problems"
import {
  type OpenedSetupCode,
  setupCodeContract,
  useOpenSetup,
} from "../data/setup"

/** Wrong, expired or ended: the API gives one answer for the three. */
const codeInvalid: FieldProblem = { code: "setup_code_invalid", params: {} }

/**
 * The step of the setup code: one field and one primary button. A code that
 * does not work shows on the field and stays as typed. Any other failure
 * shows above the form, with its trace identifier. The code lives in the
 * state of this form, and nowhere else.
 */
export function CodeStep({
  onOpened,
}: {
  onOpened: (opened: OpenedSetupCode) => void
}) {
  const { t } = useTranslation("consultancy")
  const openSetup = useOpenSetup()
  const input = useRef<HTMLInputElement>(null)

  const form = useForm({
    defaultValues: { code: "" },
    validators: {
      // What the contract declares, before anything is sent.
      onSubmit: checkTheContract(setupCodeContract),
      // The answer of the API: on the field it is about, or above the form.
      // `fields` is what makes TanStack Form read `form` as the error of the
      // form, so it is always there.
      onSubmitAsync: async ({ value }) => {
        try {
          onOpened(await openSetup.mutateAsync(value.code))
          return undefined
        } catch (error) {
          if (answerCodeOf(error) === "setup_code_invalid") {
            return { fields: { code: codeInvalid } }
          }
          const fields = refusedFields(error)
          return fields ? { fields } : { form: error, fields: {} }
        }
      },
    },
    // The field takes the focus again, to correct it or to try again.
    onSubmitInvalid: () => input.current?.focus(),
  })

  return (
    <form
      noValidate
      className="flex flex-col gap-5"
      onSubmit={(event) => {
        event.preventDefault()
        void form.handleSubmit()
      }}
    >
      <form.Subscribe selector={(state) => state.errorMap.onSubmit}>
        {(failure) => (failure ? <ErrorAlert error={failure} /> : null)}
      </form.Subscribe>

      <p>
        <Trans
          t={t}
          i18nKey="setup.code.intro"
          components={{ command: <code className="type-code" /> }}
        />
      </p>

      <FieldGroup>
        <form.Field name="code">
          {(field) => {
            const problem = field.state.meta.errors[0] as
              | FieldProblem
              | undefined
            return (
              <Field data-invalid={problem ? true : undefined}>
                <FieldLabel htmlFor="setup-code">
                  {t("setup.code.label")}
                </FieldLabel>
                <Input
                  ref={input}
                  id="setup-code"
                  className="type-code"
                  // Nothing typed here is kept by the browser.
                  autoComplete="off"
                  autoCapitalize="characters"
                  spellCheck={false}
                  value={field.state.value}
                  aria-invalid={problem ? true : undefined}
                  aria-describedby={problem ? "setup-code-problem" : undefined}
                  onChange={(event) => field.handleChange(event.target.value)}
                  onBlur={field.handleBlur}
                />
                {problem ? (
                  <FieldProblemMessage
                    id="setup-code-problem"
                    problem={problem}
                  />
                ) : null}
              </Field>
            )
          }}
        </form.Field>

        <form.Subscribe selector={(state) => state.isSubmitting}>
          {(submitting) => (
            <Button type="submit" disabled={submitting}>
              {submitting ? (
                <LoaderCircleIcon
                  data-icon="inline-start"
                  className="animate-spin"
                />
              ) : null}
              {t("setup.code.submit")}
            </Button>
          )}
        </form.Subscribe>
      </FieldGroup>
    </form>
  )
}
