import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupText,
  InputGroupTextarea,
} from "@workspace/ui/components/input-group"
import { Kbd } from "@workspace/ui/components/kbd"
import { Textarea } from "@workspace/ui/components/textarea"
import { CopyIcon, SearchIcon, SendIcon } from "lucide-react"

import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

export function TextareaBlock() {
  return (
    <Section
      title="Textarea"
      note="Several lines of text typed by a person. It grows with what is typed."
    >
      <Part title="Typical">
        <Frame className="max-w-sm">
          <Textarea
            aria-label="Description"
            placeholder="What does the client need, and why?"
          />
        </Frame>
      </Part>

      <Part
        title="States"
        note="Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={["empty", "filled"] as const}
            states={["default", "hover", "focus", "disabled", "invalid"]}
            render={(row, state) => (
              <Textarea
                aria-label={`Description, ${row}, ${state}`}
                className="w-36"
                placeholder="Description"
                defaultValue={
                  row === "filled"
                    ? "The export leaves out archived projects."
                    : undefined
                }
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long text wraps and the field grows. A word with no space breaks inside the field."
      >
        <Frame className="max-w-xs">
          <Textarea
            aria-label="Long description"
            defaultValue="O cliente pediu que a exportação inclua os projetos arquivados e os anexos de cada solicitação: https://clientes.exemplo.com.br/projetos/portal-de-cobranca/solicitacoes/exportacao"
          />
        </Frame>
      </Part>
    </Section>
  )
}

export function InputGroupBlock() {
  return (
    <Section
      title="InputGroup"
      note="A field with something attached inside its border: an icon, a unit, a shortcut or a small button. The command palette uses it for its search field."
    >
      <Part title="Typical">
        <Frame className="max-w-sm">
          <InputGroup>
            <InputGroupInput
              aria-label="Search requests"
              placeholder="Search requests"
            />
            <InputGroupAddon>
              <SearchIcon />
            </InputGroupAddon>
            <InputGroupAddon align="inline-end">
              <Kbd>/</Kbd>
            </InputGroupAddon>
          </InputGroup>
        </Frame>
      </Part>

      <Part
        title="What can be attached"
        note="A text at the start or the end, a button at the end, and a row of actions under a textarea."
      >
        <Frame className="max-w-sm flex-col flex-nowrap items-stretch">
          <InputGroup>
            <InputGroupAddon>
              <InputGroupText className="font-mono">https://</InputGroupText>
            </InputGroupAddon>
            <InputGroupInput
              aria-label="Address of the client"
              defaultValue="northwind.example"
            />
          </InputGroup>
          <InputGroup>
            <InputGroupInput
              aria-label="Hours estimated"
              defaultValue="12"
              inputMode="numeric"
            />
            <InputGroupAddon align="inline-end">
              <InputGroupText>hours</InputGroupText>
            </InputGroupAddon>
          </InputGroup>
          <InputGroup>
            <InputGroupInput
              aria-label="Link to the request"
              defaultValue="REQ-0142"
              className="font-mono"
              readOnly
            />
            <InputGroupAddon align="inline-end">
              <InputGroupButton size="icon-xs" aria-label="Copy the link">
                <CopyIcon />
              </InputGroupButton>
            </InputGroupAddon>
          </InputGroup>
          <InputGroup>
            <InputGroupTextarea
              aria-label="Comment"
              placeholder="Write a comment"
            />
            <InputGroupAddon align="block-end" className="justify-end">
              <InputGroupButton variant="default">
                <SendIcon />
                Send
              </InputGroupButton>
            </InputGroupAddon>
          </InputGroup>
        </Frame>
      </Part>

      <Part
        title="States"
        note="The group shows the focus, the error and the disabled state of the control inside it."
      >
        <Frame>
          <StateMatrix
            rows={["group"] as const}
            states={["default", "focus", "disabled", "invalid"]}
            render={(_, state) => (
              <InputGroup className="w-44">
                <InputGroupInput
                  aria-label={`Search, ${state}`}
                  placeholder="Search"
                  aria-invalid={state === "invalid" || undefined}
                  {...previewState(state)}
                />
                <InputGroupAddon>
                  <SearchIcon />
                </InputGroupAddon>
              </InputGroup>
            )}
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long attached text keeps its width and the control takes what is left."
      >
        <Frame className="max-w-xs">
          <InputGroup>
            <InputGroupAddon>
              <InputGroupText className="font-mono">
                clientes.exemplo.com.br/
              </InputGroupText>
            </InputGroupAddon>
            <InputGroupInput
              aria-label="Path of the project"
              defaultValue="portal-de-cobranca"
            />
          </InputGroup>
        </Frame>
      </Part>
    </Section>
  )
}
