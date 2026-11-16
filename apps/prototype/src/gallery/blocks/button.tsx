import { Button, ButtonShortcut } from "@workspace/ui/components/button"
import {
  ArrowRightIcon,
  LoaderCircleIcon,
  PlusIcon,
  SettingsIcon,
} from "lucide-react"

import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const variants = [
  {
    variant: "default",
    label: "Create request",
    when: "The one primary action of a group.",
  },
  {
    variant: "secondary",
    label: "Save draft",
    when: "An action beside the primary one.",
  },
  {
    variant: "outline",
    label: "Export",
    when: "An action that stands alone on a surface.",
  },
  {
    variant: "ghost",
    label: "Cancel",
    when: "The quiet way out, and actions in a toolbar.",
  },
  {
    variant: "destructive",
    label: "Delete project",
    when: "An action that removes something.",
  },
  {
    variant: "link",
    label: "See the history",
    when: "An action that reads as part of a sentence.",
  },
] as const

const sizes = ["xs", "sm", "default", "lg"] as const

export function ButtonBlock() {
  return (
    <Section
      title="Button"
      note="Starts an action. One button in a group is primary; the others are secondary or ghost."
    >
      <Part title="Typical">
        <Frame>
          <Button>Save changes</Button>
          <Button variant="ghost">Cancel</Button>
        </Frame>
      </Part>

      <Part title="Variants">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(15rem,1fr))] gap-3">
          {variants.map(({ variant, label, when }) => (
            <Frame key={variant} className="flex-col items-start gap-2">
              <Button variant={variant}>{label}</Button>
              <span className="type-code text-muted-foreground">{variant}</span>
              <span className="type-caption text-muted-foreground">{when}</span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part title="Sizes">
        <Frame className="items-end">
          {sizes.map((size) => (
            <div key={size} className="flex flex-col items-start gap-1">
              <Button variant="secondary" size={size}>
                Open request
              </Button>
              <span className="type-code text-muted-foreground">{size}</span>
            </div>
          ))}
        </Frame>
      </Part>

      <Part
        title="With an icon or a shortcut"
        note="An icon leads or trails the label. A button with an icon alone names its action for a screen reader. The shortcut of an action is a chip at the end."
      >
        <Frame>
          <Button>
            <PlusIcon data-icon="inline-start" />
            New request
          </Button>
          <Button variant="secondary">
            Next step
            <ArrowRightIcon data-icon="inline-end" />
          </Button>
          <Button variant="outline" size="icon" aria-label="Settings">
            <SettingsIcon />
          </Button>
          <Button>
            Save changes
            <ButtonShortcut>Ctrl S</ButtonShortcut>
          </Button>
          <Button variant="secondary">
            Search
            <ButtonShortcut>Ctrl K</ButtonShortcut>
          </Button>
        </Frame>
      </Part>

      <Part
        title="States"
        note="Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={variants.map(({ variant }) => variant)}
            states={[
              "default",
              "hover",
              "focus",
              "active",
              "disabled",
              "loading",
            ]}
            render={(variant, state) =>
              state === "loading" ? (
                <Button variant={variant} disabled>
                  <LoaderCircleIcon
                    data-icon="inline-start"
                    className="animate-spin"
                  />
                  Save changes
                </Button>
              ) : (
                <Button variant={variant} {...previewState(state)}>
                  Save changes
                </Button>
              )
            }
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long label is cut with an ellipsis inside the width it is given. It never pushes the layout."
      >
        <Frame className="max-w-xs">
          <Button className="max-w-full">
            <span className="truncate">
              Enviar o pedido de alteração para a aprovação do cliente
            </span>
          </Button>
          <Button variant="secondary" className="max-w-full">
            <PlusIcon data-icon="inline-start" />
            <span className="truncate">
              Agregar una solicitud nueva a este proyecto
            </span>
          </Button>
        </Frame>
      </Part>
    </Section>
  )
}
