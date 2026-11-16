import { Badge } from "@workspace/ui/components/badge"
import { Kbd, KbdGroup } from "@workspace/ui/components/kbd"
import { ArrowUpRightIcon, GitBranchIcon, TagIcon } from "lucide-react"

import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const variants = [
  {
    variant: "default",
    label: "New",
    when: "The one thing on the screen to notice first.",
  },
  {
    variant: "secondary",
    label: "Internal",
    when: "A category or a tag. The usual choice.",
  },
  {
    variant: "outline",
    label: "v2.4",
    when: "A quiet fact beside a title.",
  },
  {
    variant: "destructive",
    label: "Blocked",
    when: "Something is wrong. The word says what.",
  },
  {
    variant: "ghost",
    label: "Draft",
    when: "A tag that should almost disappear.",
  },
  {
    variant: "link",
    label: "3 linked",
    when: "A tag that leads somewhere.",
  },
] as const

export function BadgeBlock() {
  return (
    <Section
      title="Badge"
      note="A short word or a count beside the thing it describes. For the status of a record, use the status indicator."
    >
      <Part title="Typical">
        <Frame>
          <span className="type-title">Billing portal</span>
          <Badge variant="secondary">Internal</Badge>
          <Badge variant="outline" className="font-mono">
            v2.4
          </Badge>
        </Frame>
      </Part>

      <Part title="Variants">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(15rem,1fr))] gap-3">
          {variants.map(({ variant, label, when }) => (
            <Frame key={variant} className="flex-col items-start gap-2">
              <Badge variant={variant}>{label}</Badge>
              <span className="type-code text-muted-foreground">{variant}</span>
              <span className="type-caption text-muted-foreground">{when}</span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part
        title="With an icon or a count"
        note="An icon leads or trails the word. A count is set in mono."
      >
        <Frame>
          <Badge variant="secondary">
            <TagIcon data-icon="inline-start" />
            Billing
          </Badge>
          <Badge variant="outline">
            <GitBranchIcon data-icon="inline-start" />
            <span className="font-mono">release/2.4</span>
          </Badge>
          <Badge variant="secondary">
            Changelog
            <ArrowUpRightIcon data-icon="inline-end" />
          </Badge>
          <Badge variant="secondary" className="font-mono">
            12
          </Badge>
        </Frame>
      </Part>

      <Part
        title="States"
        note="A badge has states only when it is a link. Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={variants.map(({ variant }) => variant)}
            states={["default", "hover", "focus"]}
            render={(variant, state) => (
              <Badge
                variant={variant}
                render={
                  <a
                    href="#badge"
                    onClick={(event) => event.preventDefault()}
                  />
                }
                {...previewState(state)}
              >
                Billing
              </Badge>
            )}
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long word is cut with an ellipsis inside the width it is given."
      >
        <Frame className="max-w-xs">
          <Badge variant="secondary">
            <span className="truncate">
              Aguardando a aprovação do responsável pelo projeto
            </span>
          </Badge>
          <Badge variant="outline">
            <TagIcon data-icon="inline-start" />
            <span className="truncate">
              Pendiente de la respuesta del equipo de facturación
            </span>
          </Badge>
        </Frame>
      </Part>
    </Section>
  )
}

export function KbdBlock() {
  return (
    <Section
      title="Kbd"
      note="A key or a shortcut, named where the action it triggers is shown."
    >
      <Part title="Typical">
        <Frame>
          <span>
            Press <Kbd>/</Kbd> to search the requests of this project.
          </span>
        </Frame>
      </Part>

      <Part
        title="A shortcut of several keys"
        note="Each key is its own chip, in the order they are pressed."
      >
        <Frame className="gap-6">
          <KbdGroup>
            <Kbd>Ctrl</Kbd>
            <Kbd>K</Kbd>
          </KbdGroup>
          <KbdGroup>
            <Kbd>Ctrl</Kbd>
            <Kbd>Shift</Kbd>
            <Kbd>P</Kbd>
          </KbdGroup>
          <KbdGroup>
            <Kbd>Alt</Kbd>
            <Kbd>←</Kbd>
          </KbdGroup>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A key with a long name keeps its name whole and grows to fit it."
      >
        <Frame className="max-w-xs">
          <KbdGroup>
            <Kbd>Ctrl</Kbd>
            <Kbd>Backspace</Kbd>
          </KbdGroup>
          <KbdGroup>
            <Kbd>Shift</Kbd>
            <Kbd>Page Down</Kbd>
          </KbdGroup>
        </Frame>
      </Part>
    </Section>
  )
}
