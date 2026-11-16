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

const statusIndicatorVariants = cva(
  "inline-flex max-w-full min-w-0 shrink-0 items-center gap-1.5 text-sm font-medium [&>svg]:pointer-events-none [&>svg]:size-3.5 [&>svg]:shrink-0",
  {
    variants: {
      status: {
        working: "text-status-working",
        attention: "text-status-attention",
        done: "text-status-done",
        error: "text-status-error",
      },
      variant: {
        plain: "",
        badge: "h-5 gap-1 rounded-sm px-1.5 text-xs [&>svg]:size-3",
      },
    },
    compoundVariants: [
      {
        status: "working",
        variant: "badge",
        className: "bg-status-working-subtle",
      },
      {
        status: "attention",
        variant: "badge",
        className: "bg-status-attention-subtle",
      },
      { status: "done", variant: "badge", className: "bg-status-done-subtle" },
      {
        status: "error",
        variant: "badge",
        className: "bg-status-error-subtle",
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
 * colour alone (rule 8 of DESIGN.md).
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
