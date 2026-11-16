import { Badge } from "@workspace/ui/components/badge"
import { Button } from "@workspace/ui/components/button"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@workspace/ui/components/tabs"
import { cn } from "@workspace/ui/lib/utils"
import { ArrowLeftIcon } from "lucide-react"
import type { ReactNode } from "react"

export type DetailPaneTab = {
  value: string
  label: string
  /** How many things the tab holds, shown after its label in mono. */
  count?: number
  content: ReactNode
}

/**
 * The detail pane: a header with the title and the actions of the selected
 * item, and tabs that switch its content.
 *
 * The header names the item twice, by its identifier in mono and by its
 * title, and shows its status as a badge, since there the status stands
 * alone. Where the list and the detail do not fit side by side, under
 * 1024px, the screen shows one of them, and `back` leads to the list.
 */
export function DetailPane({
  label,
  identifier,
  title,
  status,
  actions,
  back,
  tabs,
  className,
}: {
  /** Names the pane for someone who does not see it. */
  label: string
  identifier: string
  title: string
  /** The status of the item: a status indicator as a badge. */
  status: ReactNode
  /** The actions on the item. At most one of them is primary. */
  actions?: ReactNode
  /** The way back to the list, shown only where the panes do not fit side by side. */
  back?: { label: string; onBack: () => void }
  tabs: DetailPaneTab[]
  className?: string
}) {
  return (
    <section
      aria-label={label}
      className={cn("flex min-h-0 min-w-0 flex-col bg-card", className)}
    >
      <header className="flex shrink-0 flex-col gap-2 px-4 pt-3 pb-2">
        <div className="flex min-h-7 flex-wrap items-center gap-2">
          {back ? (
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={back.label}
              className="-ml-2 lg:hidden"
              onClick={back.onBack}
            >
              <ArrowLeftIcon />
            </Button>
          ) : null}
          <span className="type-code text-muted-foreground">{identifier}</span>
          {status}
          {actions ? (
            <div className="ml-auto flex items-center gap-2">{actions}</div>
          ) : null}
        </div>
        <h2 className="type-title wrap-break-word">{title}</h2>
      </header>

      <Tabs defaultValue={tabs[0]?.value} className="min-h-0 flex-1 gap-0">
        <TabsList variant="line" className="w-full shrink-0 px-2">
          {tabs.map((tab) => (
            <TabsTrigger key={tab.value} value={tab.value}>
              {tab.label}
              {tab.count === undefined ? null : (
                <Badge variant="secondary" className="font-mono">
                  {tab.count}
                </Badge>
              )}
            </TabsTrigger>
          ))}
        </TabsList>
        {tabs.map((tab) => (
          <TabsContent
            key={tab.value}
            value={tab.value}
            className="min-h-0 overflow-y-auto rounded-none p-4 focus-visible:-outline-offset-3"
          >
            {tab.content}
          </TabsContent>
        ))}
      </Tabs>
    </section>
  )
}
