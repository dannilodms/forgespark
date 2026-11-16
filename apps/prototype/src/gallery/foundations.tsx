import { cn } from "@workspace/ui/lib/utils"
import {
  CircleCheckIcon,
  FolderIcon,
  SearchIcon,
  TriangleAlertIcon,
} from "lucide-react"
import { useEffect, useRef, useState } from "react"

import { ColorPair } from "./color-pair"
import { Frame, GalleryPage, Part, Section } from "./parts"
import {
  controlSizes,
  easings,
  lines,
  motion,
  pairGroups,
  pinnedSizes,
  radiusRoles,
  shadows,
  spacingSteps,
  typeRoles,
  typefaces,
} from "./style-variables"
import { useStyleVariable } from "./use-style-variable"

function Value({ variable }: { variable: string }) {
  const value = useStyleVariable(variable)
  return <span className="type-code text-muted-foreground">{value}</span>
}

function Line({ variable, role }: { variable: string; role: string }) {
  return (
    <div className="flex items-center gap-3">
      <span
        className="h-8 w-14 shrink-0 rounded-lg border"
        style={{ backgroundColor: `var(--${variable})` }}
      />
      <div className="flex min-w-0 flex-col">
        <span className="type-code">{variable}</span>
        <span className="type-caption text-muted-foreground">{role}</span>
        <Value variable={variable} />
      </div>
    </div>
  )
}

/** One type role, with the size, line height and weight the browser computed for it. */
function TypeRow({
  token,
  use,
  sample,
}: {
  token: string
  use: string
  sample: string
}) {
  const ref = useRef<HTMLSpanElement>(null)
  const [spec, setSpec] = useState("")

  useEffect(() => {
    const element = ref.current
    if (!element) {
      return
    }
    const read = () => {
      const style = getComputedStyle(element)
      const px = (value: string) => `${Math.round(parseFloat(value))}px`
      const italic = style.fontStyle === "italic" ? " italic" : ""
      setSpec(
        `${px(style.fontSize)} / ${px(style.lineHeight)} / ${style.fontWeight}${italic}`
      )
    }
    read()
    window.addEventListener("resize", read)
    return () => window.removeEventListener("resize", read)
  }, [])

  return (
    <tr className="border-b align-baseline last:border-b-0">
      <td className="py-3 pr-4 type-code whitespace-nowrap">{token}</td>
      <td className="py-3 pr-4 type-code whitespace-nowrap text-muted-foreground">
        {spec}
      </td>
      <td className="py-3 pr-4 type-caption text-muted-foreground">{use}</td>
      <td className="py-3">
        <span ref={ref} className={cn("block", token)}>
          {sample}
        </span>
      </td>
    </tr>
  )
}

