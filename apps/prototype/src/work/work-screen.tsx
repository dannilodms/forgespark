import { Link, useNavigate } from "@tanstack/react-router"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@workspace/ui/components/breadcrumb"
import {
  CommandGroup,
  CommandItem,
  CommandSeparator,
} from "@workspace/ui/components/command"
import { DropdownMenuItem } from "@workspace/ui/components/dropdown-menu"
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@workspace/ui/components/sidebar"
import { StatusIndicator } from "@workspace/ui/components/status-indicator"
import { FolderIcon, InboxIcon, LogOutIcon, PlusIcon } from "lucide-react"
import { useState, useSyncExternalStore } from "react"
import { useTranslation } from "react-i18next"

import { AppShell, UserMenu } from "../recipes/app-shell"
import { sampleClients, samplePerson } from "./sample-data"
import {
  type Place,
  type WorkItem,
  countByStatus,
  filterItems,
  places,
  statuses,
} from "./work-items"
import { useStatusWord } from "./status-word"
import { ItemDetail, ItemList, NewItemDialog } from "./work-parts"
import {
  changeWorkItemStatus,
  createWorkItem,
  useWorkItems,
} from "./work-store"

const placeIcons = { requests: InboxIcon, projects: FolderIcon }

/* The width from which the list and the detail fit side by side: `lg`. */
const sideBySideQuery = "(min-width: 64rem)"

function subscribeToWidth(onChange: () => void) {
  const query = window.matchMedia(sideBySideQuery)
  query.addEventListener("change", onChange)
  return () => query.removeEventListener("change", onChange)
}

/** Whether the screen is wide enough for the list and the detail side by side. */
function useSideBySide() {
  return useSyncExternalStore(
    subscribeToWidth,
    () => window.matchMedia(sideBySideQuery).matches
  )
}

