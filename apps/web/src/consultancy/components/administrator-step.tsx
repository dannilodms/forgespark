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
import { type SetupForm, fieldsWithProblems } from "./setup-form"

const checkTheAdministrator = checkTheContract(
  administratorContract.schema,
  administratorContract.toBody
)

/**
 * The step of the first administrator: the person who logs in first and
 * registers the others. Moving on checks the fields of this step only; going
 * back keeps them as typed, checked or not. The preferred language is not
 * asked: it is the language of the screen when the setup is concluded.
 */
export function AdministratorStep({
  form,
  onContinue,
  onBack,
}: {
  form: SetupForm
  onContinue: () => void
  onBack: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()
  const element = useRef<HTMLFormElement>(null)

  // A step opened on a field the API refused puts the focus on it.
  useEffect(() => {
    focusTheFirstProblem(element.current, fieldsWithProblems(form))
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
          <header className="flex flex-col gap-1">
            <h2 className="type-title">{t("setup.administrator.title")}</h2>
            <p className="text-muted-foreground">
              {t("setup.administrator.intro")}
            </p>
          </header>

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
