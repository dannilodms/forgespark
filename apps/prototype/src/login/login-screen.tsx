import { useNavigate } from "@tanstack/react-router"
import { Button } from "@workspace/ui/components/button"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSeparator,
  InputOTPSlot,
} from "@workspace/ui/components/input-otp"
import { CircleAlertIcon, LoaderCircleIcon } from "lucide-react"
import {
  type Dispatch,
  type FormEvent,
  useEffect,
  useReducer,
  useState,
} from "react"
import { useTranslation } from "react-i18next"

import { askForLoginCode, checkLoginCode, logInThrough } from "./fake-login"
import { LanguageSelect } from "./language-select"
import {
  type LoginAction,
  type LoginProviderId,
  type LoginState,
  initialLoginState,
  isEmail,
  loginCodeLength,
  loginProviders,
  loginReducer,
} from "./login-flow"
import { ProviderMark } from "./provider-marks"

function Waiting() {
  return <LoaderCircleIcon data-icon="inline-start" className="animate-spin" />
}

/** Seconds left until a moment, counted down once a second. */
function useSecondsUntil(moment: number) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])

  return Math.max(0, Math.ceil((moment - now) / 1000))
}

function minutesAndSeconds(seconds: number) {
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`
}

function StartStep({
  state,
  dispatch,
  onLoggedIn,
}: {
  state: Extract<LoginState, { step: "start" }>
  dispatch: Dispatch<LoginAction>
  onLoggedIn: () => void
}) {
  const { t } = useTranslation()
  const waiting = state.waiting !== null

  async function chooseProvider(provider: LoginProviderId) {
    dispatch({ type: "providerChosen", provider })
    const answer = await logInThrough(provider)
    if (answer.loggedIn) {
      onLoggedIn()
    }
  }

  async function askForCode(event: FormEvent) {
    event.preventDefault()
    if (waiting) {
      return
    }
    if (!isEmail(state.email)) {
      dispatch({ type: "emailRefused" })
      return
    }
    dispatch({ type: "codeAsked" })
    await askForLoginCode(state.email)
    dispatch({ type: "codeOnItsWay", at: Date.now() })
  }

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-col gap-2">
        {loginProviders.map((provider) => (
          <Button
            key={provider.id}
            variant="outline"
            className="w-full"
            disabled={waiting}
            onClick={() => void chooseProvider(provider.id)}
          >
            {state.waiting === provider.id ? (
              <Waiting />
            ) : (
              <ProviderMark provider={provider.id} />
            )}
            {t("login.continueWith", { provider: provider.name })}
          </Button>
        ))}
      </div>

      <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
        {t("login.or")}
      </FieldSeparator>

      <form noValidate onSubmit={(event) => void askForCode(event)}>
        <FieldGroup>
          <Field data-invalid={state.emailInvalid || undefined}>
            <FieldLabel htmlFor="login-email">
              {t("login.emailLabel")}
            </FieldLabel>
            <Input
              id="login-email"
              type="email"
              autoComplete="email"
              placeholder={t("login.emailPlaceholder")}
              value={state.email}
              aria-invalid={state.emailInvalid || undefined}
              onChange={(event) =>
                dispatch({ type: "emailTyped", email: event.target.value })
              }
            />
            {state.emailInvalid ? (
              <FieldError className="flex items-center gap-1">
                <CircleAlertIcon className="size-3.5 shrink-0" />
                {t("login.emailInvalid")}
              </FieldError>
            ) : null}
          </Field>
          <Button type="submit" disabled={waiting}>
            {state.waiting === "code" ? <Waiting /> : null}
            {t("login.sendCode")}
          </Button>
        </FieldGroup>
      </form>
    </div>
  )
}

function CodeStep({
  state,
  dispatch,
  onLoggedIn,
}: {
  state: Extract<LoginState, { step: "code" }>
  dispatch: Dispatch<LoginAction>
  onLoggedIn: () => void
}) {
  const { t } = useTranslation()
  const secondsLeft = useSecondsUntil(state.newCodeAt)
  const waiting = state.waiting !== null

  async function logIn(event: FormEvent) {
    event.preventDefault()
    if (waiting || state.code.length < loginCodeLength) {
      return
    }
    dispatch({ type: "loginAsked" })
    const answer = await checkLoginCode(state.code)
    if (answer.loggedIn) {
      onLoggedIn()
    }
  }

  async function askForNewCode() {
    dispatch({ type: "newCodeAsked" })
    await askForLoginCode(state.email)
    dispatch({ type: "codeOnItsWay", at: Date.now() })
  }

  return (
    <form
      className="flex flex-col gap-5"
      onSubmit={(event) => void logIn(event)}
    >
      <p>{t("login.codeSentence", { email: state.email })}</p>

      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="login-code">{t("login.codeLabel")}</FieldLabel>
          <InputOTP
            id="login-code"
            autoFocus
            maxLength={loginCodeLength}
            inputMode="numeric"
            pattern="^\d+$"
            autoComplete="one-time-code"
            value={state.code}
            disabled={waiting}
            onChange={(code) => dispatch({ type: "codeTyped", code })}
          >
            <InputOTPGroup>
              <InputOTPSlot index={0} />
              <InputOTPSlot index={1} />
              <InputOTPSlot index={2} />
            </InputOTPGroup>
            <InputOTPSeparator />
            <InputOTPGroup>
              <InputOTPSlot index={3} />
              <InputOTPSlot index={4} />
              <InputOTPSlot index={5} />
            </InputOTPGroup>
          </InputOTP>
        </Field>
        <Button
          type="submit"
          disabled={waiting || state.code.length < loginCodeLength}
        >
          {state.waiting === "login" ? <Waiting /> : null}
          {t("login.logIn")}
        </Button>
      </FieldGroup>

      <div className="flex flex-col items-start gap-1">
        <Button
          type="button"
          variant="link"
          size="sm"
          className="px-0"
          disabled={waiting || secondsLeft > 0}
          onClick={() => void askForNewCode()}
        >
          {secondsLeft > 0
            ? t("login.newCodeIn", { time: minutesAndSeconds(secondsLeft) })
            : t("login.newCode")}
        </Button>
        <Button
          type="button"
          variant="link"
          size="sm"
          className="px-0"
          disabled={waiting}
          onClick={() => dispatch({ type: "otherEmailWanted" })}
        >
          {t("login.otherEmail")}
        </Button>
      </div>
    </form>
  )
}

/**
 * The login screen, framed like the application: a title bar, the framed page
 * with one pane in its centre, and a status bar that holds the language
 * selector. The flow is one reducer whose state is a union of named steps.
 */
export function LoginScreen() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [state, dispatch] = useReducer(loginReducer, initialLoginState)

  const openTheProduct = () => void navigate({ to: "/work" })

  return (
    <div className="flex min-h-full flex-col">
      <header className="flex h-7 shrink-0 items-center border-b bg-chrome px-3 type-caption text-muted-foreground">
        {t("login.productName")}
      </header>

      <main className="flex flex-1 items-center justify-center page-framing p-4">
        <div className="relative w-full max-w-sm rounded-xl border bg-card p-6">
          <span className="absolute top-0 left-0 -translate-x-1/2 -translate-y-1/2 frame-node" />
          <span className="absolute top-0 right-0 translate-x-1/2 -translate-y-1/2 frame-node" />
          <span className="absolute bottom-0 left-0 -translate-x-1/2 translate-y-1/2 frame-node" />
          <span className="absolute right-0 bottom-0 translate-x-1/2 translate-y-1/2 frame-node" />

          <div className="flex flex-col gap-6">
            <div className="flex flex-col gap-1">
              <h1 className="type-display">{t("login.heading")}</h1>
              <p className="text-muted-foreground">{t("login.intro")}</p>
            </div>

            {state.step === "start" ? (
              <StartStep
                state={state}
                dispatch={dispatch}
                onLoggedIn={openTheProduct}
              />
            ) : (
              <CodeStep
                state={state}
                dispatch={dispatch}
                onLoggedIn={openTheProduct}
              />
            )}

            <p className="type-caption text-muted-foreground">
              {t("login.footer")}
            </p>
          </div>
        </div>
      </main>

      <footer className="flex h-7 shrink-0 items-center justify-end border-t bg-chrome px-2">
        <LanguageSelect />
      </footer>
    </div>
  )
}
