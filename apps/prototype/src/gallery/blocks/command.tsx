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

import { Frame, Part, Section } from "../parts"

function Commands() {
  return (
    <>
      <CommandGroup heading="Go to">
        <CommandItem>
          <InboxIcon />
          Requests
          <CommandShortcut>G R</CommandShortcut>
        </CommandItem>
        <CommandItem>
          <FolderIcon />
          Projects
          <CommandShortcut>G P</CommandShortcut>
        </CommandItem>
        <CommandItem>
          <UsersIcon />
          Clients
          <CommandShortcut>G C</CommandShortcut>
        </CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Actions">
        <CommandItem>
          <PlusIcon />
          New request
          <CommandShortcut>N</CommandShortcut>
        </CommandItem>
        <CommandItem>
          <MoonIcon />
          Change the theme
        </CommandItem>
        <CommandItem disabled>
          <SettingsIcon />
          Settings
        </CommandItem>
      </CommandGroup>
    </>
  )
}

function CommandPalette() {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button variant="outline" onClick={() => setOpen(true)}>
        Open the command palette
      </Button>
      <CommandDialog
        open={open}
        onOpenChange={setOpen}
        title="Command palette"
        description="Search for a place to go or an action to run."
      >
        <Command>
          <CommandInput placeholder="Search for a place or an action" />
          <CommandList>
            <CommandEmpty>Nothing matches this search.</CommandEmpty>
            <Commands />
          </CommandList>
        </Command>
      </CommandDialog>
    </>
  )
}

/** A palette whose search matches nothing, so the empty state shows. */
function SearchWithNoMatch() {
  const [search, setSearch] = useState("payroll")

  return (
    <Command className="h-auto border">
      <CommandInput
        placeholder="Search for a place or an action"
        value={search}
        onValueChange={setSearch}
      />
      <CommandList>
        <CommandEmpty>Nothing matches this search.</CommandEmpty>
        <Commands />
      </CommandList>
    </Command>
  )
}

export function CommandBlock() {
  return (
    <Section
      title="Command"
      note="A list of places and actions filtered by typing. As the command palette, it opens over any screen. The arrow keys move through it and Enter runs the item."
    >
      <Part
        title="Typical"
        note="The palette opens in a dialog, with the search field already focused."
      >
        <Frame>
          <CommandPalette />
        </Frame>
      </Part>

      <Part
        title="In place"
        note="The same list set into a pane, for a picker that stays on the screen."
      >
        <Frame className="max-w-md">
          <Command className="h-auto border">
            <CommandInput placeholder="Search for a place or an action" />
            <CommandList>
              <CommandEmpty>Nothing matches this search.</CommandEmpty>
              <Commands />
            </CommandList>
          </Command>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A search that matches nothing says so, and a list with no item says there is nothing to choose. A long item is cut with an ellipsis."
      >
        <div className="grid grid-cols-[repeat(auto-fill,minmax(17rem,1fr))] gap-3">
          <Frame>
            <SearchWithNoMatch />
          </Frame>
          <Frame>
            <Command className="h-auto border">
              <CommandInput placeholder="Search the recent requests" />
              <CommandList>
                <CommandEmpty>No request was opened yet.</CommandEmpty>
              </CommandList>
            </Command>
          </Frame>
          <Frame>
            <Command className="h-auto border">
              <CommandInput placeholder="Buscar un lugar o una acción" />
              <CommandList>
                <CommandEmpty>Nada coincide con esta búsqueda.</CommandEmpty>
                <CommandGroup heading="Ir a">
                  <CommandItem>
                    <InboxIcon />
                    <span className="truncate">
                      Solicitudes que esperan la aprobación del cliente
                    </span>
                    <CommandShortcut>G A</CommandShortcut>
                  </CommandItem>
                  <CommandItem>
                    <FolderIcon />
                    <span className="truncate">
                      Projetos arquivados nos últimos doze meses
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
