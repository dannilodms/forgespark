import { describe, expect, it } from "vite-plus/test"

import {
  type WorkItem,
  addItem,
  changeStatus,
  checkDraft,
  countByStatus,
  filterItems,
} from "./work-items"

function item(overrides: Partial<WorkItem>): WorkItem {
  return {
    id: "REQ-0001",
    place: "requests",
    title: "Untitled",
    status: "working",
    owner: "Ana Lima",
    group: "Billing portal",
    updated: "2026-11-02",
    description: "",
    activity: [],
    ...overrides,
  }
}

const statusWord = (entry: WorkItem) =>
  ({
    working: "In progress",
    attention: "Waiting for the client",
    done: "Delivered",
    error: "Failed",
  })[entry.status]

describe("filtering the list", () => {
  const items = [
    item({ id: "REQ-0142", title: "Export leaves out archived projects" }),
    item({ id: "REQ-0139", title: "New logo on the invoice", status: "done" }),
  ]

  it("keeps every item while nothing is typed", () => {
    expect(filterItems(items, "  ", statusWord)).toEqual(items)
  })

  it("finds an item by a word of its title, whatever the case", () => {
    expect(filterItems(items, "LOGO", statusWord)).toEqual([items[1]])
  })

  it("finds an item by its identifier and by the word of its status", () => {
    expect(filterItems(items, "0142", statusWord)).toEqual([items[0]])
    expect(filterItems(items, "delivered", statusWord)).toEqual([items[1]])
  })

  it("asks for every word that was typed, in any order", () => {
    expect(filterItems(items, "progress export", statusWord)).toEqual([
      items[0],
    ])
    expect(filterItems(items, "export delivered", statusWord)).toEqual([])
  })

  it("ignores accents", () => {
    const request = item({ title: "Solicitação de alteração" })
    expect(filterItems([request], "solicitacao", statusWord)).toEqual([request])
  })
})

describe("checking a new item", () => {
  it("asks for a title", () => {
    expect(checkDraft({ title: "   ", group: "", description: "" })).toEqual({
      title: "required",
    })
  })

  it("refuses a title longer than a row can name", () => {
    const title = "a".repeat(121)
    expect(checkDraft({ title, group: "", description: "" })).toEqual({
      title: "tooLong",
    })
  })

  it("accepts a title alone", () => {
    expect(checkDraft({ title: "Export", group: "", description: "" })).toEqual(
      {}
    )
  })
})

describe("adding an item", () => {
  const items = [
    item({ id: "REQ-0142" }),
    item({ id: "PRJ-0007", place: "projects" }),
    item({ id: "REQ-0098" }),
  ]
  const draft = {
    title: "  Report of hours by project  ",
    group: "Billing portal",
    description: "Asked by the client in the last meeting.",
  }

  it("puts the new item first, with the next identifier of its place", () => {
    const added = addItem(items, "requests", draft, "Ana Lima", "2026-11-16")

    expect(added.map((entry) => entry.id)).toEqual([
      "REQ-0143",
      "REQ-0142",
      "PRJ-0007",
      "REQ-0098",
    ])
    expect(
      addItem(items, "projects", draft, "Ana Lima", "2026-11-16")[0]
    ).toEqual(expect.objectContaining({ id: "PRJ-0008", place: "projects" }))
  })

  it("opens the item as work in progress, by who created it", () => {
    const [created] = addItem(
      items,
      "requests",
      draft,
      "Ana Lima",
      "2026-11-16"
    )

    expect(created).toEqual({
      id: "REQ-0143",
      place: "requests",
      title: "Report of hours by project",
      status: "working",
      owner: "Ana Lima",
      group: "Billing portal",
      updated: "2026-11-16",
      description: "Asked by the client in the last meeting.",
      activity: [{ kind: "opened", at: "2026-11-16", who: "Ana Lima" }],
    })
  })

  it("numbers the first item of an empty place", () => {
    expect(
      addItem([], "projects", draft, "Ana Lima", "2026-11-16")[0]?.id
    ).toBe("PRJ-0001")
  })
})

describe("changing the status of an item", () => {
  const opened = { kind: "opened", at: "2026-11-02", who: "Ana Lima" } as const
  const items = [
    item({ id: "REQ-0142", activity: [opened] }),
    item({ id: "REQ-0139" }),
  ]

  it("changes that item alone, and records who did it and when", () => {
    const changed = changeStatus(
      items,
      "REQ-0142",
      "done",
      "Bruno Tavares",
      "2026-11-16"
    )

    expect(changed[0]).toEqual({
      ...items[0],
      status: "done",
      updated: "2026-11-16",
      activity: [
        {
          kind: "statusChanged",
          status: "done",
          at: "2026-11-16",
          who: "Bruno Tavares",
        },
        opened,
      ],
    })
    expect(changed[1]).toBe(items[1])
  })

  it("leaves an item alone when the status is the one it has", () => {
    expect(
      changeStatus(items, "REQ-0142", "working", "Bruno Tavares", "2026-11-16")
    ).toEqual(items)
  })
})

describe("counting by status", () => {
  it("counts the items of each status, with a zero where there is none", () => {
    const items = [
      item({ status: "working" }),
      item({ status: "error" }),
      item({ status: "working" }),
    ]

    expect(countByStatus(items)).toEqual({
      working: 2,
      attention: 0,
      done: 0,
      error: 1,
    })
  })
})
