/*
 * What the sample work screen does to its items, as plain functions: no
 * React and no stored state. The items themselves are sample data (see
 * `sample-data.ts`) and define nothing about the product.
 */
import type { Status } from "@workspace/ui/components/status-indicator"

/** The places of the sample work screen, in the order of its navigation. */
export const places = ["requests", "projects"] as const

export type Place = (typeof places)[number]

export type ActivityEntry = {
  /** The day, as `2026-11-02`. */
  at: string
  who: string
} & ({ kind: "opened" } | { kind: "statusChanged"; status: Status })

export type WorkItem = {
  /** What a person reads and types to name the item, such as `REQ-0142`. */
  id: string
  place: Place
  title: string
  status: Status
  owner: string
  /** What the item belongs to: the project of a request, the client of a project. */
  group: string
  /** The day of the last change, as `2026-11-02`. */
  updated: string
  description: string
  /** The most recent entry first. */
  activity: ActivityEntry[]
}

/** Lower case and without accents, so "solicitacao" finds "Solicitação". */
function plain(text: string) {
  return text
    .normalize("NFD")
    .replace(/\p{Diacritic}/gu, "")
    .toLowerCase()
}

/**
 * The items that match what was typed in the filter: every word has to be in
 * the identifier, the title, what the item belongs to, its owner or the word
 * of its status. `statusWord` gives that word as the screen shows it, in the
 * current language.
 */
export function filterItems(
  items: WorkItem[],
  typed: string,
  statusWord: (item: WorkItem) => string
) {
  const wanted = plain(typed).split(/\s+/).filter(Boolean)
  if (wanted.length === 0) {
    return items
  }
  return items.filter((item) => {
    const text = plain(
      [item.id, item.title, item.group, item.owner, statusWord(item)].join(" ")
    )
    return wanted.every((word) => text.includes(word))
  })
}

/** What the form dialog collects for a new item. */
export type Draft = { title: string; group: string; description: string }

export const titleMaxLength = 120

/** What is wrong with a draft, by field. An empty answer means nothing is. */
export type DraftErrors = { title?: "required" | "tooLong" }

export function checkDraft(draft: Draft): DraftErrors {
  const title = draft.title.trim()
  if (title === "") {
    return { title: "required" }
  }
  if (title.length > titleMaxLength) {
    return { title: "tooLong" }
  }
  return {}
}

/** The letters that open the identifier of an item of each place. */
const identifierPrefix: Record<Place, string> = {
  requests: "REQ",
  projects: "PRJ",
}

function nextIdentifier(items: WorkItem[], place: Place) {
  const prefix = identifierPrefix[place]
  const highest = items
    .filter((item) => item.place === place)
    .reduce((top, item) => Math.max(top, Number(item.id.split("-")[1])), 0)
  return `${prefix}-${String(highest + 1).padStart(4, "0")}`
}

/**
 * The list with a new item in front. The item opens as work in progress,
 * owned by who created it. `today` is the day, as `2026-11-02`.
 */
export function addItem(
  items: WorkItem[],
  place: Place,
  draft: Draft,
  by: string,
  today: string
): WorkItem[] {
  const created: WorkItem = {
    id: nextIdentifier(items, place),
    place,
    title: draft.title.trim(),
    status: "working",
    owner: by,
    group: draft.group,
    updated: today,
    description: draft.description.trim(),
    activity: [{ kind: "opened", at: today, who: by }],
  }
  return [created, ...items]
}

/** The list with one item in another status. `today` is the day, as `2026-11-02`. */
export function changeStatus(
  items: WorkItem[],
  id: string,
  status: Status,
  by: string,
  today: string
): WorkItem[] {
  return items.map((item) =>
    item.id === id && item.status !== status
      ? {
          ...item,
          status,
          updated: today,
          activity: [
            { kind: "statusChanged", status, at: today, who: by },
            ...item.activity,
          ],
        }
      : item
  )
}

/** The four meanings a status has, in the order the screen shows them. */
export const statuses = ["working", "attention", "done", "error"] as const

/** How many items are in each status. */
export function countByStatus(items: WorkItem[]): Record<Status, number> {
  const counts = { working: 0, attention: 0, done: 0, error: 0 }
  for (const item of items) {
    counts[item.status] += 1
  }
  return counts
}
