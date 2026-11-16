import { Link } from "@tanstack/react-router"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@workspace/ui/components/breadcrumb"
import { Button } from "@workspace/ui/components/button"
import { CommandGroup, CommandItem } from "@workspace/ui/components/command"
import { DropdownMenuItem } from "@workspace/ui/components/dropdown-menu"
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@workspace/ui/components/sidebar"
import { cn } from "@workspace/ui/lib/utils"
import { FolderIcon, InboxIcon, LogOutIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { AppShell, UserMenu } from "../recipes/app-shell"
import { sampleItems, samplePerson } from "../work/sample-data"
import {
  type Place,
  changeStatus,
  filterItems,
  places,
} from "../work/work-items"
import { useStatusWord } from "../work/status-word"
import { ItemDetail, ItemList, NewItemDialog } from "../work/work-parts"
import { GalleryPage, Part, Section } from "./parts"

const placeIcons = { requests: InboxIcon, projects: FolderIcon }

const sampleRequests = sampleItems.filter((item) => item.place === "requests")

const sampleProjects = sampleItems
  .filter((item) => item.place === "projects")
  .map((item) => item.title)

/** Where a pane goes in the app shell. The pane itself is another recipe. */
function PaneStandIn({
  className,
  children,
}: {
  className?: string
  children: string
}) {
  return (
    <div
      className={cn(
        "flex min-w-0 items-center justify-center bg-card p-3 text-center type-caption text-muted-foreground",
        className
      )}
    >
      {children}
    </div>
  )
}

function AppShellPreview() {
  const { t } = useTranslation()
  const [place, setPlace] = useState<Place>("requests")

  return (
    <AppShell
      className="h-96 overflow-hidden rounded-xl border"
      product={t("work.productName")}
      userMenu={
        <UserMenu
          name={samplePerson.name}
          email={samplePerson.email}
          label={t("work.userMenu.label", { name: samplePerson.name })}
        >
          <DropdownMenuItem>
            <LogOutIcon />
            {t("work.userMenu.logOut")}
          </DropdownMenuItem>
        </UserMenu>
      }
      navigation={
        <SidebarGroup>
          <SidebarGroupLabel>{t("work.navigation.group")}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {places.map((entry) => {
                const Icon = placeIcons[entry]
                return (
                  <SidebarMenuItem key={entry}>
                    <SidebarMenuButton
                      isActive={entry === place}
                      tooltip={t(`work.places.${entry}`)}
                      onClick={() => setPlace(entry)}
                    >
                      <Icon />
                      <span>{t(`work.places.${entry}`)}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      }
      navigationLabel={t("work.navigation.label")}
      sidebarToggleLabel={t("work.navigation.toggle")}
      breadcrumb={
        <Breadcrumb>
          <BreadcrumbList className="flex-nowrap">
            <BreadcrumbItem className="shrink-0">
              {t(`work.places.${place}`)}
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage className="truncate font-mono">
                {place === "requests" ? "REQ-0142" : "PRJ-0012"}
              </BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
      }
      commandPalette={{
        trigger: t("work.commands.trigger"),
        shortcut: "Ctrl K",
        placeholder: t("work.commands.placeholder"),
        description: t("work.commands.description"),
        empty: t("work.commands.empty"),
        commands: (close) => (
          <CommandGroup heading={t("work.commands.goTo")}>
            {places.map((entry) => {
              const Icon = placeIcons[entry]
              return (
                <CommandItem
                  key={entry}
                  onSelect={() => {
                    close()
                    setPlace(entry)
                  }}
                >
                  <Icon />
                  {t(`work.places.${entry}`)}
                </CommandItem>
              )
            })}
          </CommandGroup>
        ),
      }}
      statusBar={<span className="truncate">{t(`work.places.${place}`)}</span>}
    >
      <PaneStandIn className="w-2/5 border-r">List pane</PaneStandIn>
      <PaneStandIn className="flex-1">Detail pane</PaneStandIn>
    </AppShell>
  )
}

function ListPanePreview() {
  const statusWord = useStatusWord()
  const [typed, setTyped] = useState("")
  const [selectedId, setSelectedId] = useState("REQ-0141")

  return (
    <ItemList
      place="requests"
      shown={filterItems(sampleRequests, typed, (item) =>
        statusWord(item.place, item.status)
      )}
      typed={typed}
      onTyped={setTyped}
      selectedId={selectedId}
      onSelect={setSelectedId}
      onNew={() => {}}
      className="h-80 max-w-md overflow-hidden rounded-xl border"
    />
  )
}

function DetailPanePreview() {
  const [item, setItem] = useState(sampleRequests[0])

  return (
    <ItemDetail
      item={item}
      onStatusChange={(status) =>
        setItem(
          changeStatus(
            [item],
            item.id,
            status,
            samplePerson.name,
            item.updated
          )[0]
        )
      }
      className="h-80 max-w-xl overflow-hidden rounded-xl border"
    />
  )
}

function FormDialogPreview() {
  const [open, setOpen] = useState(false)

  return (
    <div>
      <Button variant="outline" onClick={() => setOpen(true)}>
        Open the form dialog
      </Button>
      <NewItemDialog
        place="requests"
        groups={sampleProjects}
        open={open}
        onOpenChange={setOpen}
        onCreate={() => {}}
      />
    </div>
  )
}

/**
 * The four recipes of DESIGN.md, each on its own. The explanatory text is in
 * English and the sample content follows the language of the prototype.
 */
export function Recipes() {
  return (
    <GalleryPage
      title="Recipes"
      lead="A recipe is a named composition of components that several screens reuse. Each of the four is shown here on its own, over sample data."
    >
      <p className="max-w-prose">
        The four are put together in the{" "}
        <Link to="/work" className="text-brand underline underline-offset-2">
          sample work screen
        </Link>
        , where the login screen leads.
      </p>

      <Section
        title="App shell"
        note="Use it around every screen of the product: it holds what stays in place while the panes change."
      >
        <Part
          title="Typical"
          note="A title bar with the user menu, the sidebar navigation, a top bar with the breadcrumb and the command palette trigger, the main area of flush panes, and a status bar. The button of the top bar collapses the sidebar, and Ctrl K opens the command palette."
        >
          <AppShellPreview />
        </Part>
      </Section>

      <Section
        title="List pane"
        note="Use it for a set of items a person scans, filters and picks from, beside the detail of the one picked."
      >
        <Part
          title="Typical"
          note="A filter, rows that each carry a status indicator, and an empty state when the filter leaves nothing. The arrow keys move between the rows, and Enter selects one."
        >
          <ListPanePreview />
        </Part>
      </Section>

      <Section
        title="Detail pane"
        note="Use it for the one item picked in a list pane: what it is, and what can be done to it."
      >
        <Part
          title="Typical"
          note="A header with the identifier, the status, the actions and the title of the item, and tabs that switch its content."
        >
          <DetailPanePreview />
        </Part>
      </Section>

      <Section
        title="Form dialog"
        note="Use it to create or change one thing with a few fields, without leaving the screen."
      >
        <Part
          title="Typical"
          note="Fields, one primary action and a cancel action. Create with no title to see the validation error on the field."
        >
          <FormDialogPreview />
        </Part>
      </Section>
    </GalleryPage>
  )
}
