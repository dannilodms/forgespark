import { Avatar, AvatarFallback } from "@workspace/ui/components/avatar"
import { Button, ButtonShortcut } from "@workspace/ui/components/button"
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandInput,
  CommandList,
} from "@workspace/ui/components/command"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import {
  Sidebar,
  SidebarContent,
  SidebarProvider,
  SidebarTrigger,
} from "@workspace/ui/components/sidebar"
import { cn } from "@workspace/ui/lib/utils"
import { ChevronDownIcon, CircleUserRoundIcon, SearchIcon } from "lucide-react"
import { type ReactNode, useEffect, useState } from "react"

/*
 * A control in a bar is 24px high in 28px, so the bar would cut a focus
 * outline drawn around it. It is drawn inside the control.
 */
const focusInside = "focus-visible:-outline-offset-3"

/** The first letters of the first and the last name. */
function initials(name: string) {
  const words = name.trim().split(/\s+/)
  const first = words[0]?.[0] ?? ""
  const last = words.length > 1 ? (words.at(-1)?.[0] ?? "") : ""
  return (first + last).toUpperCase()
}

/**
 * The user menu of the app shell, at the end of the title bar: who is logged
 * in, and what that person does to the session. The children are the items
 * of the menu, and logging out is the last of them.
 */
export function UserMenu({
  name,
  email,
  label,
  children,
}: {
  name: string
  email: string
  /** Names the button for someone who does not see it. */
  label: string
  children: ReactNode
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            variant="ghost"
            size="xs"
            aria-label={label}
            className={cn("max-w-full min-w-0 text-foreground", focusInside)}
          />
        }
      >
        <CircleUserRoundIcon data-icon="inline-start" />
        <span className="hidden truncate sm:inline">{name}</span>
        <ChevronDownIcon data-icon="inline-end" />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-60">
        <div className="flex min-w-0 items-center gap-2 px-1.5 py-1">
          <Avatar>
            <AvatarFallback>{initials(name)}</AvatarFallback>
          </Avatar>
          <div className="flex min-w-0 flex-col">
            <span className="truncate text-sm font-medium">{name}</span>
            <span className="truncate type-caption text-muted-foreground">
              {email}
            </span>
          </div>
        </div>
        <DropdownMenuSeparator />
        {children}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

/** The texts of the command palette, and its commands. */
type CommandPalette = {
  /** The label of the button that opens the palette. */
  trigger: string
  /** The shortcut shown on the button. The palette opens with Ctrl K or ⌘ K. */
  shortcut: string
  placeholder: string
  /** Says what the palette is for, to someone who does not see it. */
  description: string
  /** What the palette says when nothing matches the search. */
  empty: string
  /** The groups of commands. `close` closes the palette once a command ran. */
  commands: (close: () => void) => ReactNode
}

/**
 * The app shell: a title bar with the user menu, the sidebar navigation, a
 * top bar with the breadcrumb and the command palette trigger, a main area
 * of flush panes, and a status bar. Every screen of the product sits in it.
 *
 * The shell takes the height it is given and never scrolls: each pane
 * scrolls inside itself. On a narrow screen the sidebar is a sheet, opened
 * by the button of the top bar.
 */
export function AppShell({
  product,
  userMenu,
  navigation,
  navigationLabel,
  sidebarToggleLabel,
  breadcrumb,
  commandPalette,
  statusBar,
  className,
  children,
}: {
  /** The name of the product, in the title bar. */
  product: ReactNode
  userMenu: ReactNode
  /** The groups of the sidebar. */
  navigation: ReactNode
  navigationLabel: string
  sidebarToggleLabel: string
  breadcrumb: ReactNode
  commandPalette: CommandPalette
  /** What the status bar holds. */
  statusBar: ReactNode
  className?: string
  /** The panes of the main area. They sit flush, split by a 1px line. */
  children: ReactNode
}) {
  const [paletteOpen, setPaletteOpen] = useState(false)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (
        event.key.toLowerCase() === "k" &&
        (event.ctrlKey || event.metaKey) &&
        !event.altKey
      ) {
        event.preventDefault()
        setPaletteOpen((open) => !open)
      }
    }
    window.addEventListener("keydown", onKeyDown)
    return () => window.removeEventListener("keydown", onKeyDown)
  }, [])

  return (
    <div className={cn("flex h-full min-h-0 flex-col", className)}>
      <header className="flex h-7 shrink-0 items-center justify-between gap-3 border-b bg-chrome pr-1 pl-3 type-caption text-muted-foreground">
        <span className="min-w-0 truncate">{product}</span>
        {userMenu}
      </header>

      {/* The containment makes this row the box the fixed sidebar is placed in. */}
      <SidebarProvider className="relative min-h-0 flex-1 contain-layout">
        <Sidebar collapsible="icon" label={navigationLabel} className="h-full">
          <SidebarContent>
            <nav aria-label={navigationLabel}>{navigation}</nav>
          </SidebarContent>
        </Sidebar>

        <div className="flex min-w-0 flex-1 flex-col">
          <div className="flex h-7 shrink-0 items-center gap-2 border-b bg-card pr-1 pl-0.5">
            <SidebarTrigger
              size="icon-xs"
              aria-label={sidebarToggleLabel}
              className={focusInside}
            />
            <div className="min-w-0 flex-1">{breadcrumb}</div>
            <Button
              variant="outline"
              size="xs"
              aria-label={commandPalette.trigger}
              className={cn("font-normal text-muted-foreground", focusInside)}
              onClick={() => setPaletteOpen(true)}
            >
              <SearchIcon data-icon="inline-start" />
              <span className="hidden md:inline">{commandPalette.trigger}</span>
              <ButtonShortcut className="hidden md:inline">
                {commandPalette.shortcut}
              </ButtonShortcut>
            </Button>
          </div>
          <main className="flex min-h-0 flex-1">{children}</main>
        </div>
      </SidebarProvider>

      <footer className="flex h-7 shrink-0 items-center gap-3 border-t bg-chrome px-3 type-caption text-muted-foreground">
        {statusBar}
      </footer>

      <CommandDialog
        open={paletteOpen}
        onOpenChange={setPaletteOpen}
        title={commandPalette.trigger}
        description={commandPalette.description}
      >
        <Command>
          <CommandInput placeholder={commandPalette.placeholder} />
          <CommandList>
            <CommandEmpty>{commandPalette.empty}</CommandEmpty>
            {commandPalette.commands(() => setPaletteOpen(false))}
          </CommandList>
        </Command>
      </CommandDialog>
    </div>
  )
}
