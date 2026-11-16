/**
 * The eight rules of DESIGN.md, each as a pair: what follows the rule and
 * what breaks it.
 *
 * A "Don't" example breaks its rule on purpose: it is the only place in the
 * prototype where a raw colour, a raw shadow or a long animation is right.
 */
import { Badge } from "@workspace/ui/components/badge"
import { Button } from "@workspace/ui/components/button"
import { Kbd } from "@workspace/ui/components/kbd"
import {
  Popover,
  PopoverContent,
  PopoverDescription,
  PopoverHeader,
  PopoverTitle,
  PopoverTrigger,
} from "@workspace/ui/components/popover"
import { StatusIndicator } from "@workspace/ui/components/status-indicator"
import { cn } from "@workspace/ui/lib/utils"
import { CheckIcon, PlusIcon, XIcon } from "lucide-react"
import type { ReactNode } from "react"

import { Section } from "./parts"

/** One side of a pair: the example, and one line on why it is right or wrong. */
function Example({
  kind,
  why,
  className,
  children,
}: {
  kind: "do" | "dont"
  why: string
  className?: string
  children: ReactNode
}) {
  const Sign = kind === "do" ? CheckIcon : XIcon

  return (
    <figure className="flex min-w-0 flex-col gap-2">
      <div
        className={cn(
          "flex min-h-36 flex-1 flex-wrap items-center justify-center gap-3 rounded-xl border bg-card p-4",
          className
        )}
      >
        {children}
      </div>
      <figcaption className="flex min-h-8 items-start gap-1.5 type-caption">
        <span
          className={cn(
            "flex shrink-0 items-center gap-1 font-medium",
            kind === "dont" && "text-destructive"
          )}
        >
          <Sign className="size-3.5" aria-hidden="true" />
          {kind === "do" ? "Do" : "Don't"}
        </span>
        <span className="text-muted-foreground">{why}</span>
      </figcaption>
    </figure>
  )
}

/** A rule: its number and its name as the title, the rule itself as the caption, and the pair. */
function Rule({
  number,
  name,
  rule,
  children,
}: {
  number: number
  name: string
  rule: string
  children: ReactNode
}) {
  return (
    <Section title={`${number}. ${name}`} note={rule}>
      <div className="grid gap-4 md:grid-cols-2">{children}</div>
    </Section>
  )
}

const requests = [
  { id: "REQ-0142", title: "Export leaves out archived projects" },
  { id: "REQ-0139", title: "New logo on the invoice" },
  { id: "REQ-0131", title: "Login by e-mail code" },
] as const

function RequestRows() {
  return (
    <>
      {requests.map(({ id, title }) => (
        <div
          key={id}
          className="flex h-7 items-center gap-2 border-b border-border-subtle px-2 last:border-b-0"
        >
          <span className="shrink-0 type-code text-muted-foreground">{id}</span>
          <span className="min-w-0 truncate">{title}</span>
        </div>
      ))}
    </>
  )
}

function Detail() {
  return (
    <div className="flex flex-col gap-1 p-3">
      <span className="font-medium">Export leaves out archived projects</span>
      <span className="type-caption text-muted-foreground">
        Opened 12 days ago by Ana Lima.
      </span>
    </div>
  )
}

function BordersRule() {
  return (
    <Rule
      number={1}
      name="Borders carry the structure"
      rule="Surfaces are separated by 1px lines. A shadow appears only on a floating surface: popover, dialog, command palette, toast."
    >
      <Example
        kind="do"
        why="Two panes side by side, flush, with one line between them."
        className="p-0"
      >
        <div className="grid w-full grid-cols-2 self-stretch">
          <div className="min-w-0 border-r">
            <RequestRows />
          </div>
          <div className="min-w-0">
            <Detail />
          </div>
        </div>
      </Example>
      <Example
        kind="dont"
        why="The same panes as cards with a gap and a shadow. The shadow says they float, and they do not."
        className="bg-background"
      >
        <div className="grid w-full grid-cols-2 gap-4">
          <div className="min-w-0 rounded-xl bg-card shadow-lg">
            <RequestRows />
          </div>
          <div className="min-w-0 rounded-xl bg-card shadow-lg">
            <Detail />
          </div>
        </div>
      </Example>
    </Rule>
  )
}

