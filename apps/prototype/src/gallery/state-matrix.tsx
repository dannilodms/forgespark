import { cn } from "@workspace/ui/lib/utils"
import type { ReactNode } from "react"

const defaultStates = [
  "default",
  "hover",
  "focus",
  "active",
  "disabled",
] as const

/** A table of one row per variant and one column per state, each cell rendered as given. */
export function StateMatrix<Row extends string>({
  rows,
  states = defaultStates,
  render,
  className,
}: {
  rows: readonly Row[]
  states?: readonly string[]
  render: (row: Row, state: string) => ReactNode
  className?: string
}) {
  return (
    <div className={cn("w-full overflow-x-auto", className)}>
      <table className="border-separate border-spacing-x-3 border-spacing-y-3">
        <thead>
          <tr>
            <th />
            {states.map((state) => (
              <th
                key={state}
                scope="col"
                className="text-left type-code text-muted-foreground"
              >
                {state}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row}>
              <th
                scope="row"
                className="pr-2 text-left type-code text-muted-foreground"
              >
                {row}
              </th>
              {states.map((state) => (
                <td key={state}>{render(row, state)}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

/** Spread onto a component to pin a state without interaction. */
export function previewState(state: string) {
  if (state === "disabled") {
    return { disabled: true }
  }

  if (["hover", "focus", "active"].includes(state)) {
    return { "data-preview": state }
  }

  return {}
}
