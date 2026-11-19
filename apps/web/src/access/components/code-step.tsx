import { useForm } from "@tanstack/react-form"
import { LoaderCircleIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { Trans, useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"
import { Field, FieldGroup, FieldLabel } from "#ui/components/field"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "#ui/components/input-otp"

import { StepActions } from "../../shared/components/step-actions"
import { ErrorAlert } from "../../shared/errors/error-alert"
import { FieldProblemMessage } from "../../shared/errors/field-problem"
import {
  type FieldProblem,
  answerCodeOf,
  answerOnAField,
  checkTheContract,
  refusedFields,
} from "../../shared/errors/problems"
import {
  loginCodeContract,
  useAskForLoginCode,
  useCheckLoginCode,
} from "../data/session"

/**
 * How long after asking for a code the person may ask for a new one: the
 * API takes one a minute for each e-mail, and the screen counts the minute.
 */
const newCodeAfter = 60_000

/** The six boxes of the code, one digit each. */
const digits = [0, 1, 2, 3, 4, 5]

/**
 * The step of the code: the six digits and "log in". It says a code was
 * sent only if the e-mail belongs to a user, since the API answers the same
 * for any e-mail. A wrong code shows under the field, with the attempts
 * left, and the step stays; the third, which the API answers as the code
 * ended, returns to the e-mail step. A new code may be asked for a minute
 * after the last one, and ends the one before. The code lives in the state
 * of this form, and nowhere else.
 */
export function CodeStep({
  email,
  askedAt,
  onNewCode,
  onEnded,
  onBack,
  onLoggedIn,
}: {
  email: string
  /** When the API took the last ask, by the clock of the browser. */
  askedAt: number
  /** The API took the ask of a new code. */
  onNewCode: () => void
  /** The API answered that the code ended. */
  onEnded: () => void
  /** The person goes back to type another e-mail. */
  onBack: () => void
  /** A right code logged the user in. */
  onLoggedIn: () => void
}) {
  const { t } = useTranslation("access")
  const checkLoginCode = useCheckLoginCode()
  const askForLoginCode = useAskForLoginCode()
  const input = useRef<HTMLInputElement>(null)

  // The field takes the focus when the step opens and after a new code.
  useEffect(() => {
    input.current?.focus()
  }, [askedAt])

  const form = useForm({
    defaultValues: { code: "" },
    validators: {
      // What the contract declares, before anything is sent.
      onSubmit: checkTheContract(loginCodeContract),
      // The answer of the API: on the field, the step of the e-mail, or
      // above the form.
      onSubmitAsync: async ({ value }) => {
        try {
          await checkLoginCode.mutateAsync(value.code)
          onLoggedIn()
          return undefined
        } catch (error) {
          if (answerCodeOf(error) === "login_code_ended") {
            onEnded()
            return undefined
          }
          // Wrong, expired, ended by a new one or asked in another browser:
          // the API gives one answer for all, with the attempts left.
          const codeInvalid = answerOnAField(error, "login_code_invalid")
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

  function askForANewCode() {
    askForLoginCode.mutate(email, {
      onSuccess: () => {
        form.reset()
        onNewCode()
      },
    })
  }

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
      {askForLoginCode.error ? (
        <ErrorAlert error={askForLoginCode.error} />
      ) : null}

      <p>
        <Trans
          t={t}
          i18nKey="login.code.intro"
          values={{ email }}
          components={{ email: <strong className="break-all" /> }}
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
                <FieldLabel htmlFor="login-code">
                  {t("login.code.label")}
                </FieldLabel>
                <InputOTP
                  ref={input}
                  id="login-code"
                  maxLength={digits.length}
                  pattern="^[0-9]+$"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  value={field.state.value}
                  aria-invalid={problem ? true : undefined}
                  aria-describedby={problem ? "login-code-problem" : undefined}
                  // No `onBlur`: TanStack Form drops the answer of the API
                  // when the field is left, and the line under it with it,
                  // which moves the buttons under the pointer as it presses
                  // one. The answer stays until another code is typed.
                  onChange={(value) => field.handleChange(value)}
                >
                  <InputOTPGroup className="type-code">
                    {digits.map((index) => (
                      <InputOTPSlot
                        key={index}
                        index={index}
                        aria-invalid={problem ? true : undefined}
                      />
                    ))}
                  </InputOTPGroup>
                </InputOTP>
                {problem ? (
                  <FieldProblemMessage
                    id="login-code-problem"
                    problem={problem}
                  />
                ) : null}
              </Field>
            )
          }}
        </form.Field>
      </FieldGroup>

      <NewCode
        askedAt={askedAt}
        asking={askForLoginCode.isPending}
        onAsk={askForANewCode}
      />

      <StepActions
        back={
          <Button type="button" variant="outline" onClick={onBack}>
            {t("login.code.back")}
          </Button>
        }
      >
        <form.Subscribe selector={(state) => state.isSubmitting}>
          {(submitting) => (
            <Button type="submit" disabled={submitting}>
              {submitting ? (
                <LoaderCircleIcon
                  data-icon="inline-start"
                  className="animate-spin"
                />
              ) : null}
              {t("login.code.submit")}
            </Button>
          )}
        </form.Subscribe>
      </StepActions>
    </form>
  )
}

/**
 * The way to a new code: closed for a minute after the last ask, with the
 * seconds left counted under it, then open.
 */
function NewCode({
  askedAt,
  asking,
  onAsk,
}: {
  askedAt: number
  asking: boolean
  onAsk: () => void
}) {
  const { t } = useTranslation("access")
  const wait = useSecondsUntil(askedAt + newCodeAfter)

  return (
    <div className="flex flex-col items-start gap-1">
      <Button
        type="button"
        variant="link"
        className="h-auto px-0"
        disabled={wait > 0 || asking}
        aria-describedby={wait > 0 ? "login-new-code-wait" : undefined}
        onClick={onAsk}
      >
        {asking ? (
          <LoaderCircleIcon data-icon="inline-start" className="animate-spin" />
        ) : null}
        {t("login.code.newCode")}
      </Button>
      {wait > 0 ? (
        <p
          id="login-new-code-wait"
          className="type-caption text-muted-foreground"
        >
          {t("login.code.newCodeIn", { wait })}
        </p>
      ) : null}
    </div>
  )
}

/**
 * The whole seconds from now until a moment, by the clock of the browser,
 * counted down once a second; none once it has passed. It never counts more
 * than the wait for a new code, should the clock be read late.
 */
function useSecondsUntil(moment: number) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => {
      const at = Date.now()
      setNow(at)
      if (at >= moment) {
        clearInterval(timer)
      }
    }, 1000)
    return () => clearInterval(timer)
  }, [moment])

  const left = Math.ceil((moment - now) / 1000)
  return Math.min(Math.max(left, 0), newCodeAfter / 1000)
}
