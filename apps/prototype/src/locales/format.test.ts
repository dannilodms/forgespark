import { describe, expect, it } from "vite-plus/test"

import { formatFor } from "./format"

describe("a value written for a language", () => {
  it("is a number with the separators of the language", () => {
    expect(formatFor("en-US").number(1240.5, 1)).toBe("1,240.5")
    expect(formatFor("pt-BR").number(1240.5, 1)).toBe("1.240,5")
    expect(formatFor("es").number(41240.5, 1)).toBe("41.240,5")
  })

  it("is a number with the decimal places asked for, and none by default", () => {
    expect(formatFor("en-US").number(6, 1)).toBe("6.0")
    expect(formatFor("pt-BR").number(6, 1)).toBe("6,0")
    expect(formatFor("en-US").number(128)).toBe("128")
  })

  it("is a day with the month named in the language", () => {
    expect(formatFor("en-US").day("2026-11-02")).toBe("Nov 2, 2026")
    expect(formatFor("pt-BR").day("2026-11-02")).toBe("2 de nov. de 2026")
    expect(formatFor("es").day("2026-11-02")).toBe("2 nov 2026")
  })

  it("is how long ago something happened, in the words of the language", () => {
    expect(formatFor("en-US").ago(12, "day")).toBe("12 days ago")
    expect(formatFor("pt-BR").ago(12, "day")).toBe("há 12 dias")
    expect(formatFor("es").ago(12, "minute")).toBe("hace 12 minutos")
  })
})
