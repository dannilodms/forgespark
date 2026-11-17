import { useNavigate } from "@tanstack/react-router"
import { LoaderCircleIcon } from "lucide-react"
import { type ReactNode, useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"

import { ErrorAlert } from "../../shared/errors/error-alert"
import {
  type FieldProblems,
  answerCodeOf,
  refusedFields,
} from "../../shared/errors/problems"
import { useLanguage } from "../../shared/locales/i18n"
import {
  type ListItem,
  countryItems,
  currencyItems,
  timeZoneItems,
} from "../../shared/places/places"
import {
  type OpenedSetupCode,
  type TypedSetup,
  useConcludeSetup,
} from "../data/setup"
import type { SetupForm } from "./setup-form"
import { StepHeading } from "./step-heading"

/** A step of the setup whose fields the review shows. */
type TypedStep = "consultancy" | "administrator"

/** The first step, in their order, that holds one of the fields refused. */
function stepHolding(fields: FieldProblems): TypedStep | undefined {
  const names = Object.keys(fields)
  const steps: TypedStep[] = ["consultancy", "administrator"]
  return steps.find((step) => names.some((name) => name.startsWith(`${step}.`)))
}

/** The name of the item of a list a field chose, or the code it holds. */
function nameIn(items: ListItem[], code: string) {
  return items.find((item) => item.value === code)?.label ?? code
}

/**
 * The step of the review: everything typed, by step, each with a way back to
 * its step, the e-mail the first administrator logs in with in plain sight,
 * and concluding. Nothing is saved before the setup is concluded, and
 * concluding sends everything once, with the language of the screen as the
 * preferred language of the administrator.
 *
 * A field the API refuses is shown on its field, on the step that holds it.
 * A setup code that stopped working leads back to the step of the code. A
 * setup someone else concluded leads to the login screen, which says so.
 */
export function ReviewStep({
  form,
  opened,
  onGoTo,
  onCodeStoppedWorking,
}: {
  form: SetupForm
  opened: OpenedSetupCode
  /** Goes back to a step, to change what was typed in it. */
  onGoTo: (step: TypedStep) => void
  onCodeStoppedWorking: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()
  const navigate = useNavigate()
  const conclusion = useConcludeSetup()
  const heading = useRef<HTMLHeadingElement>(null)

  useEffect(() => {
    heading.current?.focus()
  }, [])

  function conclude() {
    conclusion.mutate(
      { typed: form.state.values, code: opened.code, language },
      {
        onSuccess: (concluded) =>
          void navigate({
            to: "/login",
            search: {
              notice:
                concluded === "concluded"
                  ? "setup_concluded"
                  : "setup_already_concluded",
            },
          }),
        onError: (error) => {
          if (answerCodeOf(error) === "setup_code_invalid") {
            onCodeStoppedWorking()
            return
          }
          const fields = refusedFields(error)
          const step = fields ? stepHolding(fields) : undefined
          if (fields && step) {
            // On the fields, where each step shows it when it opens.
            form.setErrorMap({ onSubmit: { fields } })
            onGoTo(step)
          }
        },
      }
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <StepHeading
        ref={heading}
        title={t("setup.review.title")}
        intro={t("setup.review.intro")}
      />

      <form.Subscribe selector={(state) => state.values}>
        {(typed) => (
          <>
            <ConsultancyReview
              typed={typed.consultancy}
              onChange={() => onGoTo("consultancy")}
            />
            <AdministratorReview
              typed={typed.administrator}
              onChange={() => onGoTo("administrator")}
            />
          </>
        )}
      </form.Subscribe>

      {conclusion.error ? <ErrorAlert error={conclusion.error} /> : null}

      <div className="flex gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={() => onGoTo("administrator")}
        >
          {t("setup.back")}
        </Button>
        <Button
          type="button"
          className="flex-1"
          disabled={conclusion.isPending}
          onClick={conclude}
        >
          {conclusion.isPending ? (
            <LoaderCircleIcon
              data-icon="inline-start"
              className="animate-spin"
            />
          ) : null}
          {t("setup.review.conclude")}
        </Button>
      </div>
    </div>
  )
}

/** A part of the review: the title of its step, a way back to it, its rows. */
function ReviewSection({
  id,
  title,
  change,
  onChange,
  note,
  children,
}: {
  id: string
  title: string
  /** The name of the button that leads back to the step. */
  change: string
  onChange: () => void
  /** What the rows are followed by, if anything. */
  note?: ReactNode
  children: ReactNode
}) {
  const { t } = useTranslation("consultancy")

  return (
    <section
      aria-labelledby={id}
      className="flex flex-col gap-3 border-t border-border-subtle pt-4"
    >
      <div className="flex items-center justify-between gap-2">
        <h3 id={id} className="type-label">
          {title}
        </h3>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          aria-label={change}
          onClick={onChange}
        >
          {t("setup.review.change")}
        </Button>
      </div>
      <dl className="grid grid-cols-1 gap-x-4 gap-y-2 sm:grid-cols-[auto_1fr]">
        {children}
      </dl>
      {note}
    </section>
  )
}

/** One row of the review: the label of a field and what was typed in it. */
function Row({ label, value }: { label: string; value: ReactNode }) {
  const { t } = useTranslation("consultancy")
  const given = value !== "" && value !== null

  return (
    <>
      <dt className="type-caption text-muted-foreground sm:pt-0.5">{label}</dt>
      <dd className="min-w-0 break-words">
        {given ? (
          value
        ) : (
          <span className="text-faint-foreground">
            {t("setup.review.notGiven")}
          </span>
        )}
      </dd>
    </>
  )
}

function ConsultancyReview({
  typed,
  onChange,
}: {
  typed: TypedSetup["consultancy"]
  onChange: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()
  const address = [
    typed.address.line1,
    typed.address.line2,
    [typed.address.postal_code, typed.address.city, typed.address.region]
      .filter(Boolean)
      .join(" "),
  ].filter(Boolean)

  return (
    <ReviewSection
      id="review-consultancy"
      title={t("setup.consultancy.title")}
      change={t("setup.review.changeConsultancy")}
      onChange={onChange}
    >
      <Row label={t("setup.consultancy.name")} value={typed.name} />
      <Row label={t("setup.consultancy.legalName")} value={typed.legal_name} />
      <Row
        label={t("setup.consultancy.country")}
        value={nameIn(countryItems(language), typed.country)}
      />
      <Row label={t("setup.consultancy.taxId")} value={typed.tax_id} />
      <Row
        label={t("setup.consultancy.timeZone")}
        value={nameIn(timeZoneItems(language), typed.time_zone)}
      />
      <Row
        label={t("setup.consultancy.currency")}
        value={nameIn(currencyItems(language), typed.currency)}
      />
      <Row label={t("setup.consultancy.email")} value={typed.email} />
      <Row label={t("setup.telephone")} value={typed.telephone} />
      <Row label={t("setup.consultancy.website")} value={typed.website} />
      <Row
        label={t("setup.consultancy.address.legend")}
        value={
          address.length > 0 ? (
            <span className="flex flex-col">
              {address.map((line) => (
                <span key={line}>{line}</span>
              ))}
            </span>
          ) : null
        }
      />
    </ReviewSection>
  )
}

function AdministratorReview({
  typed,
  onChange,
}: {
  typed: TypedSetup["administrator"]
  onChange: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()

  return (
    <ReviewSection
      id="review-administrator"
      title={t("setup.administrator.title")}
      change={t("setup.review.changeAdministrator")}
      onChange={onChange}
      note={
        // Logging in sends a code to this e-mail: a typing mistake here
        // locks the administrator out.
        <p className="rounded-lg border border-border bg-muted px-3 py-2">
          {t("setup.review.logsInWith")}
          <span className="block type-title break-all">{typed.email}</span>
        </p>
      }
    >
      <Row label={t("setup.administrator.name")} value={typed.name} />
      <Row label={t("setup.administrator.email")} value={typed.email} />
      <Row
        label={t("setup.administrator.country")}
        value={nameIn(countryItems(language), typed.country)}
      />
      <Row
        label={t("setup.administrator.nationalId")}
        value={typed.national_id}
      />
      <Row label={t("setup.telephone")} value={typed.telephone} />
    </ReviewSection>
  )
}
