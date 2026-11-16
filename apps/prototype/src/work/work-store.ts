/*
 * The items of the sample work screen, held in memory. They start as the
 * sample data and are gone when the page reloads: there is no backend behind
 * the prototype.
 */
import type { Status } from "@workspace/ui/components/status-indicator"
import { useSyncExternalStore } from "react"

import { sampleItems, samplePerson } from "./sample-data"
import {
  type Draft,
  type Place,
  type WorkItem,
  addItem,
  changeStatus,
} from "./work-items"

let items = sampleItems

const listeners = new Set<() => void>()

function subscribe(onChange: () => void) {
  listeners.add(onChange)
  return () => listeners.delete(onChange)
}

function keep(next: WorkItem[]) {
  items = next
  for (const listener of listeners) {
    listener()
  }
}

/** Today in the time zone of the browser, as `2026-11-02`. */
function today() {
  const now = new Date()
  const month = String(now.getMonth() + 1).padStart(2, "0")
  const day = String(now.getDate()).padStart(2, "0")
  return `${now.getFullYear()}-${month}-${day}`
}

/** Every item, of both places, the most recent first. */
export function useWorkItems() {
  return useSyncExternalStore(subscribe, () => items)
}

/** Creates an item as the person who is logged in, and answers with it. */
export function createWorkItem(place: Place, draft: Draft) {
  const next = addItem(items, place, draft, samplePerson.name, today())
  keep(next)
  return next[0]
}

/** Changes the status of an item, as the person who is logged in. */
export function changeWorkItemStatus(id: string, status: Status) {
  keep(changeStatus(items, id, status, samplePerson.name, today()))
}
