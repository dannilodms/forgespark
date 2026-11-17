/**
 * The eight rules of DESIGN.md, each as a pair: what follows the rule and
 * what breaks it.
 *
 * A "Don't" example breaks its rule on purpose: it is the only place in the
 * prototype where a raw colour, a raw shadow or a long animation is right.
 *
 * The rules and the line under each example are explanatory text, in English.
 * What an example holds is sample content, in the language of the prototype.
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
import { Trans, useTranslation } from "react-i18next"

import { useFormat } from "#stores/language"

import { doAndDont } from "./explanations/do-and-dont"
import { freeText, sampleRequest, sampleRequests } from "./free-text"
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
          {doAndDont.caption[kind]}
        </span>
        <span className="text-muted-foreground">{why}</span>
      </figcaption>
    </figure>
  )
}

/** A rule: its number and its name as the title, the rule itself as the caption, and the pair. */
function Rule({
  number,
  text,
  children,
}: {
  number: number
  text: { name: string; rule: string }
  children: ReactNode
}) {
  return (
    <Section title={`${number}. ${text.name}`} note={text.rule}>
      <div className="grid gap-4 md:grid-cols-2">{children}</div>
    </Section>
  )
}

const requests = sampleRequests.slice(0, 3)

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
  const { t } = useTranslation()
  const { ago } = useFormat()

  return (
    <div className="flex flex-col gap-1 p-3">
      <span className="font-medium">{sampleRequest.title}</span>
      <span className="type-caption text-muted-foreground">
        {t("gallery.rules.opened", {
          when: ago(12, "day"),
          who: freeText.ana,
        })}
      </span>
    </div>
  )
}