function AccentRule() {
  return (
    <Rule
      number={2}
      name="One accent, and colour means status"
      rule="The one accent colour is blue. Any other colour is a status colour with a fixed meaning: blue is working, amber needs attention, green is done, red is an error."
    >
      <Example
        kind="do"
        why="Tags are neutral. Colour appears on the primary action and on the status, where it means something."
        className="flex-col items-stretch"
      >
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="secondary">Billing</Badge>
          <Badge variant="secondary">Reports</Badge>
          <Badge variant="secondary">Login</Badge>
          <StatusIndicator status="attention" className="ml-auto shrink">
            Waiting for the client
          </StatusIndicator>
        </div>
        <Button className="self-start">
          <PlusIcon data-icon="inline-start" />
          New request
        </Button>
      </Example>
      <Example
        kind="dont"
        why="A colour for each tag and a green button. Amber and green no longer mean attention and done."
        className="flex-col items-stretch"
      >
        <div className="flex flex-wrap items-center gap-2">
          <Badge className="bg-purple-200 text-purple-950">Billing</Badge>
          <Badge className="bg-amber-200 text-amber-950">Reports</Badge>
          <Badge className="bg-emerald-200 text-emerald-950">Login</Badge>
          <StatusIndicator status="attention" className="ml-auto shrink">
            Waiting for the client
          </StatusIndicator>
        </div>
        <Button className="self-start bg-emerald-700 hover:bg-emerald-800">
          <PlusIcon data-icon="inline-start" />
          New request
        </Button>
      </Example>
    </Rule>
  )
}

function MonoRule() {
  return (
    <Rule
      number={3}
      name="Mono is for metadata"
      rule="Shortcuts, counts, paths and identifiers are set in the mono typeface. Running text is set in the interface typeface."
    >
      <Example
        kind="do"
        why="The identifier, the path, the count and the shortcut are mono. The sentence is not."
        className="flex-col items-start"
      >
        <div className="flex max-w-full items-center gap-2">
          <span className="shrink-0 type-code text-muted-foreground">
            REQ-0142
          </span>
          <span className="min-w-0 truncate">
            Export leaves out archived projects
          </span>
        </div>
        <p className="text-muted-foreground">
          The fix is in{" "}
          <code className="rounded-sm bg-muted px-1 type-code">
            apps/web/export
          </code>
          , with <span className="font-mono">3</span> files changed. Press{" "}
          <Kbd>Ctrl</Kbd> <Kbd>K</Kbd> to find it.
        </p>
      </Example>
      <Example
        kind="dont"
        why="The sentence is mono and the identifier is not. Nothing tells the reader which part is data."
        className="flex-col items-start"
      >
        <div className="flex max-w-full items-center gap-2">
          <span className="shrink-0 text-muted-foreground">REQ-0142</span>
          <span className="min-w-0 truncate font-mono">
            Export leaves out archived projects
          </span>
        </div>
        <p className="font-mono text-muted-foreground">
          The fix is in apps/web/export, with 3 files changed. Press Ctrl K to
          find it.
        </p>
      </Example>
    </Rule>
  )
}

function DecorationRule() {
  return (
    <Rule
      number={4}
      name="Decoration frames the page"
      rule="The grid texture and the diamond nodes belong to the page framing. The inside of a component is plain."
    >
      <Example
        kind="do"
        why="The texture is behind the pane and the nodes sit on its corners. The pane itself is plain."
        className="page-framing bg-background"
      >
        <div className="relative border bg-card px-6 py-4">
          <span className="font-medium">Log in</span>
          <span className="absolute top-0 left-0 -translate-x-1/2 -translate-y-1/2 frame-node" />
          <span className="absolute top-0 right-0 translate-x-1/2 -translate-y-1/2 frame-node" />
          <span className="absolute bottom-0 left-0 -translate-x-1/2 translate-y-1/2 frame-node" />
          <span className="absolute right-0 bottom-0 translate-x-1/2 translate-y-1/2 frame-node" />
        </div>
      </Example>
      <Example
        kind="dont"
        why="The texture fills the pane and the button, and a node decorates a title. The content has to fight it."
        className="bg-background"
      >
        <div className="flex flex-col items-start gap-3 border page-framing px-6 py-4">
          <span className="flex items-center gap-2 font-medium">
            <span className="frame-node" />
            Log in
          </span>
          <Button variant="outline" className="page-framing">
            Send login code
          </Button>
        </div>
      </Example>
    </Rule>
  )
}