export function Foundations() {
  return (
    <GalleryPage
      title="Foundations"
      lead="Every style variable of the design system, with the value it has in the current theme. Colours are shown on the surface they are used on, with the contrast measured here."
    >
      {pairGroups.map((group) => (
        <Section key={group.title} title={group.title} note={group.note}>
          <div className="grid grid-cols-[repeat(auto-fill,minmax(13rem,1fr))] gap-3">
            {group.pairs.map((pair) => (
              <ColorPair key={pair.name} {...pair} />
            ))}
          </div>
        </Section>
      ))}

      <Section
        title="Lines and edges"
        note="Colours that separate or decorate. None of them carries text."
      >
        <div className="grid grid-cols-[repeat(auto-fill,minmax(16rem,1fr))] gap-4">
          {lines.map((line) => (
            <Line key={line.variable} {...line} />
          ))}
        </div>
      </Section>

      <Section
        title="Type"
        note="Three typefaces and seven roles. A role sets family, size, line height, weight and tracking together."
      >
        <div className="grid grid-cols-[repeat(auto-fill,minmax(16rem,1fr))] gap-4">
          {typefaces.map((face) => (
            <div key={face.variable} className="flex min-w-0 flex-col gap-1">
              <span className="type-label">{face.name}</span>
              <span
                className="truncate text-xl"
                style={{ fontFamily: `var(--${face.variable})` }}
              >
                {face.sample}
              </span>
              <span className="type-code text-muted-foreground">
                {face.variable}
              </span>
            </div>
          ))}
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[40rem] border-collapse text-left">
            <thead>
              <tr className="border-b type-label text-muted-foreground">
                <th className="py-2 pr-4 font-medium">Role</th>
                <th className="py-2 pr-4 font-medium">Size / line / weight</th>
                <th className="py-2 pr-4 font-medium">Use</th>
                <th className="py-2 font-medium">Specimen</th>
              </tr>
            </thead>
            <tbody>
              {typeRoles.map((role) => (
                <TypeRow key={role.token} {...role} />
              ))}
            </tbody>
          </table>
        </div>
        <p className="max-w-prose">
          A request moves from the first contact to delivery. Its history stays
          with the project, so whoever opens it later reads what was asked, what
          was agreed and what changed. See{" "}
          <a className="text-brand underline underline-offset-2" href="#type">
            how a request is written
          </a>{" "}
          or look for its identifier, such as{" "}
          <code className="rounded-sm bg-muted px-1 type-code">REQ-0142</code>.
        </p>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(16rem,1fr))] gap-4">
          {pinnedSizes.map((size) => (
            <div key={size.variable} className="flex flex-col">
              <span className="type-code">{size.variable}</span>
              <span className="type-caption text-muted-foreground">
                {size.role}
              </span>
              <Value variable={size.variable} />
            </div>
          ))}
        </div>
      </Section>

      <Section
        title="Spacing and control sizes"
        note="The base unit is 4px, and every spacing step is a multiple of it. Controls of the same kind share a height, so a form stays aligned."
      >
        <Part title="Spacing steps">
          <div className="flex flex-col gap-2">
            {spacingSteps.map((step) => (
              <div key={step.size} className="flex items-center gap-3">
                <span className="w-10 type-code text-muted-foreground">
                  {step.size}
                </span>
                <span className={cn("h-3 bg-brand", step.className)} />
              </div>
            ))}
          </div>
        </Part>
        <Part title="Control sizes">
          <div className="flex flex-wrap items-end gap-4">
            {controlSizes.map((control) => (
              <div key={control.size} className="flex flex-col gap-1">
                <span
                  className={cn(
                    "flex w-40 items-center rounded-lg border border-input px-2 type-ui",
                    control.className
                  )}
                >
                  {control.size}
                </span>
                <span className="w-40 type-caption text-muted-foreground">
                  {control.use}
                </span>
              </div>
            ))}
          </div>
        </Part>
      </Section>

      <Section
        title="Radius"
        note="Five roles. A corner is never rounder than its role."
      >
        <div className="grid grid-cols-[repeat(auto-fill,minmax(11rem,1fr))] gap-4">
          {radiusRoles.map((role) => (
            <div key={role.variable} className="flex flex-col gap-1">
              <span
                className="h-14 border border-input bg-card"
                style={{ borderRadius: `var(--${role.variable})` }}
              />
              <span className="type-code">{role.variable}</span>
              <Value variable={role.variable} />
              <span className="type-caption text-muted-foreground">
                {role.usedBy}
              </span>
            </div>
          ))}
        </div>
      </Section>

      <Section
        title="Elevation"
        note="Borders carry the structure. A shadow appears only on a floating surface."
      >
        <div className="grid grid-cols-[repeat(auto-fill,minmax(14rem,1fr))] gap-6">
          <div className="flex flex-col gap-2">
            <span className="flex h-20 items-center justify-center bg-background type-ui">
              Flat
            </span>
            <span className="type-caption text-muted-foreground">
              Page and chrome. No border and no shadow.
            </span>
          </div>
          <div className="flex flex-col gap-2">
            <span className="flex h-20 items-center justify-center rounded-xl border bg-card type-ui">
              Hairline
            </span>
            <span className="type-caption text-muted-foreground">
              Pane, card and control. A 1px border.
            </span>
          </div>
          <div className="flex flex-col gap-2">
            <span className="flex h-20 items-center justify-center rounded-2xl bg-popover type-ui shadow-floating">
              Floating
            </span>
            <span className="type-caption text-muted-foreground">
              Popover, dialog, command palette and toast.
            </span>
          </div>
        </div>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(16rem,1fr))] gap-4">
          {shadows.map((shadow) => (
            <div key={shadow.variable} className="flex flex-col">
              <span className="type-code">{shadow.variable}</span>
              <span className="type-caption text-muted-foreground">
                {shadow.usedBy}
              </span>
              <Value variable={shadow.variable} />
            </div>
          ))}
        </div>
      </Section>

      <Section
        title="Page framing"
        note="The grid texture and the diamond nodes frame a page. A page opts in with page-framing and frame-node; the inside of a component stays plain."
      >
        <div className="flex h-48 items-center justify-center rounded-xl border page-framing p-6">
          <div className="relative border bg-card px-8 py-5 type-ui">
            A pane on a framed page
            <span className="absolute top-0 left-0 -translate-x-1/2 -translate-y-1/2 frame-node" />
            <span className="absolute top-0 right-0 translate-x-1/2 -translate-y-1/2 frame-node" />
            <span className="absolute bottom-0 left-0 -translate-x-1/2 translate-y-1/2 frame-node" />
            <span className="absolute right-0 bottom-0 translate-x-1/2 translate-y-1/2 frame-node" />
          </div>
        </div>
      </Section>

      <Section title="Motion" note="Motion is short and plain.">
        <div className="overflow-x-auto">
          <table className="w-full min-w-[32rem] border-collapse text-left">
            <thead>
              <tr className="border-b type-label text-muted-foreground">
                <th className="py-2 pr-4 font-medium">What moves</th>
                <th className="py-2 pr-4 font-medium">Duration</th>
                <th className="py-2 font-medium">Easing</th>
              </tr>
            </thead>
            <tbody>
              {motion.map((row) => (
                <tr key={row.what} className="border-b last:border-b-0">
                  <td className="py-2 pr-4">{row.what}</td>
                  <td className="py-2 pr-4 type-code">{row.duration}</td>
                  <td className="py-2 type-code">{row.easing}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(16rem,1fr))] gap-4">
          {easings.map((easing) => (
            <div key={easing.variable} className="flex flex-col">
              <span className="type-code">{easing.variable}</span>
              <span className="type-caption text-muted-foreground">
                {easing.role}
              </span>
              <Value variable={easing.variable} />
            </div>
          ))}
        </div>
      </Section>

      <Section title="Icons" note="Lucide, at 12, 14 and 16px.">
        <Frame className="gap-8">
          {(["size-3", "size-3.5", "size-4"] as const).map((size) => (
            <div key={size} className="flex items-center gap-3">
              <SearchIcon className={size} />
              <FolderIcon className={size} />
              <CircleCheckIcon className={cn(size, "text-muted-foreground")} />
              <TriangleAlertIcon
                className={cn(size, "text-muted-foreground")}
              />
            </div>
          ))}
        </Frame>
      </Section>
    </GalleryPage>
  )
}
