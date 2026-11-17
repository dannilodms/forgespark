import { Checkbox } from "@workspace/ui/components/checkbox"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSeparator,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import { InputOTPGroup, InputOTPSlot } from "@workspace/ui/components/input-otp"
import { Label } from "@workspace/ui/components/label"
import {
  RadioGroup,
  RadioGroupItem,
} from "@workspace/ui/components/radio-group"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import { Switch } from "@workspace/ui/components/switch"
import { Textarea } from "@workspace/ui/components/textarea"
import { CircleAlertIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { inputs } from "../explanations/inputs"
import { partTitle } from "../explanations/shared"
import { freeText } from "../free-text"
import { Frame, Part, Section } from "../parts"
import { FilledInputOTP } from "./input"
import { usePriorities } from "./select"

export function FieldBlock() {
  const { t } = useTranslation()
  const priorities = usePriorities()
  const text = inputs.field

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm">
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="field-project">
                {t("gallery.common.projectName")}
              </FieldLabel>
              <Input id="field-project" defaultValue={freeText.project} />
              <FieldDescription>
                {t("gallery.field.projectNameHint")}
              </FieldDescription>
            </Field>
          </FieldGroup>
        </Frame>
      </Part>

      <Part {...text.withError}>
        <Frame className="max-w-sm">
          <FieldGroup>
            <Field data-invalid>
              <FieldLabel htmlFor="field-email">
                {t("gallery.common.email")}
              </FieldLabel>
              <Input
                id="field-email"
                defaultValue={freeText.unfinishedEmail}
                aria-invalid
              />
              <FieldError className="flex items-center gap-1">
                <CircleAlertIcon className="size-3.5 shrink-0" />
                {t("gallery.field.emailInvalid")}
              </FieldError>
            </Field>
          </FieldGroup>
        </Frame>
      </Part>

      <Part {...text.disabled}>
        <Frame className="max-w-sm">
          <FieldGroup>
            <Field data-disabled>
              <FieldLabel htmlFor="field-client">
                {t("gallery.common.client")}
              </FieldLabel>
              <Input
                id="field-client"
                defaultValue={freeText.client}
                disabled
              />
            </Field>
            <Field data-disabled>
              <FieldLabel htmlFor="field-notes">
                {t("gallery.field.notes")}
              </FieldLabel>
              <Textarea
                id="field-notes"
                defaultValue={freeText.note}
                disabled
              />
            </Field>
            <Field data-disabled>
              <FieldLabel htmlFor="field-priority">
                {t("gallery.common.priority.label")}
              </FieldLabel>
              <Select items={priorities} defaultValue="normal" disabled>
                <SelectTrigger id="field-priority" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {priorities.map((priority) => (
                      <SelectItem key={priority.value} value={priority.value}>
                        {priority.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field data-disabled>
              <FieldLabel htmlFor="field-code">
                {t("gallery.common.loginCode")}
              </FieldLabel>
              <FilledInputOTP
                id="field-code"
                maxLength={6}
                initial="481"
                disabled
              >
                <InputOTPGroup>
                  {[0, 1, 2, 3, 4, 5].map((index) => (
                    <InputOTPSlot key={index} index={index} />
                  ))}
                </InputOTPGroup>
              </FilledInputOTP>
            </Field>
            <Field orientation="horizontal" data-disabled>
              <Checkbox id="field-notify" defaultChecked disabled />
              <FieldLabel htmlFor="field-notify" className="font-normal">
                {t("gallery.common.notifyClient")}
              </FieldLabel>
            </Field>
            <FieldSet className="w-full">
              <FieldLegend variant="label" className="opacity-50">
                {t("gallery.field.billing.label")}
              </FieldLegend>
              <RadioGroup defaultValue="hourly" disabled>
                <Field orientation="horizontal" data-disabled>
                  <RadioGroupItem id="field-hourly" value="hourly" />
                  <FieldLabel htmlFor="field-hourly" className="font-normal">
                    {t("gallery.field.billing.hourly")}
                  </FieldLabel>
                </Field>
                <Field orientation="horizontal" data-disabled>
                  <RadioGroupItem id="field-fixed" value="fixed" />
                  <FieldLabel htmlFor="field-fixed" className="font-normal">
                    {t("gallery.field.billing.fixed")}
                  </FieldLabel>
                </Field>
              </RadioGroup>
            </FieldSet>
            <Field orientation="horizontal" data-disabled>
              <FieldLabel htmlFor="field-digest">
                {t("gallery.common.weeklySummary")}
              </FieldLabel>
              <Switch id="field-digest" defaultChecked disabled />
            </Field>
          </FieldGroup>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <FieldGroup>
            <Field data-invalid>
              <FieldLabel htmlFor="field-long">
                {t("gallery.field.long.label")}
              </FieldLabel>
              <Input id="field-long" aria-invalid />
              <FieldDescription>
                {t("gallery.field.long.description")}
              </FieldDescription>
              <FieldError className="flex items-start gap-1">
                <CircleAlertIcon className="mt-0.5 size-3.5 shrink-0" />
                {t("gallery.field.long.error")}
              </FieldError>
            </Field>
          </FieldGroup>
        </Frame>
      </Part>

      <Part {...text.separator}>
        <Frame className="max-w-sm">
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="field-first">
                {t("gallery.field.inviteByEmail")}
              </FieldLabel>
              <Input
                id="field-first"
                placeholder={t("gallery.common.emailPlaceholder")}
              />
            </Field>
            <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
              {t("gallery.field.or")}
            </FieldSeparator>
            <Field>
              <FieldLabel htmlFor="field-second">
                {t("gallery.field.pasteLink")}
              </FieldLabel>
              <Input id="field-second" placeholder="https://" />
            </Field>
          </FieldGroup>
        </Frame>
      </Part>
    </Section>
  )
}

export function LabelBlock() {
  const { t } = useTranslation()
  const text = inputs.label

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm flex-col items-stretch">
          <Label htmlFor="label-search">
            {t("gallery.common.searchRequests")}
          </Label>
          <Input
            id="label-search"
            placeholder={t("gallery.field.searchPlaceholder")}
          />
        </Frame>
      </Part>
      <Part {...text.edgeCases}>
        <Frame className="max-w-xs flex-col items-stretch">
          <Label htmlFor="label-long" className="leading-snug">
            {t("gallery.field.long.label")}
          </Label>
          <Input id="label-long" />
        </Frame>
      </Part>
    </Section>
  )
}
