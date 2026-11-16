import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@workspace/ui/components/input-group"
import {
  type Status,
  StatusIndicator,
} from "@workspace/ui/components/status-indicator"
import { cn } from "@workspace/ui/lib/utils"
import { SearchIcon, XIcon } from "lucide-react"
import type { KeyboardEvent, ReactNode } from "react"

export type ListPaneRow = {
  id: string
  /** What names the item, in mono: an identifier, a code. */
  identifier: string
  title: string
  status: Status
  /** The word the screen has for that status. */
  statusWord: string
}

/** Moves the focus between the rows with the arrow keys, Home and End. */
function moveBetweenRows(event: KeyboardEvent<HTMLUListElement>) {
  const rows = [
    ...event.currentTarget.querySelectorAll<HTMLButtonElement>(
      "[data-slot=list-pane-row]"
    ),
  ]
  const current = rows.indexOf(document.activeElement as HTMLButtonElement)
  if (current === -1) {
    return
  }
  const target = {
    ArrowDown: rows[current + 1],
    ArrowUp: rows[current - 1],
    Home: rows[0],
    End: rows.at(-1),
  }[event.key]
  if (target) {
    event.preventDefault()
    target.focus()
  }
}

/**
 * The list pane: a filter, a list of rows that each carry a status
 * indicator, and an empty state. Selecting a row drives the detail pane.
 *
 * A row is as high as a bar. Its status is plain: the sign carries the
 * colour and the word is text, so it reads on the hover wash and on the
 * selected fill. The focus outline of a row is drawn inside it, because rows
 * touch each other and the edge of the pane.
 */
export function ListPane({
  label,
  filter,
  action,
  rows,
  selectedId,
  onSelect,
  empty,
  className,
}: {
  /** Names the list for someone who does not see it. */
  label: string
  filter: {
    label: string
    placeholder: string
    clearLabel: string
    value: string
    onChange: (value: string) => void
  }
  /** The one action of the pane, beside the filter: creating an item. */
  action?: ReactNode
  rows: ListPaneRow[]
  selectedId?: string
  onSelect: (id: string) => void
  /** What the pane shows when it has no row: an Empty. */
  empty: ReactNode
  className?: string
}) {
  return (
    <section
      aria-label={label}
      className={cn("flex min-h-0 min-w-0 flex-col bg-card", className)}
    >
      <div className="flex shrink-0 items-center gap-2 border-b p-2">
        <InputGroup className="h-7 flex-1">
          <InputGroupAddon>
            <SearchIcon />
          </InputGroupAddon>
          <InputGroupInput
            aria-label={filter.label}
            placeholder={filter.placeholder}
            value={filter.value}
            onChange={(event) => filter.onChange(event.target.value)}
          />
          {filter.value === "" ? null : (
            <InputGroupAddon align="inline-end">
              <InputGroupButton
                size="icon-xs"
                aria-label={filter.clearLabel}
                onClick={() => filter.onChange("")}
              >
                <XIcon />
              </InputGroupButton>
            </InputGroupAddon>
          )}
        </InputGroup>
        {action}
      </div>

      {rows.length === 0 ? (
        <div className="flex min-h-0 flex-1 overflow-y-auto">{empty}</div>
      ) : (
        <ul
          className="min-h-0 flex-1 overflow-y-auto"
          onKeyDown={moveBetweenRows}
        >
          {rows.map((row) => (
            <li key={row.id} className="border-b border-border-subtle">
              <button
                type="button"
                data-slot="list-pane-row"
                aria-current={row.id === selectedId ? "true" : undefined}
                className="flex h-7 w-full items-center gap-3 px-3 text-left type-ui outline-none transition-colors hover:bg-accent focus-visible:outline-3 focus-visible:-outline-offset-3 focus-visible:outline-ring focus-visible:outline-solid aria-[current=true]:bg-selected"
                onClick={() => onSelect(row.id)}
              >
                <span className="shrink-0 type-code text-muted-foreground">
                  {row.identifier}
                </span>
                <span className="min-w-0 flex-1 truncate">{row.title}</span>
                <StatusIndicator
                  status={row.status}
                  className="max-w-1/2 shrink"
                >
                  {row.statusWord}
                </StatusIndicator>
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
