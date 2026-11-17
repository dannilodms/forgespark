import { Checkbox } from "@workspace/ui/components/checkbox"
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@workspace/ui/components/field"
import {
  RadioGroup,
  RadioGroupItem,
} from "@workspace/ui/components/radio-group"
import { Switch } from "@workspace/ui/components/switch"
import { useTranslation } from "react-i18next"

import { inputs } from "../explanations/inputs"
import { partTitle, pinnedStates } from "../explanations/shared"
import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const fieldStates = ["default", "hover", "focus", "disabled", "invalid"]

export function CheckboxBlock() {
  const { t } = useTranslation()
  const text = inputs.checkbox

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm">
          <Field orientation="horizontal">
            <Checkbox id="checkbox-notify" defaultChecked />
            <FieldContent>
              <FieldLabel htmlFor="checkbox-notify">
                {t("gallery.common.notifyClient")}
              </FieldLabel>
              <FieldDescription>
                {t("gallery.checkbox.notifyHint")}
              </FieldDescription>
            </FieldContent>
          </Field>
        </Frame>
      </Part>

      <Part {...text.several}>
        <Frame className="max-w-sm">
          <FieldSet className="w-full">
            <FieldLegend variant="label">
              {t("gallery.checkbox.includeInExport")}
            </FieldLegend>
            <FieldGroup data-slot="checkbox-group">
              <Field orientation="horizontal">
                <Checkbox id="checkbox-open" defaultChecked />
                <FieldLabel htmlFor="checkbox-open" className="font-normal">
                  {t("gallery.common.openRequests")}
                </FieldLabel>
              </Field>
              <Field orientation="horizontal">
                <Checkbox id="checkbox-closed" defaultChecked />
                <FieldLabel htmlFor="checkbox-closed" className="font-normal">
                  {t("gallery.common.closedRequests")}
                </FieldLabel>
              </Field>
              <Field orientation="horizontal">
                <Checkbox id="checkbox-archived" />
                <FieldLabel htmlFor="checkbox-archived" className="font-normal">
                  {t("gallery.common.archivedProjects")}
                </FieldLabel>
              </Field>
            </FieldGroup>
          </FieldSet>
        </Frame>
      </Part>

      <Part {...pinnedStates}>
        <Frame>
          <StateMatrix
            rows={["unchecked", "checked"] as const}
            states={fieldStates}
            render={(row, state) => (
              <Checkbox
                aria-label={`${t("gallery.common.notifyClient")}, ${row}, ${state}`}
                defaultChecked={row === "checked"}
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <Field orientation="horizontal" className="items-start">
            <Checkbox id="checkbox-long" className="mt-px" />
            <FieldLabel htmlFor="checkbox-long" className="font-normal">
              {t("gallery.checkbox.long")}
            </FieldLabel>
          </Field>
        </Frame>
      </Part>
    </Section>
  )
}

/** The sample priorities, each with the key of its label and of its description. */
const priorities = [
  {
    value: "low",
    label: "gallery.common.priority.low",
    note: "gallery.radio.lowNote",
  },
  {
    value: "normal",
    label: "gallery.common.priority.normal",
    note: "gallery.radio.normalNote",
  },
  {
    value: "urgent",
    label: "gallery.common.priority.urgent",
    note: "gallery.radio.urgentNote",
  },
] as const

export function RadioGroupBlock() {
  const { t } = useTranslation()
  const text = inputs.radioGroup
  const priority = t("gallery.common.priority.label")

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm">
          <FieldSet className="w-full">
            <FieldLegend variant="label">{priority}</FieldLegend>
            <RadioGroup defaultValue="normal">
              {priorities.map(({ value, label }) => (
                <Field key={value} orientation="horizontal">
                  <RadioGroupItem id={`radio-${value}`} value={value} />
                  <FieldLabel
                    htmlFor={`radio-${value}`}
                    className="font-normal"
                  >
                    {t(label)}
                  </FieldLabel>
                </Field>
              ))}
            </RadioGroup>
          </FieldSet>
        </Frame>
      </Part>

      <Part {...text.withDescription}>
        <Frame className="max-w-sm">
          <FieldSet className="w-full">
            <FieldLegend variant="label">{priority}</FieldLegend>
            <RadioGroup defaultValue="urgent">
              {priorities.map(({ value, label, note }) => (
                <Field key={value} orientation="horizontal">
                  <RadioGroupItem id={`radio-note-${value}`} value={value} />
                  <FieldContent>
                    <FieldLabel htmlFor={`radio-note-${value}`}>
                      {t(label)}
                    </FieldLabel>
                    <FieldDescription>{t(note)}</FieldDescription>
                  </FieldContent>
                </Field>
              ))}
            </RadioGroup>
          </FieldSet>
        </Frame>
      </Part>

      <Part {...pinnedStates}>
        <Frame>
          <StateMatrix
            rows={["unselected", "selected"] as const}
            states={fieldStates}
            render={(row, state) => (
              <RadioGroup
                aria-label={`${priority}, ${row}, ${state}`}
                className="w-fit"
                defaultValue={row === "selected" ? "urgent" : undefined}
              >
                <RadioGroupItem
                  aria-label={t("gallery.common.priority.urgent")}
                  value="urgent"
                  aria-invalid={state === "invalid" || undefined}
                  {...previewState(state)}
                />
              </RadioGroup>
            )}
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <RadioGroup
            aria-label={t("gallery.radio.whoApproves")}
            defaultValue="client"
          >
            <Field orientation="horizontal" className="items-start">
              <RadioGroupItem
                id="radio-long-client"
                value="client"
                className="mt-px"
              />
              <FieldLabel htmlFor="radio-long-client" className="font-normal">
                {t("gallery.radio.longClient")}
              </FieldLabel>
            </Field>
            <Field orientation="horizontal" className="items-start">
              <RadioGroupItem
                id="radio-long-team"
                value="team"
                className="mt-px"
              />
              <FieldLabel htmlFor="radio-long-team" className="font-normal">
                {t("gallery.radio.longTeam")}
              </FieldLabel>
            </Field>
          </RadioGroup>
        </Frame>
      </Part>
    </Section>
  )
}

export function SwitchBlock() {
  const { t } = useTranslation()
  const text = inputs.switch
  const summary = t("gallery.common.weeklySummary")

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm">
          <Field orientation="horizontal">
            <FieldContent>
              <FieldLabel htmlFor="switch-digest">{summary}</FieldLabel>
              <FieldDescription>
                {t("gallery.switch.summaryHint")}
              </FieldDescription>
            </FieldContent>
            <Switch id="switch-digest" defaultChecked />
          </Field>
        </Frame>
      </Part>

      <Part title={partTitle.sizes}>
        <Frame className="items-end">
          {(["sm", "default"] as const).map((size) => (
            <div key={size} className="flex flex-col items-start gap-1">
              <Switch
                size={size}
                aria-label={`${summary}, ${size}`}
                defaultChecked
              />
              <span className="type-code text-muted-foreground">{size}</span>
            </div>
          ))}
        </Frame>
      </Part>

      <Part {...pinnedStates}>
        <Frame>
          <StateMatrix
            rows={["off", "on"] as const}
            states={fieldStates}
            render={(row, state) => (
              <Switch
                aria-label={`${summary}, ${row}, ${state}`}
                defaultChecked={row === "on"}
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <Field orientation="horizontal" className="items-start">
            <FieldLabel htmlFor="switch-long" className="font-normal">
              {t("gallery.switch.long")}
            </FieldLabel>
            <Switch id="switch-long" />
          </Field>
        </Frame>
      </Part>
    </Section>
  )
}
