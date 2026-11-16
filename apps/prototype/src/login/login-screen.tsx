import { useNavigate } from "@tanstack/react-router"
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@workspace/ui/components/alert"
import { Badge } from "@workspace/ui/components/badge"
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
import {
  CircleAlertIcon,
  CircleXIcon,
  InfoIcon,
  LoaderCircleIcon,
} from "lucide-react"
import {
  type Dispatch,
  type FormEvent,
  useEffect,
  useReducer,
  useRef,
  useState,
} from "react"
import { useTranslation } from "react-i18next"

import {
  askForLoginCode,
  checkLoginCode,
  logInThrough,
  useLoginOpening,
} from "./fake-login"
import { LanguageSelect } from "./language-select"
import {
  type CodeProblem,
  type LoginAction,
  type LoginAlert,
  type LoginOpening,
  type LoginProviderId,
  type LoginState,
  codeIsOver,
  isEmail,
  loginCodeLength,
  loginReducer,
  openLogin,
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

/**
 * What an answer reported, above the form: what happened, then what to do.
 * A problem is red and carries a cross. The end of a session is no problem,
 * and carries the sign of a notice. Neither depends on its colour.
 */
function AlertAboveTheForm({ alert }: { alert: LoginAlert }) {
  const { t } = useTranslation()

  switch (alert.kind) {
    case "sessionEnded":
      return (
        <Alert>
          <InfoIcon />
          <AlertTitle>{t("login.sessionEnded.title")}</AlertTitle>
          <AlertDescription>{t("login.sessionEnded.todo")}</AlertDescription>
        </Alert>
      )
    case "limitReached":
      return (
        <Alert variant="destructive">
          <CircleXIcon />
          <AlertTitle>{t("login.limitReached.title")}</AlertTitle>
          <AlertDescription>{t("login.limitReached.todo")}</AlertDescription>
        </Alert>
      )
    case "accountNotRegistered":
      return (
        <Alert variant="destructive">
          <CircleXIcon />
          <AlertTitle>{t("login.accountNotRegistered.title")}</AlertTitle>
          <AlertDescription>
            {t("login.accountNotRegistered.todo")}
          </AlertDescription>
        </Alert>
      )
    case "providerFailed":
      return (
        <Alert variant="destructive">
          <CircleXIcon />
          <AlertTitle>
            {t("login.providerFailed.title", {
              provider: t(`login.provider.${alert.provider}`),
            })}
          </AlertTitle>
          <AlertDescription>{t("login.providerFailed.todo")}</AlertDescription>
        </Alert>
      )
  }
}

type StartState = Extract<LoginState, { step: "start" }>

function StartStep({
  state,
  opening,
  dispatch,
  onLoggedIn,
}: {
  state: StartState
  opening: LoginOpening
  dispatch: Dispatch<LoginAction>
  onLoggedIn: () => void
}) {
  const { t } = useTranslation()
  const waiting = state.waiting !== null
  const hasProviders = opening.providers.length > 0

  async function chooseProvider(provider: LoginProviderId) {
    dispatch({ type: "providerChosen", provider })
    const answer = await logInThrough(provider)
    if (answer.outcome === "loggedIn") {
      onLoggedIn()
    } else {
      dispatch({ type: "providerRefused", alert: answer.alert })
    }
  }

  return (
    <div className="flex flex-col gap-5">
      {state.alert ? <AlertAboveTheForm alert={state.alert} /> : null}

      {hasProviders ? (
        <div className="flex flex-col gap-2">
          {opening.providers.map((provider) => (
            <Button
              key={provider.id}
              variant="outline"
              className="h-auto min-h-8 w-full flex-wrap py-1 whitespace-normal"
              disabled={waiting}
              onClick={() => void chooseProvider(provider.id)}
            >
              {state.waiting === provider.id ? (
                <Waiting />
              ) : (
                <ProviderMark provider={provider.id} />
              )}
              {t("login.continueWith", {
                provider: t(`login.provider.${provider.id}`),
              })}
              {provider.forDevelopment ? (
                <Badge variant="secondary">{t("login.forDevelopment")}</Badge>
              ) : null}
            </Button>
          ))}
        </div>
      ) : null}

      {hasProviders && opening.loginCode ? (
        <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
          {t("login.or")}
        </FieldSeparator>
      ) : null}

      {opening.loginCode ? (
        <EmailForm state={state} dispatch={dispatch} />
      ) : null}
    </div>
  )
}

function EmailForm({
  state,
  dispatch,
}: {
  state: StartState
  dispatch: Dispatch<LoginAction>
}) {
  const { t } = useTranslation()
  const waiting = state.waiting !== null

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
    const answer = await askForLoginCode(state.email)
    dispatch(
      answer.outcome === "onItsWay"
        ? { type: "codeOnItsWay", at: Date.now() }
        : { type: "limitReached" }
    )
  }

  return (
    <form noValidate onSubmit={(event) => void askForCode(event)}>
      <FieldGroup>
        <Field data-invalid={state.emailInvalid || undefined}>
          <FieldLabel htmlFor="login-email">{t("login.emailLabel")}</FieldLabel>
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
  )
}

