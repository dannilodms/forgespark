import type { Ref } from "react"

/**
 * The title of a step of the setup and what it asks for. The title takes the
 * focus when the step opens, so that it is read and the next Tab reaches the
 * first field: the button that led here is gone.
 */
export function StepHeading({
  ref,
  title,
  intro,
}: {
  ref: Ref<HTMLHeadingElement>
  title: string
  intro: string
}) {
  return (
    <header className="flex flex-col gap-1">
      <h2 ref={ref} tabIndex={-1} className="type-title outline-none">
        {title}
      </h2>
      <p className="text-muted-foreground">{intro}</p>
    </header>
  )
}
