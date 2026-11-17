import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"
import { FieldGroup } from "#ui/components/field"

import {
  ListField,
  TextField,
  focusTheFirstProblem,
} from "../../shared/components/form-fields"
import { checkTheContract } from "../../shared/errors/problems"
import { useLanguage } from "../../shared/locales/i18n"
import { countryItems } from "../../shared/places/places"
import { administratorContract } from "../data/setup"
import { ImageField } from "./image-field"
import { type SetupForm, fieldsWithProblems } from "./setup-form"
import type { SetupImageField } from "./setup-image"
import { StepHeading } from "./step-heading"

const checkTheAdministrator = checkTheContract(
  administratorContract.schema,
  administratorContract.toBody
)

/**
 * The step of the first administrator: the person who logs in first and
 * registers the others. Moving on checks the fields of this step only; going
 * back keeps them as typed, checked or not. The preferred language is not
 * asked: it is the language of the screen when the setup is concluded. A
 * photo still on its way to the storage keeps going while the person moves.
 */
export function AdministratorStep({
  form,
  photo,
  onContinue,
  onBack,
}: {
  form: SetupForm
  /** The photo, which the screen holds so that it outlives the step. */
  photo: SetupImageField
  onContinue: () => void
  onBack: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()
  const element = useRef<HTMLFormElement>(null)
  const heading = useRef<HTMLHeadingElement>(null)

  // The focus goes to a field the API refused, or else to the step.
  useEffect(() => {
    if (!focusTheFirstProblem(element.current, fieldsWithProblems(form))) {
      heading.current?.focus()
    }
  }, [form])

  return (
    <form.FormGroup
      name="administrator"
      validators={{ onSubmit: checkTheAdministrator }}
      onGroupSubmit={onContinue}
      onGroupSubmitInvalid={() =>
        focusTheFirstProblem(element.current, fieldsWithProblems(form))
      }
    >
      {(group) => (
        <form
          ref={element}
          noValidate
          className="flex flex-col gap-6"
          onSubmit={(event) => {
            event.preventDefault()
            void group.handleSubmit()
          }}
        >
          <StepHeading
            ref={heading}
            title={t("setup.administrator.title")}
            intro={t("setup.administrator.intro")}
          />

          <FieldGroup>
            <form.Field name="administrator.name">
              {(field) => (
                <TextField
                  field={field}
                  label={t("setup.administrator.name")}
                  autoComplete="name"
                />
              )}
            </form.Field>
            <form.Field name="administrator.email">
              {(field) => (
                <TextField
                  field={field}
                  label={t("setup.administrator.email")}
                  description={t("setup.administrator.emailDescription")}
                  type="email"
                  autoComplete="email"
                  spellCheck={false}
                />
              )}
            </form.Field>
            <form.Field name="administrator.country">
              {(field) => (
                <ListField
                  field={field}
                  items={countryItems(language)}
                  label={t("setup.administrator.country")}
                  description={t("setup.administrator.countryDescription")}
                />
              )}
            </form.Field>
            <form.Field name="administrator.national_id">
              {(field) => (
                <TextField
                  field={field}
                  label={t("setup.administrator.nationalId")}
                  description={t("setup.administrator.nationalIdDescription")}
                  autoComplete="off"
                  spellCheck={false}
                />
              )}
            </form.Field>
            <form.Field name="administrator.telephone">
              {(field) => (
                <TextField
                  field={field}
                  optional
                  label={t("setup.telephone")}
                  description={t("setup.telephoneDescription")}
                  type="tel"
                  autoComplete="tel"
                />
              )}
            </form.Field>
            <form.Field name="administrator.photo_file_id">
              {(field) => (
                <ImageField
                  field={field}
                  image={photo}
                  texts={{
                    label: t("setup.administrator.photo.label"),
                    choose: t("setup.administrator.photo.choose"),
                    replace: t("setup.administrator.photo.replace"),
                  }}
                />
              )}
            </form.Field>
          </FieldGroup>

          <div className="flex gap-2">
            <Button type="button" variant="outline" onClick={onBack}>
              {t("setup.back")}
            </Button>
            <Button type="submit" className="flex-1">
              {t("setup.continue")}
            </Button>
          </div>
        </form>
      )}
    </form.FormGroup>
  )
}
