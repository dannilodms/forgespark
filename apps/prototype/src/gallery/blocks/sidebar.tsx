import { Avatar, AvatarFallback } from "@workspace/ui/components/avatar"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSkeleton,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
} from "@workspace/ui/components/sidebar"
import { cn } from "@workspace/ui/lib/utils"
import {
  ClockIcon,
  FolderIcon,
  InboxIcon,
  SettingsIcon,
  UsersIcon,
} from "lucide-react"
import type { MouseEvent, ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { useFormat } from "#stores/language"

import { navigation } from "../explanations/navigation"
import { freeText } from "../free-text"
import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const text = navigation.sidebar

/** The sample places, each with the key of its name and, for some, a count. */
const places = [
  { name: "requests", icon: InboxIcon, count: 12 },
  { name: "projects", icon: FolderIcon, count: 4 },
  { name: "clients", icon: UsersIcon },
  { name: "hours", icon: ClockIcon },
] as const

/** An item of the gallery leads nowhere: it only shows how a link looks. */
function stay(event: MouseEvent) {
  event.preventDefault()
}

function Places({ tooltips = false }: { tooltips?: boolean }) {
  const { t } = useTranslation()
  const { number } = useFormat()

  return (
    <SidebarMenu>
      {places.map((place, index) => {
        const label = t(`gallery.common.places.${place.name}`)
        return (
          <SidebarMenuItem key={place.name}>
            <SidebarMenuButton
              isActive={index === 0}
              tooltip={tooltips ? label : undefined}
            >
              <place.icon />
              <span>{label}</span>
            </SidebarMenuButton>
            {"count" in place ? (
              <SidebarMenuBadge>{number(place.count)}</SidebarMenuBadge>
            ) : null}
          </SidebarMenuItem>
        )
      })}
    </SidebarMenu>
  )
}

function Person() {
  return (
    <div className="flex min-w-0 items-center gap-2 px-2 py-1">
      <Avatar size="sm">
        <AvatarFallback>{freeText.anaInitials}</AvatarFallback>
      </Avatar>
      <span className="min-w-0 truncate group-data-[collapsible=icon]:hidden">
        {freeText.ana}
      </span>
    </div>
  )
}

/** The frame a sidebar is previewed in: the sidebar and a stand-in for the page beside it. */
function Preview({
  className,
  children,
}: {
  className?: string
  children: ReactNode
}) {
  return (
    <SidebarProvider
      className={cn(
        "h-80 min-h-0 overflow-hidden rounded-xl border",
        className
      )}
    >
      {children}
    </SidebarProvider>
  )
}

function Page({ children }: { children?: ReactNode }) {
  return (
    <div className="flex min-w-0 flex-1 flex-col gap-2 bg-background p-3">
      {children}
      <span className="type-caption text-muted-foreground">{text.page}</span>
    </div>
  )
}

export function SidebarBlock() {
  const { t } = useTranslation()
  const { number } = useFormat()
  const work = t("gallery.sidebar.work")

  return (
    <Section {...text.about}>
      <Part {...text.typical}>
        <Preview>
          <Sidebar collapsible="none" className="border-r">
            <SidebarHeader>
              <span className="px-2 py-1 type-label">{freeText.product}</span>
            </SidebarHeader>
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupLabel>{work}</SidebarGroupLabel>
                <SidebarGroupContent>
                  <Places />
                </SidebarGroupContent>
              </SidebarGroup>
              <SidebarSeparator />
              <SidebarGroup>
                <SidebarGroupLabel>
                  {t("gallery.sidebar.setup")}
                </SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <SidebarMenuButton>
                        <SettingsIcon />
                        <span>{t("gallery.common.settings")}</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
              <Person />
            </SidebarFooter>
          </Sidebar>
          <Page />
        </Preview>
      </Part>

      <Part {...text.collapsing}>
        {/* The containment makes this frame the box the fixed sidebar is placed in. */}
        <Preview className="relative contain-layout">
          <Sidebar
            collapsible="icon"
            label={t("gallery.sidebar.label")}
            className="h-full"
          >
            <SidebarHeader>
              <span className="px-2 py-1 type-label group-data-[collapsible=icon]:hidden">
                {freeText.product}
              </span>
            </SidebarHeader>
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupLabel>{work}</SidebarGroupLabel>
                <SidebarGroupContent>
                  <Places tooltips />
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
              <Person />
            </SidebarFooter>
          </Sidebar>
          <Page>
            <div className="flex items-center gap-2">
              <SidebarTrigger
                variant="outline"
                aria-label={t("gallery.sidebar.toggle")}
              />
              <span>{t("gallery.sidebar.toggle")}</span>
            </div>
          </Page>
        </Preview>
      </Part>

      <Part {...text.states}>
        <Frame className="bg-sidebar">
          <SidebarProvider className="min-h-0">
            <StateMatrix
              rows={["item", "current item"] as const}
              states={["default", "hover", "focus", "disabled"]}
              render={(row, state) => (
                <SidebarMenu className="w-36">
                  <SidebarMenuItem>
                    <SidebarMenuButton
                      isActive={row === "current item"}
                      {...previewState(state)}
                    >
                      <InboxIcon />
                      <span>{t("gallery.common.places.requests")}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </SidebarMenu>
              )}
            />
          </SidebarProvider>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Preview>
          <Sidebar collapsible="none" className="border-r">
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupLabel>{work}</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <SidebarMenuButton isActive>
                        <InboxIcon />
                        <span>{t("gallery.sidebar.long")}</span>
                      </SidebarMenuButton>
                      <SidebarMenuBadge>{number(128)}</SidebarMenuBadge>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <SidebarMenuButton>
                        <FolderIcon />
                        <span>{t("gallery.common.places.projects")}</span>
                      </SidebarMenuButton>
                      <SidebarMenuSub>
                        <SidebarMenuSubItem>
                          <SidebarMenuSubButton
                            href="#sidebar"
                            onClick={stay}
                            isActive
                          >
                            <span>{freeText.shortProject}</span>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                        <SidebarMenuSubItem>
                          <SidebarMenuSubButton href="#sidebar" onClick={stay}>
                            <span>{freeText.otherLongProject}</span>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      </SidebarMenuSub>
                    </SidebarMenuItem>
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
              <SidebarGroup>
                <SidebarGroupLabel>
                  {t("gallery.sidebar.loading")}
                </SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <SidebarMenuSkeleton showIcon />
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <SidebarMenuSkeleton showIcon />
                    </SidebarMenuItem>
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
          </Sidebar>
          <Page />
        </Preview>
      </Part>
    </Section>
  )
}
