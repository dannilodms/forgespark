import { useNavigate } from "@tanstack/react-router"
import { CircleDashedIcon, CircleXIcon, LoaderCircleIcon } from "lucide-react"
import { type ReactNode, useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"

import { ErrorAlert } from "../../shared/errors/error-alert"
import {
  type FieldProblem,
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
  useRegisterAdministrator,
} from "../data/setup"
import { ImageStatus, ImageThumbnail } from "./image-field"
import type { SetupForm } from "./setup-form"
import type { SetupImage, SetupImages } from "./setup-image"
import { StepHeading } from "./step-heading"

/** A step of the setup whose fields the review shows. */
type TypedStep = "consultancy" | "administrator"

/**
 * An e-mail that belongs to a user: the API answers it as a conflict, about
 * the e-mail of the administrator, which is where it shows.
 */
const emailInUse: FieldProblem = { code: "email_in_use", params: {} }

/** The fields a refusal of the API is about, by their names in the form. */
function fieldsRefusedBy(error: unknown): FieldProblems | undefined {
  if (answerCodeOf(error) === "email_in_use") {
    return { "administrator.email": emailInUse }
  }
  return refusedFields(error)
}

/** The first step, in their order, that holds one of the fields refused. */
function stepHolding(fields: FieldProblems): TypedStep | undefined {
  const names = Object.keys(fields)
  const steps: TypedStep[] = ["consultancy", "administrator"]
  return steps.find((step) => names.some((name) => name.startsWith(`${step}.`)))
}

/**
 * What keeps concluding from sending the files of the images on the screen:
 * an image still on its way to the storage, or one that did not reach it.
 * Concluding sends the files of the images shown, so it waits for both.
 */
function imagesHolding(images: SetupImages) {
  const steps = [images.logo.upload.step, images.photo.upload.step]
  if (steps.includes("failed")) {
    return "notSent"
  }
  if (steps.includes("sending") || steps.includes("waitingForCode")) {
    return "sending"
  }
  return undefined
}

/** The name of the item of a list a field chose, or the code it holds. */
function nameIn(items: ListItem[], code: string) {
  return items.find((item) => item.value === code)?.label ?? code
}

/**
 * The step of the review: everything typed, by step, each with a way back to
 * its step, the e-mail the administrator logs in with in plain sight, and the
 * one action. Nothing is saved before it, and it sends everything once, with
 * the language of the screen as the preferred language of the administrator.
 *
 * In the first setup the action concludes the setup. In the way back in the
 * review leaves out the consultancy, says that a new administrator will be
 * registered and that the users that exist are not changed, and the action
 * registers the administrator.
 *
 * A field the API refuses is shown on its field, on the step that holds it,
 * and so is an e-mail that belongs to a user. A setup code that stopped
 * working leads back to the step of the code. A setup someone else concluded
 * leads to the login screen, which says so.
 */
export function ReviewStep({
  form,
  images,
  opened,
  onGoTo,
  onCodeStoppedWorking,
}: {
  form: SetupForm
  images: SetupImages
  opened: OpenedSetupCode
  /** Goes back to a step, to change what was typed in it. */
  onGoTo: (step: TypedStep) => void
  onCodeStoppedWorking: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()
  const navigate = useNavigate()
  const conclusion = useConcludeSetup()
  const registration = useRegisterAdministrator()
  const heading = useRef<HTMLHeadingElement>(null)
  const waitingFor = imagesHolding(images)
  const newAdministrator = opened.opens === "new_administrator"
  const sending = conclusion.isPending || registration.isPending
  const failure = conclusion.error ?? registration.error

  useEffect(() => {
    heading.current?.focus()
  }, [])

  function showRefusal(error: unknown) {
    if (answerCodeOf(error) === "setup_code_invalid") {
      onCodeStoppedWorking()
      return
    }
    const fields = fieldsRefusedBy(error)
    const step = fields ? stepHolding(fields) : undefined
    if (fields && step) {
      // On the fields, where each step shows it when it opens.
      form.setErrorMap({ onSubmit: { fields } })
      onGoTo(step)
    }
  }

  function send() {
    const sent = { typed: form.state.values, code: opened.code, language }
    if (newAdministrator) {
      registration.mutate(sent, {
        onSuccess: () =>
          void navigate({
            to: "/login",
            search: { notice: "administrator_registered" },
          }),
        onError: showRefusal,
      })
      return
    }
    conclusion.mutate(sent, {
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
      onError: showRefusal,
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <StepHeading
        ref={heading}
        title={t("setup.review.title")}
        intro={
          newAdministrator
            ? t("setup.review.newAdministrator.intro")
            : t("setup.review.intro")
        }
      />

      <form.Subscribe selector={(state) => state.values}>
        {(typed) => (
          <>
            {newAdministrator ? null : (
              <ConsultancyReview
                typed={typed.consultancy}
                logo={images.logo.upload}
                onChange={() => onGoTo("consultancy")}
              />
            )}
            <AdministratorReview
              typed={typed.administrator}
              photo={images.photo.upload}
              newAdministrator={newAdministrator}
              onChange={() => onGoTo("administrator")}
            />
          </>
        )}
      </form.Subscribe>

      {failure ? <ErrorAlert error={failure} /> : null}

      {waitingFor === "sending" ? (
        <p className="flex items-start gap-1.5 text-muted-foreground">
          <CircleDashedIcon className="mt-0.5 size-3.5 shrink-0 stroke-status-working" />
          {newAdministrator
            ? t("setup.review.newAdministrator.photoSending")
            : t("setup.review.imagesSending")}
        </p>
      ) : null}
      {waitingFor === "notSent" ? (
        <p className="flex items-start gap-1.5">
          <CircleXIcon className="mt-0.5 size-3.5 shrink-0 stroke-status-error" />
          {newAdministrator
            ? t("setup.review.newAdministrator.photoNotSent")
            : t("setup.review.imageNotSent")}
        </p>
      ) : null}

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
          disabled={sending || waitingFor !== undefined}
          onClick={send}
        >
          {sending ? (
            <LoaderCircleIcon
              data-icon="inline-start"
              className="animate-spin"
            />
          ) : null}
          {newAdministrator
            ? t("setup.review.newAdministrator.register")
            : t("setup.review.conclude")}
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

/**
 * An image, as the review shows it: from the browser's own copy, with how
 * far it is on its way to the storage, or nothing when none was chosen.
 */
function imageReview(upload: SetupImage) {
  if (upload.step === "none") {
    return null
  }
  return (
    <span className="flex items-center gap-3">
      <ImageThumbnail image={upload.image} />
      <ImageStatus upload={upload} />
    </span>
  )
}

function ConsultancyReview({
  typed,
  logo,
  onChange,
}: {
  typed: TypedSetup["consultancy"]
  logo: SetupImage
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
      <Row
        label={t("setup.consultancy.logo.label")}
        value={imageReview(logo)}
      />
    </ReviewSection>
  )
}

function AdministratorReview({
  typed,
  photo,
  newAdministrator,
  onChange,
}: {
  typed: TypedSetup["administrator"]
  photo: SetupImage
  /** Whether it is a new administrator, the way back in, or the first one. */
  newAdministrator: boolean
  onChange: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()

  return (
    <ReviewSection
      id="review-administrator"
      title={
        newAdministrator
          ? t("setup.newAdministrator.title")
          : t("setup.administrator.title")
      }
      change={
        newAdministrator
          ? t("setup.review.newAdministrator.change")
          : t("setup.review.changeAdministrator")
      }
      onChange={onChange}
      note={
        // Logging in sends a code to this e-mail: a typing mistake here
        // locks the administrator out.
        <p className="rounded-lg border border-border bg-muted px-3 py-2">
          {newAdministrator
            ? t("setup.review.newAdministrator.logsInWith")
            : t("setup.review.logsInWith")}
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
      <Row
        label={t("setup.administrator.photo.label")}
        value={imageReview(photo)}
      />
    </ReviewSection>
  )
}
