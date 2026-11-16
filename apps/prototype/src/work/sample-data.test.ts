import { describe, expect, it } from "vite-plus/test"

import { sampleItems } from "./sample-data"
import { places, statuses } from "./work-items"

describe("the sample items", () => {
  it.each(places)("show the four statuses in %s", (place) => {
    const shown = new Set(
      sampleItems.filter((item) => item.place === place).map((i) => i.status)
    )

    expect([...shown].sort()).toEqual([...statuses].sort())
  })

  it("name each item once, with the letters of its place", () => {
    const identifiers = sampleItems.map((item) => item.id)

    expect(new Set(identifiers).size).toBe(identifiers.length)
    for (const item of sampleItems) {
      expect(item.id).toMatch(
        item.place === "requests" ? /^REQ-\d{4}$/ : /^PRJ-\d{4}$/
      )
    }
  })

  it("tell how each item came to its status", () => {
    for (const item of sampleItems) {
      const [latest] = item.activity
      const expected =
        item.status === "working"
          ? { kind: "opened" }
          : { kind: "statusChanged", status: item.status }

      expect(latest).toMatchObject(expected)
    }
  })
})
