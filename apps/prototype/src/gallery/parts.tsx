import { cn } from "@workspace/ui/lib/utils"
import type { ReactNode } from "react"

/** One route of the gallery: its title, one line on what it holds, and its sections. */
export function GalleryPage({
  title,
  lead,
  children,
}: {
  title: string
  lead: string
  children: ReactNode
}) {
  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-10 p-4 md:p-8">
      <header className="flex flex-col gap-2">
        <h1 className="type-display">{title}</h1>
        <p className="max-w-prose text-muted-foreground">{lead}</p>
      </header>
      {children}
    </div>
  )
}

/** A foundation, or the block of one component. */
export function Section({
  title,
  note,
  children,
}: {
  title: string
  note?: string
  children: ReactNode
}) {
  return (
    <section className="flex flex-col gap-5">
      <div className="flex flex-col gap-1 border-b pb-2">
        <h2 className="type-title">{title}</h2>
        {note ? (
          <p className="max-w-prose type-caption text-muted-foreground">
            {note}
          </p>
        ) : null}
      </div>
      {children}
    </section>
  )
}

/** One part of a section: the variants, the sizes, the states. */
export function Part({
  title,
  note,
  children,
}: {
  title: string
  note?: string
  children: ReactNode
}) {
  return (
    <div className="flex flex-col gap-2">
      <h3 className="type-label">{title}</h3>
      {note ? (
        <p className="max-w-prose type-caption text-muted-foreground">{note}</p>
      ) : null}
      {children}
    </div>
  )
}

/** The frame an example is shown in. */
export function Frame({
  className,
  children,
}: {
  className?: string
  children: ReactNode
}) {
  return (
    <div
      className={cn(
        "flex flex-wrap items-center gap-3 rounded-xl border bg-card p-4",
        className
      )}
    >
      {children}
    </div>
  )
}
