import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import { Label } from "@workspace/ui/components/label"
import { CircleAlertIcon } from "lucide-react"

import { Frame, Part, Section } from "../parts"

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

      <Part title="Disabled">
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
