import { describe, expect, it } from "vite-plus/test"

import { missingKeys } from "./missing-keys"

describe("the keys a translation is missing", () => {
  it("are none when it has every key of the source", () => {
    const source = { heading: "Log in", footer: "Ask for access." }
    const translation = { heading: "Entrar", footer: "Peça acesso." }

    expect(missingKeys(source, translation)).toEqual([])
  })

  it("name a key the source has and the translation lacks", () => {
    const source = { heading: "Log in", footer: "Ask for access." }
    const translation = { heading: "Entrar" }

    expect(missingKeys(source, translation)).toEqual(["footer"])
  })

  it("name a key inside a group by its whole path", () => {
    const source = { login: { code: { label: "Login code", wrong: "Wrong" } } }
    const translation = { login: { code: { label: "Código de acesso" } } }

    expect(missingKeys(source, translation)).toEqual(["login.code.wrong"])
  })

  it("name every key of a group the translation lacks", () => {
    const source = {
      heading: "Log in",
      status: { working: "In progress", done: "Delivered" },
    }
    const translation = { heading: "Entrar" }

    expect(missingKeys(source, translation)).toEqual([
      "status.working",
      "status.done",
    ])
  })

  it("name every key of a group that the translation has as one text", () => {
    const source = { status: { working: "In progress", done: "Delivered" } }
    const translation = { status: "Situação" }

    expect(missingKeys(source, translation)).toEqual([
      "status.working",
      "status.done",
    ])
  })

  it("name a text that the translation has as a group", () => {
    const source = { status: "Status" }
    const translation = { status: { working: "Em andamento" } }

    expect(missingKeys(source, translation)).toEqual(["status"])
  })

  it("leave out a key that only the translation has", () => {
    const source = { attempts_one: "One attempt", attempts_other: "Attempts" }
    const translation = {
      attempts_one: "Uma tentativa",
      attempts_many: "Tentativas",
      attempts_other: "Tentativas",
    }

    expect(missingKeys(source, translation)).toEqual([])
  })
})