function MotionRule() {
  return (
    <Rule
      number={5}
      name="Motion is short and plain"
      rule="A colour change takes 100 to 150ms. A control moves at most 1 to 2px. An overlay opens in 200 to 250ms."
    >
      <Example
        kind="do"
        why="The button changes colour at once under the pointer and moves 1px when pressed. The popover opens in 200ms."
      >
        <Button variant="secondary">Save draft</Button>
        <Popover>
          <PopoverTrigger render={<Button variant="outline" />}>
            Open the popover
          </PopoverTrigger>
          <PopoverContent className="w-56">
            <PopoverHeader>
              <PopoverTitle>Billing portal</PopoverTitle>
              <PopoverDescription>7 open requests.</PopoverDescription>
            </PopoverHeader>
          </PopoverContent>
        </Popover>
      </Example>
      <Example
        kind="dont"
        why="The badge never stops moving, and the button grows and turns slowly under the pointer. Motion that says nothing takes the eye from the work."
      >
        <Badge className="animate-bounce">New</Badge>
        <Button
          variant="secondary"
          className="duration-700 hover:scale-125 hover:-rotate-3"
        >
          Save draft
        </Button>
      </Example>
    </Rule>
  )
}

function VariablesRule() {
  return (
    <Rule
      number={6}
      name="Every value is a style variable"
      rule="A component takes each colour, radius, font, spacing step and shadow from a style variable, and hardcodes none."
    >
      <Example
        kind="do"
        why="Every value comes from a style variable, so the pane follows the theme. Change the theme in the prototype bar."
        className="flex-col bg-background"
      >
        <div className="flex flex-col gap-1 rounded-xl border bg-card px-4 py-3 text-card-foreground">
          <span className="font-medium">Billing portal</span>
          <span className="type-caption text-muted-foreground">
            7 open requests
          </span>
        </div>
        <code className="type-code text-muted-foreground">
          rounded-xl border bg-card
        </code>
      </Example>
      <Example
        kind="dont"
        why="The colours, the radius and the shadow are written as literals, so the pane stays white in the dark theme."
        className="flex-col bg-background"
      >
        <div className="flex flex-col gap-1 rounded-[10px] border border-[#d4d4d4] bg-[#ffffff] px-[18px] py-[11px] text-[#1a1a1a] shadow-[0_2px_6px_#00000033]">
          <span className="font-medium">Billing portal</span>
          <span className="type-caption text-[#6b6b6b]">7 open requests</span>
        </div>
        <code className="type-code text-muted-foreground">
          rounded-[10px] border-[#d4d4d4] bg-[#ffffff]
        </code>
      </Example>
    </Rule>
  )
}

function PrimaryRule() {
  return (
    <Rule
      number={7}
      name="One primary action per group"
      rule="In a group of actions, one button is primary. The others are secondary or ghost."
    >
      <Example
        kind="do"
        why="One primary button says what the group is for. The others step back."
      >
        <Button>Send for approval</Button>
        <Button variant="secondary">Save draft</Button>
        <Button variant="ghost">Cancel</Button>
      </Example>
      <Example
        kind="dont"
        why="Three primary buttons. Nothing says which one the person came here to press."
      >
        <Button>Send for approval</Button>
        <Button>Save draft</Button>
        <Button>Cancel</Button>
      </Example>
    </Rule>
  )
}

const statuses = [
  { status: "working", label: "In progress", dot: "bg-status-working" },
  {
    status: "attention",
    label: "Waiting for the client",
    dot: "bg-status-attention",
  },
  { status: "done", label: "Delivered", dot: "bg-status-done" },
  { status: "error", label: "Failed", dot: "bg-status-error" },
] as const

function StatusRule() {
  return (
    <Rule
      number={8}
      name="Status is colour plus a sign"
      rule="A status always shows an icon or a word together with its colour."
    >
      <Example
        kind="do"
        why="Each status has its sign and its word. It reads without the colour."
        className="flex-col items-start"
      >
        {statuses.map(({ status, label }, index) => (
          <div key={status} className="flex w-full items-center gap-3">
            <span className="min-w-0 flex-1 truncate">
              {requests[index % requests.length].title}
            </span>
            <StatusIndicator status={status} className="shrink">
              {label}
            </StatusIndicator>
          </div>
        ))}
      </Example>
      <Example
        kind="dont"
        why="A dot of colour and nothing else. Someone who does not tell red from green cannot read it, and nobody knows what amber means."
        className="flex-col items-start"
      >
        {statuses.map(({ status, dot }, index) => (
          <div key={status} className="flex w-full items-center gap-3">
            <span className="min-w-0 flex-1 truncate">
              {requests[index % requests.length].title}
            </span>
            <span className={cn("size-2 shrink-0 rounded-full", dot)} />
          </div>
        ))}
      </Example>
    </Rule>
  )
}

/** The eight pairs, in the order of the rules in DESIGN.md. */
export function Rules() {
  return (
    <>
      <BordersRule />
      <AccentRule />
      <MonoRule />
      <DecorationRule />
      <MotionRule />
      <VariablesRule />
      <PrimaryRule />
      <StatusRule />
    </>
  )
}