/** Why the login code was refused, as the error of its field. */
function CodeError({ problem }: { problem: CodeProblem }) {
  const { t } = useTranslation()

  return (
    <FieldError className="flex items-start gap-1">
      <CircleAlertIcon className="mt-0.5 size-3.5 shrink-0" />
      {problem.kind === "wrong"
        ? t("login.codeWrong", { count: problem.attemptsLeft })
        : problem.kind === "ended"
          ? t("login.codeEnded")
          : t("login.codeExpired")}
    </FieldError>
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
  const boxes = useRef<HTMLInputElement>(null)
  const boxesOff = waiting || codeIsOver(state)
  const codeRefused = state.codeProblem !== null || undefined

  // Focus goes to the boxes when the step opens and each time they can be
  // typed in again: after a wrong code, and after a new one.
  useEffect(() => {
    if (!boxesOff) {
      boxes.current?.focus()
    }
  }, [boxesOff])

  async function logIn(event: FormEvent) {
    event.preventDefault()
    if (boxesOff || state.code.length < loginCodeLength) {
      return
    }
    dispatch({ type: "loginAsked" })
    const answer = await checkLoginCode(state.code)
    if (answer.outcome === "loggedIn") {
      onLoggedIn()
    } else {
      dispatch({ type: "codeRefused", problem: answer.problem })
    }
  }

  async function askForNewCode() {
    dispatch({ type: "newCodeAsked" })
    const answer = await askForLoginCode(state.email)
    dispatch(
      answer.outcome === "onItsWay"
        ? { type: "codeOnItsWay", at: Date.now() }
        : { type: "limitReached" }
    )
  }

  return (
    <form
      className="flex flex-col gap-5"
      onSubmit={(event) => void logIn(event)}
    >
      {state.alert ? <AlertAboveTheForm alert={state.alert} /> : null}

      <p>{t("login.codeSentence", { email: state.email })}</p>

      <FieldGroup>
        <Field data-invalid={codeRefused}>
          <FieldLabel htmlFor="login-code">{t("login.codeLabel")}</FieldLabel>
          <InputOTP
            ref={boxes}
            id="login-code"
            maxLength={loginCodeLength}
            inputMode="numeric"
            pattern="^\d+$"
            autoComplete="one-time-code"
            value={state.code}
            disabled={boxesOff}
            aria-invalid={codeRefused}
            onChange={(code) => dispatch({ type: "codeTyped", code })}
          >
            <InputOTPGroup>
              <InputOTPSlot index={0} aria-invalid={codeRefused} />
              <InputOTPSlot index={1} aria-invalid={codeRefused} />
              <InputOTPSlot index={2} aria-invalid={codeRefused} />
            </InputOTPGroup>
            <InputOTPSeparator />
            <InputOTPGroup>
              <InputOTPSlot index={3} aria-invalid={codeRefused} />
              <InputOTPSlot index={4} aria-invalid={codeRefused} />
              <InputOTPSlot index={5} aria-invalid={codeRefused} />
            </InputOTPGroup>
          </InputOTP>
          {state.codeProblem ? <CodeError problem={state.codeProblem} /> : null}
        </Field>
        <Button
          type="submit"
          disabled={boxesOff || state.code.length < loginCodeLength}
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
  const opening = useLoginOpening()
  const [state, dispatch] = useReducer(loginReducer, opening, openLogin)

  // What the screen reads can change while it is open, and the flow moves to
  // a state it still allows. In the prototype that happens when a scenario is
  // chosen in the prototype bar.
  useEffect(() => {
    dispatch({ type: "reopened", opening })
  }, [opening])

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
                opening={opening}
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
