import type { ReactNode } from "react"

/**
 * The end of a step of a form in the framed pane, the setup's or the
 * login's, on a line above a border. In a row, as in the setup, the way back
 * to the step before, when there is one, is at the start of the line, and
 * the one primary action at its end; on a narrow screen the primary action
 * takes the rest of the line. Stacked, as in the login, whose pane is laid
 * out on its centre line, each takes the width of the pane, the primary
 * action above the way back.
 */
export function StepActions({
  back,
  stacked = false,
  children,
}: {
  back?: ReactNode
  stacked?: boolean
  children: ReactNode
}) {
  if (stacked) {
    return (
      <div className="flex flex-col gap-2 border-t border-border-subtle pt-4 *:w-full">
        {children}
        {back}
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2 border-t border-border-subtle pt-4">
      {back}
      <div className="ml-auto flex max-sm:flex-1 max-sm:*:flex-1">
        {children}
      </div>
    </div>
  )
}
