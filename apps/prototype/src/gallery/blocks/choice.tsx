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

import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const fieldStates = ["default", "hover", "focus", "disabled", "invalid"]

export function CheckboxBlock() {
  return (
    <Section
      title="Checkbox"
      note="One option that is on or off, or several of a list. It takes effect when the form is saved."
    >
      <Part title="Typical">
        <Frame className="max-w-sm">
          <Field orientation="horizontal">
            <Checkbox id="checkbox-notify" defaultChecked />
            <FieldContent>
              <FieldLabel htmlFor="checkbox-notify">
                Notify the client
              </FieldLabel>
              <FieldDescription>
                An e-mail goes out when the request is closed.
              </FieldDescription>
            </FieldContent>
          </Field>
        </Frame>
      </Part>

      <Part title="Several of a list">
        <Frame className="max-w-sm">
          <FieldSet className="w-full">
            <FieldLegend variant="label">Include in the export</FieldLegend>
            <FieldGroup data-slot="checkbox-group">
              <Field orientation="horizontal">
                <Checkbox id="checkbox-open" defaultChecked />
                <FieldLabel htmlFor="checkbox-open" className="font-normal">
                  Open requests
                </FieldLabel>
              </Field>
              <Field orientation="horizontal">
                <Checkbox id="checkbox-closed" defaultChecked />
                <FieldLabel htmlFor="checkbox-closed" className="font-normal">
                  Closed requests
                </FieldLabel>
              </Field>
              <Field orientation="horizontal">
                <Checkbox id="checkbox-archived" />
                <FieldLabel htmlFor="checkbox-archived" className="font-normal">
                  Archived projects
                </FieldLabel>
              </Field>
            </FieldGroup>
          </FieldSet>
        </Frame>
      </Part>

      <Part
        title="States"
        note="Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={["unchecked", "checked"] as const}
            states={fieldStates}
            render={(row, state) => (
              <Checkbox
                aria-label={`Notify the client, ${row}, ${state}`}
                defaultChecked={row === "checked"}
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long label wraps beside the box, which stays on the first line."
      >
        <Frame className="max-w-xs">
          <Field orientation="horizontal" className="items-start">
            <Checkbox id="checkbox-long" className="mt-px" />
            <FieldLabel htmlFor="checkbox-long" className="font-normal">
              Enviar um e-mail ao responsável pelo projeto sempre que uma
              solicitação mudar de situação
            </FieldLabel>
          </Field>
        </Frame>
      </Part>
    </Section>
  )
}

const priorities = [
  { value: "low", label: "Low", note: "It can wait for the next cycle." },
  { value: "normal", label: "Normal", note: "It enters the queue." },
  { value: "urgent", label: "Urgent", note: "It stops other work." },
]

export function RadioGroupBlock() {
  return (
    <Section
      title="RadioGroup"
      note="One choice from two to five options that all fit on the screen. For a longer list, use a select."
    >
      <Part title="Typical">
        <Frame className="max-w-sm">
          <FieldSet className="w-full">
            <FieldLegend variant="label">Priority</FieldLegend>
            <RadioGroup defaultValue="normal">
              {priorities.map(({ value, label }) => (
                <Field key={value} orientation="horizontal">
                  <RadioGroupItem id={`radio-${value}`} value={value} />
                  <FieldLabel
                    htmlFor={`radio-${value}`}
                    className="font-normal"
                  >
                    {label}
                  </FieldLabel>
                </Field>
              ))}
            </RadioGroup>
          </FieldSet>
        </Frame>
      </Part>

      <Part title="With a description for each option">
        <Frame className="max-w-sm">
          <FieldSet className="w-full">
            <FieldLegend variant="label">Priority</FieldLegend>
            <RadioGroup defaultValue="urgent">
              {priorities.map(({ value, label, note }) => (
                <Field key={value} orientation="horizontal">
                  <RadioGroupItem id={`radio-note-${value}`} value={value} />
                  <FieldContent>
                    <FieldLabel htmlFor={`radio-note-${value}`}>
                      {label}
                    </FieldLabel>
                    <FieldDescription>{note}</FieldDescription>
                  </FieldContent>
                </Field>
              ))}
            </RadioGroup>
          </FieldSet>
        </Frame>
      </Part>

      <Part
        title="States"
        note="Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={["unselected", "selected"] as const}
            states={fieldStates}
            render={(row, state) => (
              <RadioGroup
                aria-label={`Priority, ${row}, ${state}`}
                className="w-fit"
                defaultValue={row === "selected" ? "urgent" : undefined}
              >
                <RadioGroupItem
                  aria-label="Urgent"
                  value="urgent"
                  aria-invalid={state === "invalid" || undefined}
                  {...previewState(state)}
                />
              </RadioGroup>
            )}
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long label wraps beside its circle. The options stay one under the other."
      >
        <Frame className="max-w-xs">
          <RadioGroup aria-label="Quién aprueba" defaultValue="client">
            <Field orientation="horizontal" className="items-start">
              <RadioGroupItem
                id="radio-long-client"
                value="client"
                className="mt-px"
              />
              <FieldLabel htmlFor="radio-long-client" className="font-normal">
                La persona responsable del proyecto en la empresa del cliente
              </FieldLabel>
            </Field>
            <Field orientation="horizontal" className="items-start">
              <RadioGroupItem
                id="radio-long-team"
                value="team"
                className="mt-px"
              />
              <FieldLabel htmlFor="radio-long-team" className="font-normal">
                Cualquier persona del equipo que atiende la solicitud
              </FieldLabel>
            </Field>
          </RadioGroup>
        </Frame>
      </Part>
    </Section>
  )
}

export function SwitchBlock() {
  return (
    <Section
      title="Switch"
      note="A setting that is on or off and takes effect at once, without a save button."
    >
      <Part title="Typical">
        <Frame className="max-w-sm">
          <Field orientation="horizontal">
            <FieldContent>
              <FieldLabel htmlFor="switch-digest">Weekly summary</FieldLabel>
              <FieldDescription>
                An e-mail on Monday with the requests that moved.
              </FieldDescription>
            </FieldContent>
            <Switch id="switch-digest" defaultChecked />
          </Field>
        </Frame>
      </Part>

      <Part title="Sizes">
        <Frame className="items-end">
          {(["sm", "default"] as const).map((size) => (
            <div key={size} className="flex flex-col items-start gap-1">
              <Switch
                size={size}
                aria-label={`Weekly summary, ${size}`}
                defaultChecked
              />
              <span className="type-code text-muted-foreground">{size}</span>
            </div>
          ))}
        </Frame>
      </Part>

      <Part
        title="States"
        note="Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={["off", "on"] as const}
            states={fieldStates}
            render={(row, state) => (
              <Switch
                aria-label={`Weekly summary, ${row}, ${state}`}
                defaultChecked={row === "on"}
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long label wraps and the switch keeps its place at the end of the first line."
      >
        <Frame className="max-w-xs">
          <Field orientation="horizontal" className="items-start">
            <FieldLabel htmlFor="switch-long" className="font-normal">
              Avisar o cliente por e-mail quando uma solicitação ficar parada
              por mais de três dias
            </FieldLabel>
            <Switch id="switch-long" />
          </Field>
        </Frame>
      </Part>
    </Section>
  )
}
