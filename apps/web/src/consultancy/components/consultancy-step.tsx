import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"
import { FieldGroup, FieldLegend, FieldSet } from "#ui/components/field"

import {
  ListField,
  TextField,
  focusTheFirstProblem,
} from "../../shared/components/form-fields"
import { checkTheContract } from "../../shared/errors/problems"
import { useLanguage } from "../../shared/locales/i18n"
import {
  countryItems,
  currencyItems,
  timeZoneItems,
} from "../../shared/places/places"
import { consultancyContract } from "../data/setup"
import {
  type SetupForm,
  fieldsWithProblems,
  suggestTheCurrency,
} from "./setup-form"
import { StepHeading } from "./step-heading"

const checkTheConsultancy = checkTheContract(
  consultancyContract.schema,
  consultancyContract.toBody
)

/**
 * The step of the consultancy: how it is known, where it is, and how it is
 * reached. The country suggests the currency. Moving on checks the fields of
 * this step only, as the contract declares them.
 */
export function ConsultancyStep({
  form,
  onContinue,
}: {
  form: SetupForm
  onContinue: () => void
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
      name="consultancy"
      validators={{ onSubmit: checkTheConsultancy }}
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
            title={t("setup.consultancy.title")}
            intro={t("setup.consultancy.intro")}
          />

          <FieldGroup>
            <form.Field name="consultancy.name">
              {(field) => (
                <TextField
                  field={field}
                  label={t("setup.consultancy.name")}
                  description={t("setup.consultancy.nameDescription")}
                  autoComplete="organization"
                />
              )}
            </form.Field>
            <form.Field name="consultancy.legal_name">
              {(field) => (
                <TextField
                  field={field}
                  label={t("setup.consultancy.legalName")}
                  description={t("setup.consultancy.legalNameDescription")}
                  autoComplete="off"
                />
              )}
            </form.Field>
            <form.Field
              name="consultancy.country"
              listeners={{
                onChange: ({ value }) => suggestTheCurrency(form, value),
              }}
            >
              {(field) => (
                <ListField
                  field={field}
                  items={countryItems(language)}
                  label={t("setup.consultancy.country")}
                />
              )}
            </form.Field>
            <form.Field name="consultancy.tax_id">
              {(field) => (
                <TextField
                  field={field}
                  label={t("setup.consultancy.taxId")}
                  description={t("setup.consultancy.taxIdDescription")}
                  autoComplete="off"
                  spellCheck={false}
                />
              )}
            </form.Field>
            <form.Field name="consultancy.time_zone">
              {(field) => (
                <ListField
                  field={field}
                  items={timeZoneItems(language)}
                  label={t("setup.consultancy.timeZone")}
                />
              )}
            </form.Field>
            <form.Field name="consultancy.currency">
              {(field) => (
                <ListField
                  field={field}
                  items={currencyItems(language)}
                  label={t("setup.consultancy.currency")}
                  description={t("setup.consultancy.currencyDescription")}
                />
              )}
            </form.Field>
            <form.Field name="consultancy.email">
              {(field) => (
                <TextField
                  field={field}
                  optional
                  label={t("setup.consultancy.email")}
                  type="email"
                  autoComplete="email"
                  spellCheck={false}
                />
              )}
            </form.Field>
            <form.Field name="consultancy.telephone">
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
            <form.Field name="consultancy.website">
              {(field) => (
                <TextField
                  field={field}
                  optional
                  label={t("setup.consultancy.website")}
                  description={t("setup.consultancy.websiteDescription")}
                  type="url"
                  autoComplete="url"
                  spellCheck={false}
                />
              )}
            </form.Field>

            <FieldSet>
              <FieldLegend variant="label">
                {t("setup.consultancy.address.legend")}
              </FieldLegend>
              <FieldGroup>
                <form.Field name="consultancy.address.line1">
                  {(field) => (
                    <TextField
                      field={field}
                      optional
                      label={t("setup.consultancy.address.line1")}
                      autoComplete="address-line1"
                    />
                  )}
                </form.Field>
                <form.Field name="consultancy.address.line2">
                  {(field) => (
                    <TextField
                      field={field}
                      optional
                      label={t("setup.consultancy.address.line2")}
                      autoComplete="address-line2"
                    />
                  )}
                </form.Field>
                <form.Field name="consultancy.address.city">
                  {(field) => (
                    <TextField
                      field={field}
                      optional
                      label={t("setup.consultancy.address.city")}
                      autoComplete="address-level2"
                    />
                  )}
                </form.Field>
                <form.Field name="consultancy.address.region">
                  {(field) => (
                    <TextField
                      field={field}
                      optional
                      label={t("setup.consultancy.address.region")}
                      autoComplete="address-level1"
                    />
                  )}
                </form.Field>
                <form.Field name="consultancy.address.postal_code">
                  {(field) => (
                    <TextField
                      field={field}
                      optional
                      label={t("setup.consultancy.address.postalCode")}
                      autoComplete="postal-code"
                    />
                  )}
                </form.Field>
              </FieldGroup>
            </FieldSet>
          </FieldGroup>

          <Button type="submit">{t("setup.continue")}</Button>
        </form>
      )}
    </form.FormGroup>
  )
}
