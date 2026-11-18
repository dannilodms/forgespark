import type { ReactNode } from "react"

/**
 * The end of a step of the setup: the way back to the step before, when
 * there is one, at the start of the line, and the one primary action at its
 * end. On a narrow screen the primary action takes the rest of the line.
 */
export function StepActions({
  back,
  children,
}: {
  back?: ReactNode
  children: ReactNode
}) {
  return (
    <div className="flex items-center gap-2 border-t border-border-subtle pt-4">
      {back}
      <div className="ml-auto flex max-sm:flex-1 max-sm:*:flex-1">
        {children}
      </div>
    </div>
  )
}
