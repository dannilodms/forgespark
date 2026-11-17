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
import { recipes } from "./explanations/recipes"
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
        <Breadcrumb aria-label={t("work.navigation.path")}>
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
      <PaneStandIn className="w-2/5 border-r">
        {recipes.appShell.listPane}
      </PaneStandIn>
      <PaneStandIn className="flex-1">
        {recipes.appShell.detailPane}
      </PaneStandIn>
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
        {recipes.formDialog.open}
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
 * English, in `explanations/recipes.ts`, and the sample content follows the
 * language of the prototype.
 */
export function Recipes() {
  return (
    <GalleryPage section="recipes" lead={recipes.lead}>
      <p className="max-w-prose">
        {recipes.together.before}
        <Link to="/work" className="text-brand underline underline-offset-2">
          {recipes.together.link}
        </Link>
        {recipes.together.after}
      </p>

      <Section {...recipes.appShell.about}>
        <Part {...recipes.appShell.typical}>
          <AppShellPreview />
        </Part>
      </Section>

      <Section {...recipes.listPane.about}>
        <Part {...recipes.listPane.typical}>
          <ListPanePreview />
        </Part>
      </Section>

      <Section {...recipes.detailPane.about}>
        <Part {...recipes.detailPane.typical}>
          <DetailPanePreview />
        </Part>
      </Section>

      <Section {...recipes.formDialog.about}>
        <Part {...recipes.formDialog.typical}>
          <FormDialogPreview />
        </Part>
      </Section>
    </GalleryPage>
  )
}
