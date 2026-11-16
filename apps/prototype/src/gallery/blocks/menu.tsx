import { Button } from "@workspace/ui/components/button"
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import { Field, FieldGroup, FieldLabel } from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import { Kbd } from "@workspace/ui/components/kbd"
import {
  Popover,
  PopoverContent,
  PopoverDescription,
  PopoverHeader,
  PopoverTitle,
  PopoverTrigger,
} from "@workspace/ui/components/popover"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@workspace/ui/components/tooltip"
import {
  ArchiveIcon,
  ChevronDownIcon,
  CopyIcon,
  PencilIcon,
  SearchIcon,
  SettingsIcon,
  Trash2Icon,
  UserPlusIcon,
} from "lucide-react"
import { useState } from "react"

import { Frame, Part, Section } from "../parts"

function ViewMenu() {
  const [showClosed, setShowClosed] = useState(true)
  const [showArchived, setShowArchived] = useState(false)
  const [order, setOrder] = useState("recent")

  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="outline" />}>
        Open the view menu
        <ChevronDownIcon data-icon="inline-end" />
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56">
        <DropdownMenuGroup>
          <DropdownMenuLabel>Show</DropdownMenuLabel>
          <DropdownMenuCheckboxItem
            checked={showClosed}
            onCheckedChange={setShowClosed}
          >
            Closed requests
          </DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem
            checked={showArchived}
            onCheckedChange={setShowArchived}
          >
            Archived projects
          </DropdownMenuCheckboxItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuLabel>Order by</DropdownMenuLabel>
          <DropdownMenuRadioGroup value={order} onValueChange={setOrder}>
            <DropdownMenuRadioItem value="recent">
              Most recent
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="priority">
              Priority
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="client">Client</DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function DropdownMenuBlock() {
  return (
    <Section
      title="DropdownMenu"
      note="The actions on one thing, behind one button. The arrow keys move through it, a letter jumps to an item, and Escape closes it."
    >
      <Part title="Typical">
        <Frame>
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="outline" />}>
              Open the request menu
              <ChevronDownIcon data-icon="inline-end" />
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56">
              <DropdownMenuGroup>
                <DropdownMenuItem>
                  <PencilIcon />
                  Rename
                  <DropdownMenuShortcut>F2</DropdownMenuShortcut>
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <CopyIcon />
                  Copy the link
                  <DropdownMenuShortcut>Ctrl L</DropdownMenuShortcut>
                </DropdownMenuItem>
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger>
                    <UserPlusIcon />
                    Assign to
                  </DropdownMenuSubTrigger>
                  <DropdownMenuSubContent>
                    <DropdownMenuItem>Ana Lima</DropdownMenuItem>
                    <DropdownMenuItem>Bruno Tavares</DropdownMenuItem>
                    <DropdownMenuItem>Carla Mendes</DropdownMenuItem>
                  </DropdownMenuSubContent>
                </DropdownMenuSub>
                <DropdownMenuItem disabled>
                  <ArchiveIcon />
                  Archive
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem variant="destructive">
                <Trash2Icon />
                Delete request
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </Frame>
      </Part>

      <Part
        title="Options that stay chosen"
        note="An item can be a checkbox or one of a group of radios, under a label that names the group."
      >
        <Frame>
          <ViewMenu />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long item is cut with an ellipsis inside the width of the menu."
      >
        <Frame>
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="outline" />}>
              Open a menu with long items
              <ChevronDownIcon data-icon="inline-end" />
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-64">
              <DropdownMenuItem>
                <PencilIcon />
                <span className="truncate">
                  Alterar o responsável pela aprovação das solicitações
                </span>
              </DropdownMenuItem>
              <DropdownMenuItem>
                <CopyIcon />
                <span className="truncate">
                  Copiar el enlace de la solicitud al portapapeles
                </span>
                <DropdownMenuShortcut>Ctrl L</DropdownMenuShortcut>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </Frame>
      </Part>
    </Section>
  )
}

