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

import { Frame, Part, Section } from "../parts"
import { FilledInputOTP } from "./input"

const priorities = [
  { value: "low", label: "Low" },
  { value: "normal", label: "Normal" },
  { value: "urgent", label: "Urgent" },
]

export function FieldBlock() {
  return (
    <Section
      title="Field"
      note="A control with its label, its description and its error. Every form is a group of fields."
    >
      <Part title="Typical">
        <Frame className="max-w-sm">
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="field-project">Project name</FieldLabel>
              <Input id="field-project" defaultValue="Billing portal" />
              <FieldDescription>
                The client sees this name on every request.
              </FieldDescription>
            </Field>
          </FieldGroup>
        </Frame>
      </Part>

      <Part
        title="With an error"
        note="The error says what is wrong and carries an icon, so it does not depend on colour."
      >
        <Frame className="max-w-sm">
          <FieldGroup>
            <Field data-invalid>
              <FieldLabel htmlFor="field-email">E-mail</FieldLabel>
              <Input id="field-email" defaultValue="ana@example" aria-invalid />
              <FieldError className="flex items-center gap-1">
                <CircleAlertIcon className="size-3.5 shrink-0" />
                Type a valid e-mail address.
              </FieldError>
            </Field>
          </FieldGroup>
        </Frame>
      </Part>

      <Part
        title="Every field disabled"
        note="The field carries the disabled state, so its label fades with its control."
      >
        <Frame className="max-w-sm">
          <FieldGroup>
            <Field data-disabled>
              <FieldLabel htmlFor="field-client">Client</FieldLabel>
              <Input
                id="field-client"
                defaultValue="Northwind Traders"
                disabled
              />
            </Field>
            <Field data-disabled>
              <FieldLabel htmlFor="field-notes">Notes</FieldLabel>
              <Textarea
                id="field-notes"
                defaultValue="Billing closes on the last working day."
                disabled
              />
            </Field>
            <Field data-disabled>
              <FieldLabel htmlFor="field-priority">Priority</FieldLabel>
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
              <FieldLabel htmlFor="field-code">Login code</FieldLabel>
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
                Notify the client
              </FieldLabel>
            </Field>
            <FieldSet className="w-full">
              <FieldLegend variant="label" className="opacity-50">
                Billing
              </FieldLegend>
              <RadioGroup defaultValue="hourly" disabled>
                <Field orientation="horizontal" data-disabled>
                  <RadioGroupItem id="field-hourly" value="hourly" />
                  <FieldLabel htmlFor="field-hourly" className="font-normal">
                    By the hour
                  </FieldLabel>
                </Field>
                <Field orientation="horizontal" data-disabled>
                  <RadioGroupItem id="field-fixed" value="fixed" />
                  <FieldLabel htmlFor="field-fixed" className="font-normal">
                    Fixed price
                  </FieldLabel>
                </Field>
              </RadioGroup>
            </FieldSet>
            <Field orientation="horizontal" data-disabled>
              <FieldLabel htmlFor="field-digest">Weekly summary</FieldLabel>
              <Switch id="field-digest" defaultChecked disabled />
            </Field>
          </FieldGroup>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long label, a long description and a long error wrap inside the width of the field."
      >
        <Frame className="max-w-xs">
          <FieldGroup>
            <Field data-invalid>
              <FieldLabel htmlFor="field-long">
                Nome do responsável pela aprovação das solicitações deste
                projeto
              </FieldLabel>
              <Input id="field-long" aria-invalid />
              <FieldDescription>
                Esta pessoa recebe um e-mail a cada solicitação que aguarda
                aprovação.
              </FieldDescription>
              <FieldError className="flex items-start gap-1">
                <CircleAlertIcon className="mt-0.5 size-3.5 shrink-0" />
                Informe o nome de uma pessoa que já está cadastrada neste
                cliente.
              </FieldError>
            </Field>
          </FieldGroup>
        </Frame>
      </Part>

      <Part
        title="Separator between two ways"
        note="A word on the line splits two ways of doing the same thing."
      >
        <Frame className="max-w-sm">
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="field-first">Invite by e-mail</FieldLabel>
              <Input id="field-first" placeholder="you@company.com" />
            </Field>
            <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
              or
            </FieldSeparator>
            <Field>
              <FieldLabel htmlFor="field-second">Paste a link</FieldLabel>
              <Input id="field-second" placeholder="https://" />
            </Field>
          </FieldGroup>
        </Frame>
      </Part>
    </Section>
  )
}

export function LabelBlock() {
  return (
    <Section
      title="Label"
      note="Names a control. Inside a field, use the field's own label."
    >
      <Part title="Typical">
        <Frame className="max-w-sm flex-col items-stretch">
          <Label htmlFor="label-search">Search requests</Label>
          <Input id="label-search" placeholder="A word of the title" />
        </Frame>
      </Part>
      <Part title="Edge cases" note="A long label wraps above its control.">
        <Frame className="max-w-xs flex-col items-stretch">
          <Label htmlFor="label-long" className="leading-snug">
            Nome do responsável pela aprovação das solicitações deste projeto
          </Label>
          <Input id="label-long" />
        </Frame>
      </Part>
    </Section>
  )
}
