import { useForm } from "@tanstack/react-form"
import { LoaderCircleIcon, TriangleAlertIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { Alert, AlertDescription, AlertTitle } from "#ui/components/alert"
import { Button } from "#ui/components/button"
import { FieldGroup } from "#ui/components/field"

import {
  TextField,
  focusTheFirstProblem,
} from "../../shared/components/form-fields"
import { StepActions } from "../../shared/components/step-actions"
import { ErrorAlert } from "../../shared/errors/error-alert"
import { checkTheContract, refusedFields } from "../../shared/errors/problems"
import { emailContract, useAskForLoginCode } from "../data/session"

/**
 * The step of the e-mail: one field and "send a code". It looks and answers
 * the same whether or not the e-mail belongs to a user. A field the contract
 * or the API refuses shows on the field; any other failure, the limit of
 * codes for the address and the limit of Caddy among them, shows above the
 * form, with its trace identifier. While the code is asked for (the step
 * `sending`), the field and the button wait.
 *
 * Back from the step of the code, it holds the e-mail typed before, and says
 * so when the code ended.
 */
export function EmailStep({
  email,
  sending,
  codeEnded = false,
  onSending,
  onSent,
  onRefused,
}: {
  /** The e-mail typed before, back from the step of the code, or none. */
  email: string
  /** Whether a code is being asked for. */
  sending: boolean
  /** Whether the person is back because the code ended. */
  codeEnded?: boolean
  /** The step asks the API for a code for this e-mail. */
  onSending: (email: string) => void
  /** The API took the ask. */
  onSent: () => void
  /** The API refused the ask: the step shows why. */
  onRefused: () => void
}) {
  const { t } = useTranslation("access")
  const askForLoginCode = useAskForLoginCode()
  const container = useRef<HTMLFormElement>(null)
  const [cameBack] = useState(email !== "")

  // Back from the step of the code: the field takes the focus.
  useEffect(() => {
    if (cameBack) {
      focusTheFirstProblem(container.current, ["email"])
    }
  }, [cameBack])

  const form = useForm({
    defaultValues: { email },
    validators: {
      // What the contract declares, before anything is sent.
      onSubmit: checkTheContract(emailContract),
      // The answer of the API: on the field it is about, or above the form.
      // `fields` is what makes TanStack Form read `form` as the error of the
      // form, so it is always there.
      onSubmitAsync: async ({ value }) => {
        onSending(value.email)
        try {
          await askForLoginCode.mutateAsync(value.email)
          onSent()
          return undefined
        } catch (error) {
          onRefused()
          const fields = refusedFields(error)
          return fields ? { fields } : { form: error, fields: {} }
        }
      },
    },
    // The field takes the focus again, to correct it or to try again.
    onSubmitInvalid: () => focusTheFirstProblem(container.current, ["email"]),
  })

  return (
    <form
      ref={container}
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

      {codeEnded ? (
        <Alert>
          <TriangleAlertIcon className="text-status-attention" />
          <AlertTitle>{t("login.email.codeEnded.title")}</AlertTitle>
          <AlertDescription>{t("login.email.codeEnded.todo")}</AlertDescription>
        </Alert>
      ) : null}

      <p>{t("login.email.intro")}</p>

      <FieldGroup>
        <form.Field name="email">
          {(field) => (
            <TextField
              field={field}
              label={t("login.email.label")}
              type="email"
              inputMode="email"
              autoComplete="email"
              spellCheck={false}
            />
          )}
        </form.Field>
      </FieldGroup>

      <StepActions>
        <Button type="submit" disabled={sending}>
          {sending ? (
            <LoaderCircleIcon
              data-icon="inline-start"
              className="animate-spin"
            />
          ) : null}
          {t("login.email.submit")}
        </Button>
      </StepActions>
    </form>
  )
}
