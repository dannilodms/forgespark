import { useRef } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"
import { FieldGroup, FieldLegend, FieldSet } from "#ui/components/field"

import {
  FieldRow,
  ListField,
  TextField,
  focusTheFirstProblem,
} from "../../shared/components/form-fields"
import {
  type FieldProblems,
  checkTheContract,
} from "../../shared/errors/problems"
import { useLanguage } from "../../shared/locales/i18n"
import {
  countryItems,
  currencyItems,
  timeZoneItems,
} from "../../shared/places/places"
import { consultancyContract } from "../data/setup"
import { ImageField } from "./image-field"
import {
  type SetupForm,
  fieldsWithProblems,
  suggestTheCurrency,
  useStepOpening,
} from "./setup-form"
import type { SetupImageField } from "./setup-image"
import { StepActions } from "./step-actions"
import { StepHeading } from "./step-heading"

const checkTheConsultancy = checkTheContract(
  consultancyContract.schema,
  consultancyContract.toBody
)

/**
 * The step of the consultancy: how it is known, where it is, how it is
 * reached, and its logo. The country suggests the currency. Moving on checks
 * the fields of this step only, as the contract declares them; a logo still
 * on its way to the storage keeps going meanwhile.
 */
export function ConsultancyStep({
  form,
  logo,
  refused,
  onProblemsShown,
  onContinue,
}: {
  form: SetupForm
  /** The logo, which the screen holds so that it outlives the step. */
  logo: SetupImageField
  /** The fields the API refused, which the step shows when it opens. */
  refused: FieldProblems | undefined
  onProblemsShown: () => void
  onContinue: () => void
}) {
  const { t } = useTranslation("consultancy")
  const { language } = useLanguage()
  const element = useRef<HTMLFormElement>(null)
  const heading = useRef<HTMLHeadingElement>(null)

  useStepOpening({
    form,
    step: "consultancy",
    refused,
    onProblemsShown,
    element,
    heading,
  })

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
            <FieldRow>
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
            </FieldRow>
            <FieldRow>
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
            </FieldRow>
            <FieldRow>
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
            </FieldRow>
            <FieldRow>
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
            </FieldRow>
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
                <FieldRow>
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
                </FieldRow>
                <FieldRow columns={3}>
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
                </FieldRow>
              </FieldGroup>
            </FieldSet>

            <form.Field name="consultancy.logo_file_id">
              {(field) => (
                <ImageField
                  field={field}
                  image={logo}
                  texts={{
                    label: t("setup.consultancy.logo.label"),
                    choose: t("setup.consultancy.logo.choose"),
                    replace: t("setup.consultancy.logo.replace"),
                    remove: t("setup.consultancy.logo.remove"),
                  }}
                />
              )}
            </form.Field>
          </FieldGroup>

          <StepActions>
            <Button type="submit">{t("setup.continue")}</Button>
          </StepActions>
        </form>
      )}
    </form.FormGroup>
  )
}
