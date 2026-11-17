import { Button } from "@workspace/ui/components/button"
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@workspace/ui/components/command"
import {
  FolderIcon,
  InboxIcon,
  MoonIcon,
  PlusIcon,
  SettingsIcon,
  UsersIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { overlays } from "../explanations/overlays"
import { freeText } from "../free-text"
import { Frame, Part, Section } from "../parts"

const text = overlays.command

function Commands() {
  const { t } = useTranslation()

  return (
    <>
      <CommandGroup heading={t("gallery.command.goTo")}>
        <CommandItem>
          <InboxIcon />
          {t("gallery.common.places.requests")}
          <CommandShortcut>G R</CommandShortcut>
        </CommandItem>
        <CommandItem>
          <FolderIcon />
          {t("gallery.common.places.projects")}
          <CommandShortcut>G P</CommandShortcut>
        </CommandItem>
        <CommandItem>
          <UsersIcon />
          {t("gallery.common.places.clients")}
          <CommandShortcut>G C</CommandShortcut>
        </CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading={t("gallery.command.actions")}>
        <CommandItem>
          <PlusIcon />
          {t("gallery.common.newRequest")}
          <CommandShortcut>N</CommandShortcut>
        </CommandItem>
        <CommandItem>
          <MoonIcon />
          {t("gallery.command.changeTheme")}
        </CommandItem>
        <CommandItem disabled>
          <SettingsIcon />
          {t("gallery.common.settings")}
        </CommandItem>
      </CommandGroup>
    </>
  )
}

function CommandPalette() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button variant="outline" onClick={() => setOpen(true)}>
        {text.open}
      </Button>
      <CommandDialog
        open={open}
        onOpenChange={setOpen}
        title={t("gallery.command.title")}
        description={t("gallery.command.description")}
      >
        <Command>
          <CommandInput placeholder={t("gallery.command.placeholder")} />
          <CommandList>
            <CommandEmpty>{t("gallery.command.empty")}</CommandEmpty>
            <Commands />
          </CommandList>
        </Command>
      </CommandDialog>
    </>
  )
}

/** A palette whose search matches nothing, so the empty state shows. */
function SearchWithNoMatch() {
  const { t } = useTranslation()
  const [search, setSearch] = useState<string>(freeText.searchWithNoMatch)

  return (
    <Command className="h-auto border">
      <CommandInput
        placeholder={t("gallery.command.placeholder")}
        value={search}
        onValueChange={setSearch}
      />
      <CommandList>
        <CommandEmpty>{t("gallery.command.empty")}</CommandEmpty>
        <Commands />
      </CommandList>
    </Command>
  )
}

export function CommandBlock() {
  const { t } = useTranslation()

  return (
    <Section {...text.about}>
      <Part {...text.typical}>
        <Frame>
          <CommandPalette />
        </Frame>
      </Part>

      <Part {...text.inPlace}>
        <Frame className="max-w-md">
          <Command className="h-auto border">
            <CommandInput placeholder={t("gallery.command.placeholder")} />
            <CommandList>
              <CommandEmpty>{t("gallery.command.empty")}</CommandEmpty>
              <Commands />
            </CommandList>
          </Command>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(17rem,1fr))] gap-3">
          <Frame>
            <SearchWithNoMatch />
          </Frame>
          <Frame>
            <Command className="h-auto border">
              <CommandInput
                placeholder={t("gallery.command.recentPlaceholder")}
              />
              <CommandList>
                <CommandEmpty>{t("gallery.command.noneOpened")}</CommandEmpty>
              </CommandList>
            </Command>
          </Frame>
          <Frame>
            <Command className="h-auto border">
              <CommandInput placeholder={t("gallery.command.placeholder")} />
              <CommandList>
                <CommandEmpty>{t("gallery.command.empty")}</CommandEmpty>
                <CommandGroup heading={t("gallery.command.goTo")}>
                  <CommandItem>
                    <InboxIcon />
                    <span className="truncate">
                      {t("gallery.command.long.waiting")}
                    </span>
                    <CommandShortcut>G A</CommandShortcut>
                  </CommandItem>
                  <CommandItem>
                    <FolderIcon />
                    <span className="truncate">
                      {t("gallery.command.long.archived")}
                    </span>
                  </CommandItem>
                </CommandGroup>
              </CommandList>
            </Command>
          </Frame>
        </div>
      </Part>
    </Section>
  )
}
