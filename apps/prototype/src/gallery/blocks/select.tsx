import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import { Separator } from "@workspace/ui/components/separator"

import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const priorities = [
  { value: "low", label: "Low" },
  { value: "normal", label: "Normal" },
  { value: "urgent", label: "Urgent" },
]

function PrioritySelect({
  size,
  label,
  ...trigger
}: {
  size?: "sm" | "default"
  label: string
  disabled?: boolean
  "aria-invalid"?: boolean
  "data-preview"?: string
}) {
  return (
    <Select items={priorities} defaultValue="normal">
      <SelectTrigger size={size} aria-label={label} {...trigger}>
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
  )
}

export function SelectBlock() {
  return (
    <Section
      title="Select"
      note="One choice from a short list that does not fit on the screen as buttons."
    >
      <Part title="Typical">
        <Frame>
          <PrioritySelect label="Priority" />
        </Frame>
      </Part>

      <Part title="Sizes">
        <Frame>
          <PrioritySelect label="Priority, small" size="sm" />
          <PrioritySelect label="Priority, default" />
        </Frame>
      </Part>

      <Part
        title="States"
        note="Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={["trigger"] as const}
            states={["default", "hover", "focus", "disabled", "invalid"]}
            render={(_, state) => (
              <PrioritySelect
                label={`Priority, ${state}`}
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>
    </Section>
  )
}

export function SeparatorBlock() {
  return (
    <Section
      title="Separator"
      note="A line between two groups inside one surface. Between surfaces, the border of the surface does this."
    >
      <Part title="Horizontal and vertical">
        <Frame className="max-w-sm flex-col items-stretch">
          <span>Open requests</span>
          <Separator />
          <span>Closed requests</span>
        </Frame>
        <Frame className="h-14 max-w-sm">
          <span>Clients</span>
          <Separator orientation="vertical" />
          <span>Projects</span>
          <Separator orientation="vertical" />
          <span>Requests</span>
        </Frame>
      </Part>
    </Section>
  )
}