function BordersRule() {
  const text = doAndDont.borders

  return (
    <Rule number={1} text={text}>
      <Example kind="do" why={text.do} className="p-0">
        <div className="grid w-full grid-cols-2 self-stretch">
          <div className="min-w-0 border-r">
            <RequestRows />
          </div>
          <div className="min-w-0">
            <Detail />
          </div>
        </div>
      </Example>
      <Example kind="dont" why={text.dont} className="bg-background">
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
  const { t } = useTranslation()
  const text = doAndDont.accent
  const billing = t("gallery.common.tags.billing")
  const reports = t("gallery.common.tags.reports")
  const login = t("gallery.common.tags.login")
  const waiting = t("gallery.common.status.attention")
  const newRequest = t("gallery.common.newRequest")

  return (
    <Rule number={2} text={text}>
      <Example kind="do" why={text.do} className="flex-col items-stretch">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="secondary">{billing}</Badge>
          <Badge variant="secondary">{reports}</Badge>
          <Badge variant="secondary">{login}</Badge>
          <StatusIndicator status="attention" className="ml-auto shrink">
            {waiting}
          </StatusIndicator>
        </div>
        <Button className="self-start">
          <PlusIcon data-icon="inline-start" />
          {newRequest}
        </Button>
      </Example>
      <Example kind="dont" why={text.dont} className="flex-col items-stretch">
        <div className="flex flex-wrap items-center gap-2">
          <Badge className="bg-purple-200 text-purple-950">{billing}</Badge>
          <Badge className="bg-amber-200 text-amber-950">{reports}</Badge>
          <Badge className="bg-emerald-200 text-emerald-950">{login}</Badge>
          <StatusIndicator status="attention" className="ml-auto shrink">
            {waiting}
          </StatusIndicator>
        </div>
        <Button className="self-start bg-emerald-700 hover:bg-emerald-800">
          <PlusIcon data-icon="inline-start" />
          {newRequest}
        </Button>
      </Example>
    </Rule>
  )
}

/** Where the sample fix is, and how many files it changed. */
const fix = { path: "apps/web/export", files: 3 }

function MonoRule() {
  const { number } = useFormat()
  const text = doAndDont.mono
  const values = { path: fix.path, files: number(fix.files) }

  return (
    <Rule number={3} text={text}>
      <Example kind="do" why={text.do} className="flex-col items-start">
        <div className="flex max-w-full items-center gap-2">
          <span className="shrink-0 type-code text-muted-foreground">
            {sampleRequest.id}
          </span>
          <span className="min-w-0 truncate">{sampleRequest.title}</span>
        </div>
        <p className="text-muted-foreground">
          <Trans
            i18nKey="gallery.rules.fix"
            values={values}
            components={{
              path: <code className="rounded-sm bg-muted px-1 type-code" />,
              mono: <span className="font-mono" />,
              kbd: <Kbd />,
            }}
          />
        </p>
      </Example>
      <Example kind="dont" why={text.dont} className="flex-col items-start">
        <div className="flex max-w-full items-center gap-2">
          <span className="shrink-0 text-muted-foreground">
            {sampleRequest.id}
          </span>
          <span className="min-w-0 truncate font-mono">
            {sampleRequest.title}
          </span>
        </div>
        <p className="font-mono text-muted-foreground">
          <Trans
            i18nKey="gallery.rules.fix"
            values={values}
            components={{ path: <span />, mono: <span />, kbd: <span /> }}
          />
        </p>
      </Example>
    </Rule>
  )
}

function DecorationRule() {
  const { t } = useTranslation()
  const text = doAndDont.decoration

  return (
    <Rule number={4} text={text}>
      <Example kind="do" why={text.do} className="page-framing bg-background">
        <div className="relative border bg-card px-6 py-4">
          <span className="font-medium">{t("login.heading")}</span>
          <span className="absolute top-0 left-0 -translate-x-1/2 -translate-y-1/2 frame-node" />
          <span className="absolute top-0 right-0 translate-x-1/2 -translate-y-1/2 frame-node" />
          <span className="absolute bottom-0 left-0 -translate-x-1/2 translate-y-1/2 frame-node" />
          <span className="absolute right-0 bottom-0 translate-x-1/2 translate-y-1/2 frame-node" />
        </div>
      </Example>
      <Example kind="dont" why={text.dont} className="bg-background">
        <div className="flex flex-col items-start gap-3 border page-framing px-6 py-4">
          <span className="flex items-center gap-2 font-medium">
            <span className="frame-node" />
            {t("login.heading")}
          </span>
          <Button variant="outline" className="page-framing">
            {t("login.sendCode")}
          </Button>
        </div>
      </Example>
    </Rule>
  )
}

function MotionRule() {
  const { t } = useTranslation()
  const text = doAndDont.motion
  const saveDraft = t("gallery.common.saveDraft")

  return (
    <Rule number={5} text={text}>
      <Example kind="do" why={text.do}>
        <Button variant="secondary">{saveDraft}</Button>
        <Popover>
          <PopoverTrigger render={<Button variant="outline" />}>
            {text.openPopover}
          </PopoverTrigger>
          <PopoverContent className="w-56">
            <PopoverHeader>
              <PopoverTitle>{freeText.project}</PopoverTitle>
              <PopoverDescription>
                {t("gallery.common.openRequestsCount", { count: 7 })}
              </PopoverDescription>
            </PopoverHeader>
          </PopoverContent>
        </Popover>
      </Example>
      <Example kind="dont" why={text.dont}>
        <Badge className="animate-bounce">{t("gallery.badge.new")}</Badge>
        <Button
          variant="secondary"
          className="duration-700 hover:scale-125 hover:-rotate-3"
        >
          {saveDraft}
        </Button>
      </Example>
    </Rule>
  )
}

function VariablesRule() {
  const { t } = useTranslation()
  const text = doAndDont.variables
  const openRequests = t("gallery.common.openRequestsCount", { count: 7 })

  return (
    <Rule number={6} text={text}>
      <Example kind="do" why={text.do} className="flex-col bg-background">
        <div className="flex flex-col gap-1 rounded-xl border bg-card px-4 py-3 text-card-foreground">
          <span className="font-medium">{freeText.project}</span>
          <span className="type-caption text-muted-foreground">
            {openRequests}
          </span>
        </div>
        <code className="type-code text-muted-foreground">
          rounded-xl border bg-card
        </code>
      </Example>
      <Example kind="dont" why={text.dont} className="flex-col bg-background">
        <div className="flex flex-col gap-1 rounded-[10px] border border-[#d4d4d4] bg-[#ffffff] px-[18px] py-[11px] text-[#1a1a1a] shadow-[0_2px_6px_#00000033]">
          <span className="font-medium">{freeText.project}</span>
          <span className="type-caption text-[#6b6b6b]">{openRequests}</span>
        </div>
        <code className="type-code text-muted-foreground">
          rounded-[10px] border-[#d4d4d4] bg-[#ffffff]
        </code>
      </Example>
    </Rule>
  )
}

function PrimaryRule() {
  const { t } = useTranslation()
  const text = doAndDont.primary
  const send = t("gallery.rules.sendForApproval")
  const saveDraft = t("gallery.common.saveDraft")
  const cancel = t("gallery.common.cancel")

  return (
    <Rule number={7} text={text}>
      <Example kind="do" why={text.do}>
        <Button>{send}</Button>
        <Button variant="secondary">{saveDraft}</Button>
        <Button variant="ghost">{cancel}</Button>
      </Example>
      <Example kind="dont" why={text.dont}>
        <Button>{send}</Button>
        <Button>{saveDraft}</Button>
        <Button>{cancel}</Button>
      </Example>
    </Rule>
  )
}

const dots = {
  working: "bg-status-working",
  attention: "bg-status-attention",
  done: "bg-status-done",
  error: "bg-status-error",
} as const

function StatusRule() {
  const { t } = useTranslation()
  const text = doAndDont.status

  return (
    <Rule number={8} text={text}>
      <Example kind="do" why={text.do} className="flex-col items-start">
        {sampleRequests.map(({ id, title, status }) => (
          <div key={id} className="flex w-full items-center gap-3">
            <span className="min-w-0 flex-1 truncate">{title}</span>
            <StatusIndicator status={status} className="shrink">
              {t(`gallery.common.status.${status}`)}
            </StatusIndicator>
          </div>
        ))}
      </Example>
      <Example kind="dont" why={text.dont} className="flex-col items-start">
        {sampleRequests.map(({ id, title, status }) => (
          <div key={id} className="flex w-full items-center gap-3">
            <span className="min-w-0 flex-1 truncate">{title}</span>
            <span
              className={cn("size-2 shrink-0 rounded-full", dots[status])}
            />
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
