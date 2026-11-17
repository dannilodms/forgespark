import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import { Separator } from "@workspace/ui/components/separator"
import { useTranslation } from "react-i18next"

import { dataDisplay } from "../explanations/data-display"
import { inputs } from "../explanations/inputs"
import { partTitle, pinnedStates } from "../explanations/shared"
import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

/** The sample priorities, in the language of the prototype. */
export function usePriorities() {
  const { t } = useTranslation()
  return [
    { value: "low", label: t("gallery.common.priority.low") },
    { value: "normal", label: t("gallery.common.priority.normal") },
    { value: "urgent", label: t("gallery.common.priority.urgent") },
  ] as const
}

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
  const priorities = usePriorities()

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
  const { t } = useTranslation()
  const label = t("gallery.common.priority.label")

  return (
    <Section {...inputs.select.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <PrioritySelect label={label} />
        </Frame>
      </Part>

      <Part title={partTitle.sizes}>
        <Frame>
          <PrioritySelect label={`${label}, sm`} size="sm" />
          <PrioritySelect label={`${label}, default`} />
        </Frame>
      </Part>

      <Part {...pinnedStates}>
        <Frame>
          <StateMatrix
            rows={["trigger"] as const}
            states={["default", "hover", "focus", "disabled", "invalid"]}
            render={(_, state) => (
              <PrioritySelect
                label={`${label}, ${state}`}
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
  const { t } = useTranslation()
  const text = dataDisplay.separator

  return (
    <Section {...text.about}>
      <Part {...text.directions}>
        <Frame className="max-w-sm flex-col items-stretch">
          <span>{t("gallery.common.openRequests")}</span>
          <Separator />
          <span>{t("gallery.common.closedRequests")}</span>
        </Frame>
        <Frame className="h-14 max-w-sm">
          <span>{t("gallery.common.places.clients")}</span>
          <Separator orientation="vertical" />
          <span>{t("gallery.common.places.projects")}</span>
          <Separator orientation="vertical" />
          <span>{t("gallery.common.places.requests")}</span>
        </Frame>
      </Part>
    </Section>
  )
}
