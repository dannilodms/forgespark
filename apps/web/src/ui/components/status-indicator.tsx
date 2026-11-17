import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "cn"
import {
  CircleCheckIcon,
  CircleDashedIcon,
  CircleXIcon,
  TriangleAlertIcon,
} from "lucide-react"

/**
 * The four meanings a status colour has (rule 2 of DESIGN.md). Which status
 * of a record shows which meaning is decided with each screen.
 */
type Status = "working" | "attention" | "done" | "error"

const signs = {
  working: CircleDashedIcon,
  attention: TriangleAlertIcon,
  done: CircleCheckIcon,
  error: CircleXIcon,
}

/*
 * The sign always takes the status colour, which reaches 3:1 on every fill a
 * row can have. The word takes it only in a badge, on the subtle background
 * of its colour: on a hovered or a selected row a status colour is below
 * 4.5:1, so a plain word is set in the text colour, at the weight of text.
 *
 * The colour of the sign is set on its stroke, not on the text colour, so a
 * menu item that recolours its content when it is highlighted leaves the
 * sign as it is.
 */
const statusIndicatorVariants = cva(
  "inline-flex max-w-full min-w-0 shrink-0 items-center gap-1.5 align-middle text-sm [&>svg]:pointer-events-none [&>svg]:size-3.5 [&>svg]:shrink-0",
  {
    variants: {
      status: {
        working: "[&>svg]:stroke-status-working",
        attention: "[&>svg]:stroke-status-attention",
        done: "[&>svg]:stroke-status-done",
        error: "[&>svg]:stroke-status-error",
      },
      variant: {
        plain: "text-foreground",
        badge: "h-5 gap-1 rounded-sm px-1.5 text-xs font-medium [&>svg]:size-3",
      },
    },
    compoundVariants: [
      {
        status: "working",
        variant: "badge",
        className: "bg-status-working-subtle text-status-working",
      },
      {
        status: "attention",
        variant: "badge",
        className: "bg-status-attention-subtle text-status-attention",
      },
      {
        status: "done",
        variant: "badge",
        className: "bg-status-done-subtle text-status-done",
      },
      {
        status: "error",
        variant: "badge",
        className: "bg-status-error-subtle text-status-error",
      },
    ],
    defaultVariants: {
      variant: "plain",
    },
  }
)

/**
 * A status: its colour, its sign and the word for it. The sign is fixed for
 * each meaning and the word is the children, so a status never shows as
 * colour alone (rule 8 of DESIGN.md). Plain, the sign carries the colour;
 * as a badge, the sign and the word both do.
 */
function StatusIndicator({
  status,
  variant = "plain",
  className,
  children,
  ...props
}: React.ComponentProps<"span"> & {
  status: Status
} & Pick<VariantProps<typeof statusIndicatorVariants>, "variant">) {
  const Sign = signs[status]

  return (
    <span
      data-slot="status-indicator"
      data-status={status}
      className={cn(statusIndicatorVariants({ status, variant }), className)}
      {...props}
    >
      <Sign aria-hidden="true" />
      <span className="truncate">{children}</span>
    </span>
  )
}

export { StatusIndicator, statusIndicatorVariants }
export type { Status }