/** The places of the sidebar. On a narrow screen, going to one closes the sheet. */
function Places({ current, items }: { current: Place; items: WorkItem[] }) {
  const { t } = useTranslation()
  const { setOpenMobile } = useSidebar()

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{t("work.navigation.group")}</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {places.map((place) => {
            const Icon = placeIcons[place]
            return (
              <SidebarMenuItem key={place}>
                <SidebarMenuButton
                  isActive={place === current}
                  tooltip={t(`work.places.${place}`)}
                  render={
                    <Link
                      to="/work"
                      search={{ place }}
                      aria-current={place === current ? "page" : undefined}
                      onClick={() => setOpenMobile(false)}
                    />
                  }
                >
                  <Icon />
                  <span>{t(`work.places.${place}`)}</span>
                </SidebarMenuButton>
                <SidebarMenuBadge>
                  {items.filter((item) => item.place === place).length}
                </SidebarMenuBadge>
              </SidebarMenuItem>
            )
          })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}

/**
 * The sample work screen: the four recipes of DESIGN.md put together, over
 * sample data, so the style can be judged on a full work surface. The place
 * and the selected item are in the address, so the browser goes back through
 * them; what is typed in the filter and the open dialog are state of the
 * screen.
 *
 * On a wide screen the list and the detail sit side by side, and the first
 * item fills the detail while none was picked. On a narrow one the screen
 * shows the list, then the detail of the item picked, with a way back.
 */
export function WorkScreen({
  place,
  itemId,
}: {
  place: Place
  itemId?: string
}) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const statusWord = useStatusWord()
  const items = useWorkItems()
  const [filter, setFilter] = useState({ place, typed: "" })
  const [creating, setCreating] = useState(false)
  const sideBySide = useSideBySide()

  // The filter belongs to a place: another place starts with none.
  const typed = filter.place === place ? filter.typed : ""

  const inPlace = items.filter((item) => item.place === place)
  const shown = filterItems(inPlace, typed, (item) =>
    statusWord(item.place, item.status)
  )
  const asked = inPlace.find((item) => item.id === itemId)
  const selected = asked ?? (sideBySide ? inPlace[0] : undefined)
  const counts = countByStatus(inPlace)

  const groups =
    place === "requests"
      ? items.filter((item) => item.place === "projects").map((i) => i.title)
      : sampleClients

  const goTo = (to: Place, item?: string) =>
    void navigate({ to: "/work", search: { place: to, item } })

  return (
    <>
      <AppShell
        product={t("work.productName")}
        userMenu={
          <UserMenu
            name={samplePerson.name}
            email={samplePerson.email}
            label={t("work.userMenu.label", { name: samplePerson.name })}
          >
            <DropdownMenuItem onClick={() => void navigate({ to: "/login" })}>
              <LogOutIcon />
              {t("work.userMenu.logOut")}
            </DropdownMenuItem>
          </UserMenu>
        }
        navigation={<Places current={place} items={items} />}
        navigationLabel={t("work.navigation.label")}
        sidebarToggleLabel={t("work.navigation.toggle")}
        breadcrumb={
          <Breadcrumb>
            <BreadcrumbList className="flex-nowrap">
              <BreadcrumbItem className="shrink-0">
                <BreadcrumbLink render={<Link to="/work" search={{ place }} />}>
                  {t(`work.places.${place}`)}
                </BreadcrumbLink>
              </BreadcrumbItem>
              {selected ? (
                <>
                  <BreadcrumbSeparator />
                  <BreadcrumbItem>
                    <BreadcrumbPage className="truncate font-mono">
                      {selected.id}
                    </BreadcrumbPage>
                  </BreadcrumbItem>
                </>
              ) : null}
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
            <>
              <CommandGroup heading={t("work.commands.goTo")}>
                {places.map((to) => {
                  const Icon = placeIcons[to]
                  return (
                    <CommandItem
                      key={to}
                      onSelect={() => {
                        close()
                        goTo(to)
                      }}
                    >
                      <Icon />
                      {t(`work.places.${to}`)}
                    </CommandItem>
                  )
                })}
              </CommandGroup>
              <CommandSeparator />
              <CommandGroup heading={t("work.commands.actions")}>
                <CommandItem
                  onSelect={() => {
                    close()
                    setCreating(true)
                  }}
                >
                  <PlusIcon />
                  {t(`work.new.${place}`)}
                </CommandItem>
                <CommandItem
                  onSelect={() => {
                    close()
                    void navigate({ to: "/login" })
                  }}
                >
                  <LogOutIcon />
                  {t("work.userMenu.logOut")}
                </CommandItem>
              </CommandGroup>
              <CommandSeparator />
              <CommandGroup heading={t("work.commands.open")}>
                {items.map((item) => (
                  <CommandItem
                    key={item.id}
                    value={`${item.id} ${item.title}`}
                    onSelect={() => {
                      close()
                      goTo(item.place, item.id)
                    }}
                  >
                    <span className="shrink-0 type-code text-muted-foreground">
                      {item.id}
                    </span>
                    <span className="truncate">{item.title}</span>
                  </CommandItem>
                ))}
              </CommandGroup>
            </>
          ),
        }}
        statusBar={
          <>
            <span className="flex min-w-0 items-center gap-1.5">
              <span className="truncate">{t(`work.places.${place}`)}</span>
              <span className="type-code">{inPlace.length}</span>
            </span>
            {statuses.map((status) => (
              <StatusIndicator
                key={status}
                status={status}
                title={statusWord(place, status)}
                className="gap-1 font-mono text-xs font-normal text-muted-foreground"
              >
                {counts[status]}
                <span className="sr-only"> {statusWord(place, status)}</span>
              </StatusIndicator>
            ))}
            {typed.trim() === "" ? null : (
              <span className="ml-auto truncate">
                {t("work.statusBar.showing", {
                  shown: shown.length,
                  total: inPlace.length,
                })}
              </span>
            )}
          </>
        }
      >
        {sideBySide || !selected ? (
          <ItemList
            place={place}
            shown={shown}
            typed={typed}
            onTyped={(next) => setFilter({ place, typed: next })}
            selectedId={selected?.id}
            onSelect={(id) => goTo(place, id)}
            onNew={() => setCreating(true)}
            className="w-full lg:w-96 lg:shrink-0 lg:border-r xl:w-md"
          />
        ) : null}
        {selected ? (
          <ItemDetail
            item={selected}
            onStatusChange={(status) =>
              changeWorkItemStatus(selected.id, status)
            }
            onBack={() => goTo(place)}
            className="flex-1"
          />
        ) : null}
      </AppShell>

      <NewItemDialog
        place={place}
        groups={groups}
        open={creating}
        onOpenChange={setCreating}
        onCreate={(draft) => {
          const created = createWorkItem(place, draft)
          setFilter({ place, typed: "" })
          goTo(place, created.id)
        }}
      />
    </>
  )
}
