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
import { useTranslation } from "react-i18next"

import { useFormat } from "#stores/language"

import { overlays } from "../explanations/overlays"
import { partTitle } from "../explanations/shared"
import { freeText } from "../free-text"
import { Frame, Part, Section } from "../parts"

function ViewMenu() {
  const { t } = useTranslation()
  const [showClosed, setShowClosed] = useState(true)
  const [showArchived, setShowArchived] = useState(false)
  const [order, setOrder] = useState("recent")

  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="outline" />}>
        {overlays.dropdownMenu.openView}
        <ChevronDownIcon data-icon="inline-end" />
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56">
        <DropdownMenuGroup>
          <DropdownMenuLabel>{t("gallery.menu.show")}</DropdownMenuLabel>
          <DropdownMenuCheckboxItem
            checked={showClosed}
            onCheckedChange={setShowClosed}
          >
            {t("gallery.common.closedRequests")}
          </DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem
            checked={showArchived}
            onCheckedChange={setShowArchived}
          >
            {t("gallery.common.archivedProjects")}
          </DropdownMenuCheckboxItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuLabel>{t("gallery.menu.orderBy")}</DropdownMenuLabel>
          <DropdownMenuRadioGroup value={order} onValueChange={setOrder}>
            <DropdownMenuRadioItem value="recent">
              {t("gallery.menu.mostRecent")}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="priority">
              {t("gallery.common.priority.label")}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="client">
              {t("gallery.common.client")}
            </DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function DropdownMenuBlock() {
  const { t } = useTranslation()
  const text = overlays.dropdownMenu

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="outline" />}>
              {text.openRequest}
              <ChevronDownIcon data-icon="inline-end" />
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56">
              <DropdownMenuGroup>
                <DropdownMenuItem>
                  <PencilIcon />
                  {t("gallery.menu.rename")}
                  <DropdownMenuShortcut>F2</DropdownMenuShortcut>
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <CopyIcon />
                  {t("gallery.common.copyLink")}
                  <DropdownMenuShortcut>Ctrl L</DropdownMenuShortcut>
                </DropdownMenuItem>
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger>
                    <UserPlusIcon />
                    {t("gallery.menu.assignTo")}
                  </DropdownMenuSubTrigger>
                  <DropdownMenuSubContent>
                    <DropdownMenuItem>{freeText.ana}</DropdownMenuItem>
                    <DropdownMenuItem>{freeText.bruno}</DropdownMenuItem>
                    <DropdownMenuItem>{freeText.carla}</DropdownMenuItem>
                  </DropdownMenuSubContent>
                </DropdownMenuSub>
                <DropdownMenuItem disabled>
                  <ArchiveIcon />
                  {t("gallery.menu.archive")}
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem variant="destructive">
                <Trash2Icon />
                {t("gallery.menu.deleteRequest")}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </Frame>
      </Part>

      <Part {...text.options}>
        <Frame>
          <ViewMenu />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame>
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="outline" />}>
              {text.openLong}
              <ChevronDownIcon data-icon="inline-end" />
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-64">
              <DropdownMenuItem>
                <PencilIcon />
                <span className="truncate">{t("gallery.menu.longChange")}</span>
              </DropdownMenuItem>
              <DropdownMenuItem>
                <CopyIcon />
                <span className="truncate">{t("gallery.menu.longCopy")}</span>
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
  const { t } = useTranslation()
  const { number } = useFormat()
  const text = overlays.popover

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Popover>
            <PopoverTrigger render={<Button variant="outline" />}>
              {text.openEstimate}
            </PopoverTrigger>
            <PopoverContent align="start">
              <PopoverHeader>
                <PopoverTitle>
                  {t("gallery.popover.estimate.title")}
                </PopoverTitle>
                <PopoverDescription>
                  {t("gallery.popover.estimate.description")}
                </PopoverDescription>
              </PopoverHeader>
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="popover-hours">
                    {t("gallery.common.places.hours")}
                  </FieldLabel>
                  <Input
                    id="popover-hours"
                    defaultValue={number(12)}
                    inputMode="numeric"
                  />
                </Field>
              </FieldGroup>
              <Button size="sm" className="self-end">
                {t("gallery.popover.estimate.save")}
              </Button>
            </PopoverContent>
          </Popover>
        </Frame>
      </Part>

      <Part {...text.sides}>
        <Frame>
          {(["top", "right", "bottom", "left"] as const).map((side) => (
            <Popover key={side}>
              <PopoverTrigger render={<Button variant="outline" />}>
                {text.openOn[side]}
              </PopoverTrigger>
              <PopoverContent side={side} className="w-56">
                <PopoverHeader>
                  <PopoverTitle>{freeText.project}</PopoverTitle>
                  <PopoverDescription>
                    {t("gallery.popover.summary", {
                      count: 7,
                      hours: number(42.5, 1),
                    })}
                  </PopoverDescription>
                </PopoverHeader>
              </PopoverContent>
            </Popover>
          ))}
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame>
          <Popover>
            <PopoverTrigger render={<Button variant="outline" />}>
              {text.openLong}
            </PopoverTrigger>
            <PopoverContent align="start">
              <PopoverHeader>
                <PopoverTitle>{t("gallery.popover.long.title")}</PopoverTitle>
                <PopoverDescription>
                  {t("gallery.popover.long.description")}
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
  const { t } = useTranslation()
  const text = overlays.tooltip
  const settings = t("gallery.common.settings")
  const search = t("gallery.common.search")

  return (
    <Section {...text.about}>
      <Part {...text.typical}>
        <Frame>
          <Tooltip>
            <TooltipTrigger
              render={
                <Button variant="outline" size="icon" aria-label={settings} />
              }
            >
              <SettingsIcon />
            </TooltipTrigger>
            <TooltipContent>{settings}</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger
              render={
                <Button variant="outline" size="icon" aria-label={search} />
              }
            >
              <SearchIcon />
            </TooltipTrigger>
            <TooltipContent>
              {search}
              <Kbd>/</Kbd>
            </TooltipContent>
          </Tooltip>
        </Frame>
      </Part>

      <Part {...text.sides}>
        <Frame>
          {(["top", "right", "bottom", "left"] as const).map((side) => (
            <Tooltip key={side}>
              <TooltipTrigger render={<Button variant="outline" />}>
                {text.side[side].trigger}
              </TooltipTrigger>
              <TooltipContent side={side}>
                {text.side[side].content}
              </TooltipContent>
            </Tooltip>
          ))}
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame>
          <Tooltip>
            <TooltipTrigger render={<Button variant="outline" />}>
              {text.withLong}
            </TooltipTrigger>
            <TooltipContent>{t("gallery.tooltip.long")}</TooltipContent>
          </Tooltip>
        </Frame>
      </Part>
    </Section>
  )
}
