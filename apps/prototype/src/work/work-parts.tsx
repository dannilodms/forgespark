/*
 * The three recipes filled with the sample items: the list of a place, the
 * detail of an item and the dialog that creates one. The sample work screen
 * puts them in the app shell, and the Recipes route of the gallery shows
 * each on its own.
 */
import { Button } from "@workspace/ui/components/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import {
  Field,
  FieldDescription,
  FieldLabel,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import {
  type Status,
  StatusIndicator,
} from "@workspace/ui/components/status-indicator"
import { Textarea } from "@workspace/ui/components/textarea"
import { ChevronDownIcon, PlusIcon, SearchXIcon } from "lucide-react"
import { useId, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { DetailPane } from "../recipes/detail-pane"
import { FormDialog, FormDialogError } from "../recipes/form-dialog"
import { ListPane } from "../recipes/list-pane"
import { useFormat } from "../stores/language"
import { useStatusWord } from "./status-word"
import {
  type Draft,
  type DraftErrors,
  type Place,
  type WorkItem,
  checkDraft,
  statuses,
  titleMaxLength,
} from "./work-items"

/** The list pane of a place: `shown` is what the filter left of its items. */
export function ItemList({
  place,
  shown,
  typed,
  onTyped,
  selectedId,
  onSelect,
  onNew,
  className,
}: {
  place: Place
  shown: WorkItem[]
  /** What is typed in the filter. */
  typed: string
  onTyped: (typed: string) => void
  selectedId?: string
  onSelect: (id: string) => void
  onNew: () => void
  className?: string
}) {
  const { t } = useTranslation()
  const statusWord = useStatusWord()

  return (
    <ListPane
      label={t(`work.places.${place}`)}
      filter={{
        label: t(`work.filter.label.${place}`),
        placeholder: t("work.filter.placeholder"),
        clearLabel: t("work.filter.clear"),
        value: typed,
        onChange: onTyped,
      }}
      action={
        <Button size="sm" onClick={onNew}>
          <PlusIcon data-icon="inline-start" />
          {t(`work.new.${place}`)}
        </Button>
      }
      rows={shown.map((item) => ({
        id: item.id,
        identifier: item.id,
        title: item.title,
        status: item.status,
        statusWord: statusWord(item.place, item.status),
      }))}
      selectedId={selectedId}
      onSelect={onSelect}
      empty={
        <Empty>
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <SearchXIcon />
            </EmptyMedia>
            <EmptyTitle>{t("work.noMatch.title")}</EmptyTitle>
            <EmptyDescription>
              {t("work.noMatch.description")}{" "}
              <span className="font-mono wrap-anywhere">{typed.trim()}</span>
            </EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            <Button variant="outline" onClick={() => onTyped("")}>
              {t("work.filter.clear")}
            </Button>
          </EmptyContent>
        </Empty>
      }
      className={className}
    />
  )
}

/** The menu that moves an item to another status. */
function StatusMenu({
  item,
  onStatusChange,
}: {
  item: WorkItem
  onStatusChange: (status: Status) => void
}) {
  const { t } = useTranslation()
  const statusWord = useStatusWord()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="outline" size="sm" />}>
        {t("work.detail.changeStatus")}
        <ChevronDownIcon data-icon="inline-end" />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-60">
        <DropdownMenuGroup>
          <DropdownMenuLabel>{t("work.detail.status")}</DropdownMenuLabel>
          <DropdownMenuRadioGroup
            value={item.status}
            onValueChange={(status: Status) => onStatusChange(status)}
          >
            {statuses.map((status) => (
              <DropdownMenuRadioItem key={status} value={status} closeOnClick>
                <StatusIndicator status={status} className="shrink">
                  {statusWord(item.place, status)}
                </StatusIndicator>
              </DropdownMenuRadioItem>
            ))}
          </DropdownMenuRadioGroup>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

/** The detail pane of an item: what it is, and what happened to it. */
export function ItemDetail({
  item,
  onStatusChange,
  onBack,
  className,
}: {
  item: WorkItem
  onStatusChange: (status: Status) => void
  /** Leads back to the list where the two panes do not fit side by side. */
  onBack?: () => void
  className?: string
}) {
  const { t } = useTranslation()
  const statusWord = useStatusWord()
  const { day, number } = useFormat()

  const facts = [
    { label: t("work.detail.owner"), value: item.owner },
    { label: t(`work.detail.group.${item.place}`), value: item.group },
    { label: t("work.detail.updated"), value: day(item.updated) },
  ]

  return (
    <DetailPane
      label={t("work.detail.label", { id: item.id })}
      identifier={item.id}
      title={item.title}
      status={
        <StatusIndicator status={item.status} variant="badge">
          {statusWord(item.place, item.status)}
        </StatusIndicator>
      }
      actions={<StatusMenu item={item} onStatusChange={onStatusChange} />}
      back={onBack ? { label: t("work.detail.back"), onBack } : undefined}
      tabs={[
        {
          value: "details",
          label: t("work.detail.tabs.details"),
          content: (
            <div className="flex max-w-prose flex-col gap-4">
              {item.description === "" ? (
                <p className="text-muted-foreground">
                  {t("work.detail.noDescription")}
                </p>
              ) : (
                <p className="wrap-break-word">{item.description}</p>
              )}
              <dl className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 border-t pt-4">
                {facts.map((fact) => (
                  <div key={fact.label} className="contents">
                    <dt className="type-label text-muted-foreground">
                      {fact.label}
                    </dt>
                    <dd className="min-w-0 wrap-break-word">{fact.value}</dd>
                  </div>
                ))}
              </dl>
            </div>
          ),
        },
        {
          value: "activity",
          label: t("work.detail.tabs.activity"),
          count: number(item.activity.length),
          content: (
            <ol className="flex max-w-prose flex-col">
              {item.activity.map((entry) => (
                <li
                  key={`${entry.at} ${entry.kind} ${entry.kind === "statusChanged" ? entry.status : ""}`}
                  className="flex flex-wrap gap-x-3 border-b border-border-subtle py-2 first:pt-0 last:border-b-0"
                >
                  <time
                    dateTime={entry.at}
                    className="shrink-0 type-code leading-5 text-muted-foreground"
                  >
                    {day(entry.at)}
                  </time>
                  <span className="min-w-0">
                    {entry.kind === "opened"
                      ? t("work.activity.opened", { who: entry.who })
                      : t("work.activity.statusChanged", {
                          who: entry.who,
                          status: statusWord(item.place, entry.status),
                        })}
                  </span>
                </li>
              ))}
            </ol>
          ),
        },
      ]}
      className={className}
    />
  )
}

/**
 * The form dialog that creates an item of a place. `groups` is what the item
 * can belong to: the projects for a request, the clients for a project.
 */
export function NewItemDialog({
  place,
  groups,
  open,
  onOpenChange,
  onCreate,
}: {
  place: Place
  groups: string[]
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreate: (draft: Draft) => void
}) {
  const { t } = useTranslation()
  const { number } = useFormat()
  const id = useId()
  const titleField = useRef<HTMLInputElement>(null)
  const [title, setTitle] = useState("")
  const [chosenGroup, setChosenGroup] = useState<string | null>(null)
  const [description, setDescription] = useState("")
  const [errors, setErrors] = useState<DraftErrors>({})

  const group = chosenGroup ?? groups[0] ?? ""

  /** A dialog that closes keeps nothing of what was typed. */
  function changeOpen(next: boolean) {
    onOpenChange(next)
    if (!next) {
      setTitle("")
      setChosenGroup(null)
      setDescription("")
      setErrors({})
    }
  }

  function submit() {
    const draft = { title, group, description }
    const found = checkDraft(draft)
    setErrors(found)
    if (found.title) {
      titleField.current?.focus()
      return
    }
    onCreate(draft)
    changeOpen(false)
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={changeOpen}
      title={t(`work.form.heading.${place}`)}
      description={t("work.form.intro")}
      submitLabel={t(`work.form.submit.${place}`)}
      cancelLabel={t("work.form.cancel")}
      closeLabel={t("work.form.close")}
      onSubmit={submit}
    >
      <Field data-invalid={errors.title ? true : undefined}>
        <FieldLabel htmlFor={`${id}-title`}>
          {t("work.form.titleLabel")}
        </FieldLabel>
        <Input
          ref={titleField}
          id={`${id}-title`}
          value={title}
          aria-invalid={errors.title ? true : undefined}
          aria-describedby={errors.title ? `${id}-title-error` : undefined}
          onChange={(event) => {
            setTitle(event.target.value)
            setErrors({})
          }}
        />
        {errors.title ? (
          <FormDialogError id={`${id}-title-error`}>
            {t(`work.form.errors.${errors.title}`, {
              max: number(titleMaxLength),
            })}
          </FormDialogError>
        ) : null}
      </Field>
      <Field>
        <FieldLabel htmlFor={`${id}-group`}>
          {t(`work.detail.group.${place}`)}
        </FieldLabel>
        <Select
          items={groups.map((name) => ({ value: name, label: name }))}
          value={group}
          onValueChange={(value) => setChosenGroup(value)}
        >
          <SelectTrigger id={`${id}-group`} className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              {groups.map((name) => (
                <SelectItem key={name} value={name}>
                  {name}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field>
      <Field>
        <FieldLabel htmlFor={`${id}-description`}>
          {t("work.form.descriptionLabel")}
        </FieldLabel>
        <Textarea
          id={`${id}-description`}
          value={description}
          aria-describedby={`${id}-description-hint`}
          onChange={(event) => setDescription(event.target.value)}
        />
        <FieldDescription id={`${id}-description-hint`}>
          {t("work.form.descriptionHint")}
        </FieldDescription>
      </Field>
    </FormDialog>
  )
}
