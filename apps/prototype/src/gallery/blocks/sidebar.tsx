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

import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const places = [
  { label: "Requests", icon: InboxIcon, count: "12" },
  { label: "Projects", icon: FolderIcon, count: "4" },
  { label: "Clients", icon: UsersIcon },
  { label: "Hours", icon: ClockIcon },
] as const

/** An item of the gallery leads nowhere: it only shows how a link looks. */
function stay(event: MouseEvent) {
  event.preventDefault()
}

function Places({ tooltips = false }: { tooltips?: boolean }) {
  return (
    <SidebarMenu>
      {places.map((place, index) => (
        <SidebarMenuItem key={place.label}>
          <SidebarMenuButton
            isActive={index === 0}
            tooltip={tooltips ? place.label : undefined}
          >
            <place.icon />
            <span>{place.label}</span>
          </SidebarMenuButton>
          {"count" in place ? (
            <SidebarMenuBadge>{place.count}</SidebarMenuBadge>
          ) : null}
        </SidebarMenuItem>
      ))}
    </SidebarMenu>
  )
}

function Person() {
  return (
    <div className="flex min-w-0 items-center gap-2 px-2 py-1">
      <Avatar size="sm">
        <AvatarFallback>AL</AvatarFallback>
      </Avatar>
      <span className="min-w-0 truncate group-data-[collapsible=icon]:hidden">
        Ana Lima
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
      <span className="type-caption text-muted-foreground">
        The page beside the sidebar.
      </span>
    </div>
  )
}

export function SidebarBlock() {
  return (
    <Section
      title="Sidebar"
      note="The navigation of the app: the places a person can go, in groups. It sits flush against the page, split from it by a line."
    >
      <Part
        title="Typical"
        note="The current place takes the selected fill. A count sits at the end of its item, in mono."
      >
        <Preview>
          <Sidebar collapsible="none" className="border-r">
            <SidebarHeader>
              <span className="px-2 py-1 type-label">Software Factory</span>
            </SidebarHeader>
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupLabel>Work</SidebarGroupLabel>
                <SidebarGroupContent>
                  <Places />
                </SidebarGroupContent>
              </SidebarGroup>
              <SidebarSeparator />
              <SidebarGroup>
                <SidebarGroupLabel>Setup</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <SidebarMenuButton>
                        <SettingsIcon />
                        <span>Settings</span>
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

      <Part
        title="Collapsing"
        note="The button, or Ctrl B, collapses the sidebar to its icons, and each icon then names its place in a tooltip. On a narrow screen the sidebar is a sheet that the same button opens."
      >
        {/* The containment makes this frame the box the fixed sidebar is placed in. */}
        <Preview className="relative contain-layout">
          <Sidebar collapsible="icon" className="h-full">
            <SidebarHeader>
              <span className="px-2 py-1 type-label group-data-[collapsible=icon]:hidden">
                Software Factory
              </span>
            </SidebarHeader>
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupLabel>Work</SidebarGroupLabel>
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
                aria-label="Collapse or expand the sidebar"
              />
              <span>Collapse or expand the sidebar</span>
            </div>
          </Page>
        </Preview>
      </Part>

      <Part
        title="States of an item"
        note="Each state is pinned, so all of them are visible at once."
      >
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
                      <span>Requests</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </SidebarMenu>
              )}
            />
          </SidebarProvider>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long label is cut with an ellipsis and never widens the sidebar. An item can hold a second level. While the places load, each is a skeleton."
      >
        <Preview>
          <Sidebar collapsible="none" className="border-r">
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupLabel>Trabalho</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <SidebarMenuButton isActive>
                        <InboxIcon />
                        <span>
                          Solicitações que aguardam a aprovação do cliente
                        </span>
                      </SidebarMenuButton>
                      <SidebarMenuBadge>128</SidebarMenuBadge>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <SidebarMenuButton>
                        <FolderIcon />
                        <span>Projetos</span>
                      </SidebarMenuButton>
                      <SidebarMenuSub>
                        <SidebarMenuSubItem>
                          <SidebarMenuSubButton
                            href="#sidebar"
                            onClick={stay}
                            isActive
                          >
                            <span>Portal de cobrança</span>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                        <SidebarMenuSubItem>
                          <SidebarMenuSubButton href="#sidebar" onClick={stay}>
                            <span>
                              Aplicativo de acompanhamento das entregas
                            </span>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      </SidebarMenuSub>
                    </SidebarMenuItem>
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
              <SidebarGroup>
                <SidebarGroupLabel>Loading</SidebarGroupLabel>
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
