import { useForm } from "@tanstack/react-form"
import { LoaderCircleIcon, TriangleAlertIcon } from "lucide-react"
import { useEffect, useRef } from "react"
import { Trans, useTranslation } from "react-i18next"

import { Alert, AlertDescription, AlertTitle } from "#ui/components/alert"
import { Button } from "#ui/components/button"
import { Field, FieldGroup, FieldLabel } from "#ui/components/field"
import { Input } from "#ui/components/input"

import { ErrorAlert } from "../../shared/errors/error-alert"
import { FieldProblemMessage } from "../../shared/errors/field-problem"
import {
  type FieldProblem,
  answerOnAField,
  checkTheContract,
  refusedFields,
} from "../../shared/errors/problems"
import {
  type OpenedSetupCode,
  setupCodeContract,
  useOpenSetup,
} from "../data/setup"

/**
 * The step of the setup code: one field and one primary button. A code that
 * does not work shows on the field and stays as typed. Any other failure
 * shows above the form, with its trace identifier. The code lives in the
 * state of this form, and nowhere else.
 *
 * A setup code that stopped working while the person filled in the setup
 * brings them back here, which says so: what they typed is kept, and a new
 * code returns them to the step they were on.
 */
export function CodeStep({
  stoppedWorking = false,
  onOpened,
}: {
  stoppedWorking?: boolean
  onOpened: (opened: OpenedSetupCode) => void
}) {
  const { t } = useTranslation("consultancy")
  const openSetup = useOpenSetup()
  const input = useRef<HTMLInputElement>(null)

  // Back from a step: the field takes the focus, for the new code.
  useEffect(() => {
    if (stoppedWorking) {
      input.current?.focus()
    }
  }, [stoppedWorking])

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
          // Wrong, expired or ended: the API gives one answer for the three.
          const codeInvalid = answerOnAField(error, "setup_code_invalid")
          if (codeInvalid) {
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

      {stoppedWorking ? (
        <Alert>
          <TriangleAlertIcon className="text-status-attention" />
          <AlertTitle>{t("setup.code.stopped.title")}</AlertTitle>
          <AlertDescription>{t("setup.code.stopped.todo")}</AlertDescription>
        </Alert>
      ) : null}

      <p>
        <Trans
          t={t}
          i18nKey="setup.code.intro"
          components={{
            command: <code className="type-code whitespace-nowrap" />,
          }}
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