export function PopoverBlock() {
  return (
    <Section
      title="Popover"
      note="A small panel beside the button that opens it, for a few fields or a detail. The page behind stays usable, and Escape closes it."
    >
      <Part title="Typical">
        <Frame>
          <Popover>
            <PopoverTrigger render={<Button variant="outline" />}>
              Open the estimate popover
            </PopoverTrigger>
            <PopoverContent align="start">
              <PopoverHeader>
                <PopoverTitle>Estimate</PopoverTitle>
                <PopoverDescription>
                  The hours this request is expected to take.
                </PopoverDescription>
              </PopoverHeader>
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="popover-hours">Hours</FieldLabel>
                  <Input
                    id="popover-hours"
                    defaultValue="12"
                    inputMode="numeric"
                  />
                </Field>
              </FieldGroup>
              <Button size="sm" className="self-end">
                Save estimate
              </Button>
            </PopoverContent>
          </Popover>
        </Frame>
      </Part>

      <Part
        title="Sides"
        note="A popover opens under its button by default, and moves when there is no room."
      >
        <Frame>
          {(["top", "right", "bottom", "left"] as const).map((side) => (
            <Popover key={side}>
              <PopoverTrigger render={<Button variant="outline" />}>
                Open a popover on the {side}
              </PopoverTrigger>
              <PopoverContent side={side} className="w-56">
                <PopoverHeader>
                  <PopoverTitle>Billing portal</PopoverTitle>
                  <PopoverDescription>
                    7 open requests, 42.5 hours this month.
                  </PopoverDescription>
                </PopoverHeader>
              </PopoverContent>
            </Popover>
          ))}
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long text wraps inside the popover, which never grows wider than the screen."
      >
        <Frame>
          <Popover>
            <PopoverTrigger render={<Button variant="outline" />}>
              Open a popover with a long text
            </PopoverTrigger>
            <PopoverContent align="start">
              <PopoverHeader>
                <PopoverTitle>
                  Por que esta solicitação está parada há mais de três dias
                </PopoverTitle>
                <PopoverDescription>
                  O responsável pelo projeto no cliente ainda não respondeu ao
                  pedido de aprovação enviado na última segunda-feira.
                </PopoverDescription>
              </PopoverHeader>
            </PopoverContent>
          </Popover>
        </Frame>
      </Part>
    </Section>
  )
}

export function TooltipBlock() {
  return (
    <Section
      title="Tooltip"
      note="The name of a control that shows only an icon, with its shortcut. It opens on hover and on keyboard focus, and never holds something a person has to read to go on."
    >
      <Part
        title="Typical"
        note="Move the pointer over the button, or reach it with Tab."
      >
        <Frame>
          <Tooltip>
            <TooltipTrigger
              render={
                <Button variant="outline" size="icon" aria-label="Settings" />
              }
            >
              <SettingsIcon />
            </TooltipTrigger>
            <TooltipContent>Settings</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger
              render={
                <Button variant="outline" size="icon" aria-label="Search" />
              }
            >
              <SearchIcon />
            </TooltipTrigger>
            <TooltipContent>
              Search
              <Kbd>/</Kbd>
            </TooltipContent>
          </Tooltip>
        </Frame>
      </Part>

      <Part
        title="Sides"
        note="A tooltip opens above its control by default, and moves when there is no room."
      >
        <Frame>
          {(["top", "right", "bottom", "left"] as const).map((side) => (
            <Tooltip key={side}>
              <TooltipTrigger render={<Button variant="outline" />}>
                Tooltip on the {side}
              </TooltipTrigger>
              <TooltipContent side={side}>Opens on the {side}</TooltipContent>
            </Tooltip>
          ))}
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long text wraps at the widest a tooltip gets."
      >
        <Frame>
          <Tooltip>
            <TooltipTrigger render={<Button variant="outline" />}>
              Tooltip with a long text
            </TooltipTrigger>
            <TooltipContent>
              Esta solicitação aguarda a aprovação do responsável pelo projeto
              no cliente desde a última segunda-feira.
            </TooltipContent>
          </Tooltip>
        </Frame>
      </Part>
    </Section>
  )
}
