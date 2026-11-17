import { cn } from "@workspace/ui/lib/utils"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import type { GallerySection } from "./sections"

/**
 * One route of the gallery: its name, one line on what it holds, and its
 * sections. The name is the one the navigation shows, in the language of the
 * prototype. The line is explanatory text, in English.
 */
export function GalleryPage({
  section,
  lead,
  children,
}: {
  section: GallerySection
  lead: string
  children: ReactNode
}) {
  const { t } = useTranslation()

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-10 p-4 md:p-8">
      <header className="flex flex-col gap-2">
        <h1 className="type-display">
          {t(`gallery.navigation.sections.${section}`)}
        </h1>
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
