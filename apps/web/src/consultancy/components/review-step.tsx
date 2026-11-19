import { useNavigate } from "@tanstack/react-router"
import { CircleDashedIcon, CircleXIcon, LoaderCircleIcon } from "lucide-react"
import { type ReactNode, useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"
import { cn } from "#ui/lib/utils"

import { ErrorAlert } from "../../shared/errors/error-alert"
import {
  type FieldProblems,
  answerCodeOf,
  answerOnAField,
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
import type { SetupForm, TypedStep } from "./setup-form"
import {
  type SetupImage,
  type SetupImages,
  whyTheImagesAreNotStored,
} from "./setup-image"
import { type SetupMode, modeOf } from "./setup-mode"
import { StepActions } from "../../shared/components/step-actions"
import { StepHeading } from "./step-heading"

/**
 * The fields a refusal of the API is about, by their names in the form. An
 * e-mail that belongs to a user is answered as a conflict, about the e-mail
 * of the administrator, which is where it shows.
 */
function fieldsRefusedBy(error: unknown): FieldProblems | undefined {
  const emailInUse = answerOnAField(error, "email_in_use")
  if (emailInUse) {
    return { "administrator.email": emailInUse }
  }
  return refusedFields(error)
}

/** What the login screen says after concluding, by how concluding ended. */
const noticeOf = {
  concluded: "setup_concluded",
  concludedBefore: "setup_already_concluded",
  concludedWithoutAnswer: "setup_concluded_without_answer",
} as const

/**
 * What the login screen says after registering a new administrator, by how
 * registering ended.
 */
const registrationNoticeOf = {
  registered: "administrator_registered",
  registeredWithoutAnswer: "administrator_registered_without_answer",
} as const

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
 * leads to the login screen, which says so. After an attempt of this page
 * that got no answer, a refusal that attempt may explain leads to the login
 * screen too, which says the setup is concluded, or the administrator may be
 * registered, and never that nothing was saved.
 */
export function ReviewStep({
  form,
  images,
  opened,
  onGoTo,
  onFieldsRefused,
  onCodeStoppedWorking,
}: {
  form: SetupForm
  images: SetupImages
  opened: OpenedSetupCode
  /** Goes back to a step, to change what was typed in it. */
  onGoTo: (step: TypedStep) => void
  /** The API refused fields, which the steps that hold them show. */
  onFieldsRefused: (fields: FieldProblems) => void
  onCodeStoppedWorking: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()
  const navigate = useNavigate()
  const conclusion = useConcludeSetup()
  const registration = useRegisterAdministrator()
  const heading = useRef<HTMLHeadingElement>(null)
  const waitingFor = whyTheImagesAreNotStored(images)
  const mode = modeOf(opened)
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
    if (fields) {
      onFieldsRefused(fields)
    }
  }

  function send() {
    const sent = { typed: form.state.values, code: opened.code, language }
    if (!mode.hasConsultancy) {
      registration.mutate(sent, {
        onSuccess: (registered) =>
          void navigate({
            to: "/login",
            search: { notice: registrationNoticeOf[registered] },
          }),
        onError: showRefusal,
      })
      return
    }
    conclusion.mutate(sent, {
      onSuccess: (concluded) =>
        void navigate({
          to: "/login",
          search: { notice: noticeOf[concluded] },
        }),
      onError: showRefusal,
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <StepHeading
        ref={heading}
        title={t("setup.review.title")}
        intro={t(mode.texts.reviewIntro)}
      />

      <form.Subscribe selector={(state) => state.values}>
        {(typed) => (
          <>
            {mode.hasConsultancy ? (
              <ConsultancyReview
                typed={typed.consultancy}
                logo={images.logo.upload}
                onChange={() => onGoTo("consultancy")}
              />
            ) : null}
            <AdministratorReview
              typed={typed.administrator}
              photo={images.photo.upload}
              mode={mode}
              onChange={() => onGoTo("administrator")}
            />
          </>
        )}
      </form.Subscribe>

      {failure ? <ErrorAlert error={failure} /> : null}

      {waitingFor === "sending" ? (
        <p className="flex items-start gap-1.5 text-muted-foreground">
          <CircleDashedIcon className="mt-1 size-4 shrink-0 stroke-status-working" />
          {t(mode.texts.imagesSending)}
        </p>
      ) : null}
      {waitingFor === "notSent" ? (
        <p className="flex items-start gap-1.5">
          <CircleXIcon className="mt-1 size-4 shrink-0 stroke-status-error" />
          {t(mode.texts.imagesNotSent)}
        </p>
      ) : null}

      <StepActions
        back={
          <Button
            type="button"
            variant="outline"
            onClick={() => onGoTo("administrator")}
          >
            {t("setup.back")}
          </Button>
        }
      >
        <Button
          type="button"
          disabled={sending || waitingFor !== undefined}
          onClick={send}
        >
          {sending ? (
            <LoaderCircleIcon
              data-icon="inline-start"
              className="animate-spin"
            />
          ) : null}
          {t(mode.texts.send)}
        </Button>
      </StepActions>
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
      className="@container/review flex flex-col gap-3 border-t border-border-subtle pt-4"
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
      <dl className="grid gap-x-6 gap-y-3 @lg/review:grid-cols-2">
        {children}
      </dl>
      {note}
    </section>
  )
}

/**
 * One item of the review: the label of a field above what was typed in it.
 * The items go in two columns where the section is wide enough; a wide item
 * takes both.
 */
function Row({
  label,
  value,
  wide = false,
}: {
  label: string
  value: ReactNode
  wide?: boolean
}) {
  const { t } = useTranslation("consultancy")
  const given = value !== "" && value !== null

  return (
    <div
      className={cn(
        "flex min-w-0 flex-col gap-0.5",
        wide && "@lg/review:col-span-2"
      )}
    >
      <dt className="type-caption text-muted-foreground">{label}</dt>
      <dd className="min-w-0 break-words">
        {given ? (
          value
        ) : (
          <span className="text-faint-foreground">
            {t("setup.review.notGiven")}
          </span>
        )}
      </dd>
    </div>
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
        label={t("setup.consultancy.logo.label")}
        value={imageReview(logo)}
      />
      <Row
        wide
        label={t("setup.consultancy.address.legend")}
        value={address.length > 0 ? address.join(" · ") : null}
      />
    </ReviewSection>
  )
}

function AdministratorReview({
  typed,
  photo,
  mode,
  onChange,
}: {
  typed: TypedSetup["administrator"]
  photo: SetupImage
  /** The mode of the setup: the first administrator, or a new one. */
  mode: SetupMode
  onChange: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()

  return (
    <ReviewSection
      id="review-administrator"
      title={t(mode.texts.administratorTitle)}
      change={t(mode.texts.changeAdministrator)}
      onChange={onChange}
      note={
        // Logging in sends a code to this e-mail: a typing mistake here
        // locks the administrator out.
        <p className="rounded-lg border border-border bg-muted px-3 py-2">
          {t(mode.texts.logsInWith)}
          <span className="block type-title wrap-anywhere">{typed.email}</span>
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
