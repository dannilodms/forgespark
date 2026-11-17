import { cn } from "@workspace/ui/lib/utils"
import { CheckIcon, CircleCheckIcon, TriangleAlertIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"

import { foundations } from "./explanations/foundations"

const text = foundations.contrast

type Rgb = [number, number, number]

// Resolves any CSS colour to sRGB by painting one pixel. A translucent colour
// is composited over the backdrop, as the browser does.
function toRgb(color: string, backdrop: string): Rgb | null {
  const canvas = document.createElement("canvas")
  canvas.width = canvas.height = 1
  const context = canvas.getContext("2d", { willReadFrequently: true })
  if (!context) {
    return null
  }
  context.fillStyle = backdrop
  context.fillRect(0, 0, 1, 1)
  context.fillStyle = color
  context.fillRect(0, 0, 1, 1)
  const [r = 0, g = 0, b = 0] = context.getImageData(0, 0, 1, 1).data
  return [r, g, b]
}

function toHex(rgb: Rgb) {
  return `#${rgb.map((value) => value.toString(16).padStart(2, "0")).join("")}`
}

function luminance(rgb: Rgb) {
  const [r, g, b] = rgb.map((value) => {
    const channel = value / 255
    return channel <= 0.03928
      ? channel / 12.92
      : ((channel + 0.055) / 1.055) ** 2.4
  }) as Rgb
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

function contrast(a: Rgb, b: Rgb) {
  const [light = 0, dark = 0] = [luminance(a), luminance(b)].sort(
    (x, y) => y - x
  )
  return (light + 0.05) / (dark + 0.05)
}

/**
 * A foreground on a surface, with the contrast measured in the browser and
 * the level it has to reach: 4.5:1 for text, 3:1 for the border of a control,
 * for the focus indicator and for the sign of a status. A line between
 * surfaces has no level to reach (`minimum` is `null`): its contrast is
 * measured and shown all the same. The measure is taken again when the theme
 * changes.
 */
export function ColorPair({
  name,
  background,
  foreground,
  minimum = 4.5,
  kind = "text",
}: {
  name: string
  background: string
  foreground: string
  minimum?: 3 | 4.5 | null
  kind?: "text" | "line" | "sign"
}) {
  const ref = useRef<HTMLDivElement>(null)
  const [measure, setMeasure] = useState<{
    ratio: number
    surface: string
  } | null>(null)

  useEffect(() => {
    const element = ref.current
    if (!element) {
      return
    }
    const read = () => {
      const style = getComputedStyle(element)
      const page = getComputedStyle(document.body).backgroundColor
      const surface = toRgb(style.backgroundColor, page)
      if (!surface) {
        return
      }
      const ink = toRgb(style.color, toHex(surface))
      if (ink) {
        setMeasure({ ratio: contrast(surface, ink), surface: toHex(surface) })
      }
    }
    read()
    const observer = new MutationObserver(read)
    observer.observe(document.documentElement, { attributes: true })
    return () => observer.disconnect()
  }, [])

  const passes = measure && minimum ? measure.ratio >= minimum : true
  const level = !measure
    ? ""
    : minimum === null
      ? text.noLevel
      : !passes
        ? `${text.fails} ${minimum}:1`
        : minimum === 3
          ? `${text.passes} 3:1`
          : measure.ratio >= 7
            ? "AAA"
            : "AA"

  return (
    <div
      data-contrast={measure ? (passes ? "pass" : "fail") : undefined}
      className="flex flex-col overflow-hidden rounded-xl border"
    >
      <div
        ref={ref}
        style={{
          backgroundColor: `var(--${background})`,
          color: `var(--${foreground})`,
        }}
        className="flex h-14 items-center px-3"
      >
        {kind === "text" ? (
          <span className="type-title">Aa</span>
        ) : kind === "sign" ? (
          <CircleCheckIcon className="size-4" />
        ) : (
          <span className="h-7 w-14 rounded-lg border-2 border-current" />
        )}
      </div>
      <div className="flex flex-1 flex-col gap-0.5 border-t bg-card p-3">
        <span className="type-label">{name}</span>
        <span className="type-code text-muted-foreground">
          {foreground} {text.on} {background}
        </span>
        {measure ? (
          <span
            className={cn(
              "flex items-center gap-1 type-code",
              passes ? "text-muted-foreground" : "text-destructive"
            )}
          >
            {passes ? (
              <CheckIcon className="size-3" />
            ) : (
              <TriangleAlertIcon className="size-3" />
            )}
            {measure.ratio.toFixed(2)}:1, {level}
          </span>
        ) : null}
      </div>
    </div>
  )
}
