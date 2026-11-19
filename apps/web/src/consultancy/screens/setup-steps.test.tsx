import { HttpResponse, http } from "msw"
import { describe, expect, it, onTestFinished, vi } from "vite-plus/test"
import { userEvent } from "vite-plus/test/browser"

import {
  getConcludeSetupMockHandler,
  getConcludeSetupMockHandler409,
  getConcludeSetupMockHandler422,
  getConcludeSetupMockHandler429,
  getConcludeSetupMockHandler500,
  getOpenSetupMockHandler,
  getTellSetupStateMockHandler,
} from "../../shared/api/generated/client.msw"
import type { ConcludeSetupRequest } from "../../shared/api/generated/model"
import {
  api,
  noLoginProviderIsOn,
  nobodyIsLoggedIn,
  problem,
  traceId,
} from "../../testing/api"
import { renderApp } from "../../testing/render-app"

type Screen = Awaited<ReturnType<typeof renderApp>>["screen"]

/** The step of the setup code, in each language. */
const codeStep = {
  "en-US": { field: "Setup code", submit: "Continue" },
  "pt-BR": { field: "Código de configuração", submit: "Continuar" },
  es: { field: "Código de configuración", submit: "Continuar" },
}

/** The code the person types, as the command printed it. */
const typedCode = "ABCDE-FGHJK-MNPQR-STVWX"

/**
 * Opens the setup screen of an install not yet set up, in a language the
 * browser asks for, and types a setup code that opens the first setup.
 */
async function openTheFirstSetup(language: keyof typeof codeStep = "en-US") {
  api.use(
    getTellSetupStateMockHandler({
      concluded: false,
      setup_code_waiting: true,
    }),
    // The login screen, where concluding leads, finds nobody logged in.
    nobodyIsLoggedIn(),
    noLoginProviderIsOn()
  )
  api.use(
    getOpenSetupMockHandler({
      opens: "first_setup",
      expires_at: "2026-11-17T16:00:00Z",
    })
  )
  const opened = await renderApp({ at: "/setup", languages: [language] })
  const { screen } = opened
  const texts = codeStep[language]
  await screen.getByRole("textbox", { name: texts.field }).fill(typedCode)
  await screen.getByRole("button", { name: texts.submit }).click()
  return opened
}

/** A field of text, by its label. */
function textField(screen: Screen, label: string) {
  return screen.getByRole("textbox", { name: label, exact: true })
}

/** A field chosen from a list, by its label. */
function listField(screen: Screen, label: string) {
  return screen.getByRole("combobox", { name: label, exact: true })
}

/** Chooses an item of a list, as a person does: types part of it, then picks it. */
async function choose(screen: Screen, label: string, item: string) {
  await listField(screen, label).fill(item.slice(0, 4))
  await screen.getByRole("option", { name: item, exact: true }).click()
}

/** A text found anywhere in what is read, as it was typed. */
function containing(text: string) {
  return new RegExp(text.replaceAll(/[.*+?^${}()|[\]\\]/g, "\\$&"))
}

/** Every field of the consultancy as typed, the optional ones included. */
const consultancy = {
  Name: "Acme Software",
  "Legal name": "Acme Software Ltda.",
  "Tax ID": "12.345.678/0001-90",
  "E-mail (optional)": "contact@acme.example",
  "Telephone (optional)": "+55 11 3333 4444",
  "Website (optional)": "https://acme.example",
  "Line 1 (optional)": "Avenida Paulista, 1000",
  "Line 2 (optional)": "10th floor",
  "City (optional)": "São Paulo",
  "Region (optional)": "SP",
  "Postal code (optional)": "01310-100",
}

/** Every field of the first administrator as typed. */
const administrator = {
  Name: "Ana Souza",
  "E-mail": "Ana@Acme.example",
  "National ID": "123.456.789-09",
  "Telephone (optional)": "+55 11 98888 7777",
}

/** Types the fields given, each by its label. */
async function fillIn(screen: Screen, fields: Record<string, string>) {
  for (const [label, text] of Object.entries(fields)) {
    await textField(screen, label).fill(text)
  }
}

/** Types the whole consultancy, in Brazil, and moves on. */
async function fillTheConsultancy(screen: Screen) {
  await fillIn(screen, consultancy)
  await choose(screen, "Country", "Brazil")
  await screen.getByRole("button", { name: "Continue" }).click()
}

/** Types the whole first administrator, of Brazil, and moves on. */
async function fillTheAdministrator(screen: Screen) {
  await fillIn(screen, administrator)
  await choose(screen, "Country", "Brazil")
  await screen.getByRole("button", { name: "Continue" }).click()
}

/** Opens the first setup in English and types everything up to the review. */
async function reachTheReview() {
  const opened = await openTheFirstSetup()
  await fillTheConsultancy(opened.screen)
  await fillTheAdministrator(opened.screen)
  await expect
    .element(opened.screen.getByRole("heading", { name: "Review" }))
    .toBeVisible()
  return opened
}

describe("the step of the consultancy", () => {
  it("sends nothing and moves on nowhere while a required field is empty", async () => {
    const { screen } = await openTheFirstSetup()
    await expect
      .element(screen.getByRole("heading", { name: "The consultancy" }))
      .toBeVisible()

    await screen.getByRole("button", { name: "Continue" }).click()

    for (const field of ["Name", "Legal name", "Tax ID"]) {
      await expect
        .element(textField(screen, field))
        .toHaveAccessibleDescription(/Fill in this field\./)
    }
    for (const list of ["Country", "Currency"]) {
      await expect
        .element(listField(screen, list))
        .toHaveAccessibleDescription(/Fill in this field\./)
    }
    await expect
      .element(textField(screen, "E-mail (optional)"))
      .not.toHaveAttribute("aria-invalid")
    await expect
      .element(screen.getByRole("heading", { name: "The consultancy" }))
      .toBeVisible()
    await expect.element(textField(screen, "Name")).toHaveFocus()
  })

  it("sends nothing and moves on nowhere while a value is not in the form the contract declares", async () => {
    const { screen } = await openTheFirstSetup()
    await fillIn(screen, {
      ...consultancy,
      "E-mail (optional)": "contact at acme",
      "Telephone (optional)": "11 3333 4444",
      "Website (optional)": "acme.example",
      "Tax ID": "1".repeat(51),
    })
    await choose(screen, "Country", "Brazil")

    await screen.getByRole("button", { name: "Continue" }).click()

    for (const field of [
      "E-mail (optional)",
      "Telephone (optional)",
      "Website (optional)",
    ]) {
      await expect
        .element(textField(screen, field))
        .toHaveAccessibleDescription(/This value is not accepted/)
    }
    await expect
      .element(textField(screen, "Tax ID"))
      .toHaveAccessibleDescription(/Type at most 50 characters\./)
    await expect
      .element(screen.getByRole("heading", { name: "The consultancy" }))
      .toBeVisible()
    await expect.element(textField(screen, "Tax ID")).toHaveFocus()
  })

  it("is filled in and left with the keys alone", async () => {
    const { screen } = await openTheFirstSetup()
    const name = textField(screen, "Name")
    await expect.element(name).toBeVisible()
    await name.click()

    await userEvent.keyboard("Acme Software")
    await userEvent.tab()
    await userEvent.keyboard("Acme Software Ltda.")
    await userEvent.tab()
    await expect.element(listField(screen, "Country")).toHaveFocus()
    await userEvent.keyboard("Brazi")
    await expect
      .element(screen.getByRole("option", { name: "Brazil" }))
      .toBeVisible()
    await userEvent.keyboard("{Enter}")
    await expect.element(listField(screen, "Country")).toHaveValue("Brazil")
    await userEvent.tab()
    await expect.element(textField(screen, "Tax ID")).toHaveFocus()
    await userEvent.keyboard("12.345.678/0001-90{Enter}")

    // The step that opens takes the focus, so it is read, and the next key
    // reaches its first field.
    await expect
      .element(screen.getByRole("heading", { name: "The first administrator" }))
      .toHaveFocus()
    await userEvent.tab()
    await expect.element(textField(screen, "Name")).toHaveFocus()
  })
})

describe("the lists of the step of the consultancy", () => {
  it("name what was chosen again when the language of the screen changes", async () => {
    const { screen } = await openTheFirstSetup()
    await choose(screen, "Country", "Germany")

    await screen.getByRole("combobox", { name: "Language" }).click()
    await screen.getByRole("option", { name: "Português (Brasil)" }).click()

    await expect.element(listField(screen, "País")).toHaveValue("Alemanha")
    await expect.element(listField(screen, "Moeda")).toHaveValue("Euro (EUR)")
  })

  it("names the countries in the language of the screen and filters them by what is typed, whatever the accents", async () => {
    const { screen } = await openTheFirstSetup("pt-BR")
    const country = listField(screen, "País")

    await country.fill("japao")

    await expect
      .element(screen.getByRole("option", { name: "Japão" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("option", { name: "Brasil" }))
      .not.toBeInTheDocument()
    await screen.getByRole("option", { name: "Japão" }).click()
    await expect.element(country).toHaveValue("Japão")
  })

  it("suggests the currency of the country chosen, and again for another country", async () => {
    const { screen } = await openTheFirstSetup()

    await choose(screen, "Country", "Brazil")
    await expect
      .element(listField(screen, "Currency"))
      .toHaveValue("Brazilian Real (BRL)")

    await choose(screen, "Country", "Portugal")
    await expect
      .element(listField(screen, "Currency"))
      .toHaveValue("Euro (EUR)")
  })

  it("keeps a currency chosen by hand when a country is chosen afterwards", async () => {
    const { screen } = await openTheFirstSetup()

    await choose(screen, "Currency", "US Dollar (USD)")
    await choose(screen, "Country", "Brazil")

    await expect
      .element(listField(screen, "Currency"))
      .toHaveValue("US Dollar (USD)")
  })

  it("names the button that shows a list after its field, in the language of the screen", async () => {
    const { screen } = await openTheFirstSetup("pt-BR")

    await screen.getByRole("button", { name: "País: mostrar a lista" }).click()

    await expect
      .element(screen.getByRole("option", { name: "Brasil" }))
      .toBeVisible()
  })

  it("starts the time zone at the one of the browser, by its name in the language of the screen", async () => {
    const { screen } = await openTheFirstSetup("pt-BR")

    await expect
      .element(listField(screen, "Fuso horário"))
      .toHaveValue("America/Sao Paulo (Horário Padrão de Brasília)")
  })
})

describe("the time zone of a browser the API does not know", () => {
  it("starts empty, and the person chooses one", async () => {
    const real = Intl.DateTimeFormat().resolvedOptions()
    const spy = vi
      .spyOn(Intl.DateTimeFormat.prototype, "resolvedOptions")
      .mockReturnValue({ ...real, timeZone: "Etc/Unknown" })
    onTestFinished(() => spy.mockRestore())
    const { screen } = await openTheFirstSetup()

    await screen.getByRole("button", { name: "Continue" }).click()

    await expect.element(listField(screen, "Time zone")).toHaveValue("")
    await expect
      .element(listField(screen, "Time zone"))
      .toHaveAccessibleDescription(/Fill in this field\./)
  })
})

describe("the step of the first administrator", () => {
  it("asks for the administrator once the consultancy is typed, and sends nothing while a required field is empty", async () => {
    const { screen } = await openTheFirstSetup()
    await fillTheConsultancy(screen)
    await expect
      .element(screen.getByRole("heading", { name: "The first administrator" }))
      .toBeVisible()

    await screen.getByRole("button", { name: "Continue" }).click()

    for (const field of ["Name", "E-mail", "National ID"]) {
      await expect
        .element(textField(screen, field))
        .toHaveAccessibleDescription(/Fill in this field\./)
    }
    await expect
      .element(listField(screen, "Country"))
      .toHaveAccessibleDescription(/Fill in this field\./)
    await expect.element(textField(screen, "Name")).toHaveFocus()
  })

  it("keeps what was typed in it when going back to the consultancy and on again", async () => {
    const { screen } = await openTheFirstSetup()
    await fillTheConsultancy(screen)
    await textField(screen, "Name").fill("Ana")

    await screen.getByRole("button", { name: "Back" }).click()
    await expect
      .element(textField(screen, "Name"))
      .toHaveValue(consultancy.Name)
    await screen.getByRole("button", { name: "Continue" }).click()

    await expect.element(textField(screen, "Name")).toHaveValue("Ana")
  })
})

describe("the review", () => {
  it("shows everything typed, the e-mail of the administrator in plain sight", async () => {
    const { screen } = await reachTheReview()

    const ofTheConsultancy = screen.getByRole("region", {
      name: "The consultancy",
    })
    for (const typed of [
      ...Object.values(consultancy),
      "Brazil",
      "Brazilian Real (BRL)",
      "America/Sao Paulo",
    ]) {
      await expect
        .element(ofTheConsultancy)
        .toMatchTextContent(containing(typed))
    }
    const ofTheAdministrator = screen.getByRole("region", {
      name: "The first administrator",
    })
    for (const typed of [...Object.values(administrator), "Brazil"]) {
      await expect
        .element(ofTheAdministrator)
        .toMatchTextContent(containing(typed))
    }
    await expect
      .element(screen.getByText(/The first administrator logs in with/))
      .toMatchTextContent(containing(administrator["E-mail"]))
  })

  it("leads back to each step, with what was typed, and on to the review again", async () => {
    const { screen } = await reachTheReview()

    await screen.getByRole("button", { name: "Change the consultancy" }).click()
    await expect
      .element(textField(screen, "Legal name"))
      .toHaveValue(consultancy["Legal name"])
    await textField(screen, "Name").fill("Acme")
    await screen.getByRole("button", { name: "Continue" }).click()
    await expect
      .element(textField(screen, "National ID"))
      .toHaveValue(administrator["National ID"])
    await screen.getByRole("button", { name: "Continue" }).click()

    await expect
      .element(screen.getByRole("region", { name: "The consultancy" }))
      .toMatchTextContent(/Name\s*Acme\s*Legal name/)
    await screen
      .getByRole("button", { name: "Change the first administrator" })
      .click()
    await expect
      .element(textField(screen, "Name"))
      .toHaveValue(administrator.Name)
  })
})

/** The body the setup typed above concludes with, in a language. */
function concludedSetup(
  language: ConcludeSetupRequest["administrator"]["preferred_language"],
  code = typedCode
): ConcludeSetupRequest {
  return {
    code,
    consultancy: {
      name: consultancy.Name,
      legal_name: consultancy["Legal name"],
      country: "BR",
      tax_id: consultancy["Tax ID"],
      time_zone: "America/Sao_Paulo",
      currency: "BRL",
      email: consultancy["E-mail (optional)"],
      telephone: consultancy["Telephone (optional)"],
      website: consultancy["Website (optional)"],
      address: {
        line1: consultancy["Line 1 (optional)"],
        line2: consultancy["Line 2 (optional)"],
        city: consultancy["City (optional)"],
        region: consultancy["Region (optional)"],
        postal_code: consultancy["Postal code (optional)"],
      },
    },
    administrator: {
      name: administrator.Name,
      email: administrator["E-mail"],
      country: "BR",
      national_id: administrator["National ID"],
      telephone: administrator["Telephone (optional)"],
      preferred_language: language,
    },
  }
}

/** Every body the API received to conclude the setup, in order. */
function concludingRequests() {
  const sent: unknown[] = []
  api.use(
    getConcludeSetupMockHandler(async ({ request }) => {
      sent.push(await request.json())
    })
  )
  return sent
}

/** The API answers the state of the setup with this, from now on. */
function setupStateBecomes(concluded: boolean, setupCodeWaiting: boolean) {
  api.use(
    getTellSetupStateMockHandler({
      concluded,
      setup_code_waiting: setupCodeWaiting,
    })
  )
}

describe("concluding the setup", () => {
  it("sends everything once, and leads to the login screen with the notice that the setup is concluded", async () => {
    const { screen, address } = await reachTheReview()
    const sent = concludingRequests()

    await screen.getByRole("button", { name: "Conclude the setup" }).click()

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The setup is concluded/)
    expect(address()).toBe("/login?notice=setup_concluded")
    expect(sent).toEqual([concludedSetup("en-US")])
  })

  it("makes the language of the screen when concluding the preferred language of the administrator", async () => {
    const { screen } = await reachTheReview()
    const sent = concludingRequests()

    await screen.getByRole("combobox", { name: "Language" }).click()
    await screen.getByRole("option", { name: "Español" }).click()
    await screen
      .getByRole("button", { name: "Concluir la configuración" })
      .click()

    await expect
      .element(screen.getByRole("heading", { name: "Iniciar sesión" }))
      .toBeVisible()
    expect(sent).toEqual([concludedSetup("es")])
  })

  it("shows that it waits for the answer, and cannot be pressed again meanwhile", async () => {
    const { screen } = await reachTheReview()
    const sent: unknown[] = []
    let answer = () => {}
    api.use(
      getConcludeSetupMockHandler(async ({ request }) => {
        sent.push(await request.json())
        await new Promise<void>((resolve) => {
          answer = resolve
        })
      })
    )
    const conclude = screen.getByRole("button", { name: "Conclude the setup" })

    await conclude.click()

    await expect.element(conclude).toBeDisabled()
    await expect.poll(() => sent.length).toBe(1)
    await conclude.click({ force: true })
    answer()
    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(sent).toHaveLength(1)
  })

  it("shows a field the API refused on its field, on the step that holds it", async () => {
    const { screen } = await reachTheReview()
    api.use(
      getConcludeSetupMockHandler422(
        problem(422, "validation_failed", {
          errors: [
            {
              location: "body.administrator.email",
              code: "email_invalid",
              params: {},
              message: "the e-mail is not an address",
            },
          ],
        })
      )
    )

    await screen.getByRole("button", { name: "Conclude the setup" }).click()

    await expect
      .element(screen.getByRole("heading", { name: "The first administrator" }))
      .toBeVisible()
    const email = textField(screen, "E-mail")
    await expect
      .element(email)
      .toHaveAccessibleDescription(/Type an e-mail address/)
    await expect.element(email).toHaveAttribute("aria-invalid", "true")
    await expect.element(email).toHaveFocus()
    await expect.element(email).toHaveValue(administrator["E-mail"])
    await expect
      .element(email)
      .toHaveAccessibleDescription(
        new RegExp(`Identifier to pass on: ?${traceId}`)
      )
  })

  it("leads to the first of the steps that hold a refused field, then to the next", async () => {
    const { screen } = await reachTheReview()
    api.use(
      getConcludeSetupMockHandler422(
        problem(422, "validation_failed", {
          errors: [
            {
              location: "body.administrator.country",
              code: "country_unknown",
              params: {},
              message: "the country is not an ISO 3166-1 alpha-2 code",
            },
            {
              location: "body.consultancy.address.city",
              code: "too_long",
              params: { max_length: 100 },
              message: "the text is longer than 100 characters",
            },
          ],
        })
      )
    )

    await screen.getByRole("button", { name: "Conclude the setup" }).click()

    await expect
      .element(textField(screen, "City (optional)"))
      .toHaveAccessibleDescription(/^Type at most 100 characters\./)
    await screen.getByRole("button", { name: "Continue" }).click()
    await expect
      .element(listField(screen, "Country"))
      .toHaveAccessibleDescription(/Choose a country from the list\./)
  })

  it("shows a failure about no field above the review, with its trace identifier", async () => {
    const { screen } = await reachTheReview()
    api.use(
      getConcludeSetupMockHandler500(problem(500, "internal_server_error"))
    )

    await screen.getByRole("button", { name: "Conclude the setup" }).click()

    const alert = screen.getByRole("alert")
    await expect.element(alert).toMatchTextContent(/The server failed/)
    await expect.element(alert).toMatchTextContent(traceId)
    await expect
      .element(screen.getByRole("heading", { name: "Review" }))
      .toBeVisible()
  })

  it("says to wait when the limit of attempts is reached", async () => {
    const { screen } = await reachTheReview()
    api.use(getConcludeSetupMockHandler429(problem(429, "too_many_requests")))

    await screen.getByRole("button", { name: "Conclude the setup" }).click()

    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/Wait a minute and try again/)
  })
})

describe("a setup code that stops working halfway", () => {
  it("returns to the step of the code with everything kept, and a new code resumes at the review", async () => {
    const { screen } = await reachTheReview()
    api.use(getConcludeSetupMockHandler422(problem(422, "setup_code_invalid")))
    setupStateBecomes(false, false)

    await screen.getByRole("button", { name: "Conclude the setup" }).click()

    const code = textField(screen, "Setup code")
    await expect.element(code).toHaveValue("")
    await expect.element(code).toHaveFocus()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The setup code stopped working/)
    const sent = concludingRequests()
    const newCode = "QRSTV-WXYZ0-12345-6789A"
    await code.fill(newCode)
    await screen.getByRole("button", { name: "Continue" }).click()
    await expect
      .element(screen.getByRole("heading", { name: "Review" }))
      .toBeVisible()
    await screen.getByRole("button", { name: "Conclude the setup" }).click()
    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(sent).toEqual([concludedSetup("en-US", newCode)])
  })
})

describe("a setup concluded by someone else first", () => {
  it("tells the second person, whose setup code the first one ended, and leads to the login screen", async () => {
    const { screen, address } = await reachTheReview()
    api.use(getConcludeSetupMockHandler422(problem(422, "setup_code_invalid")))
    setupStateBecomes(true, false)

    await screen.getByRole("button", { name: "Conclude the setup" }).click()

    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The setup was already concluded/)
    expect(address()).toBe("/login?notice=setup_already_concluded")
  })

  it("does not tell a person that nothing was saved when their own earlier attempt got no answer", async () => {
    const { screen, address } = await reachTheReview()
    // The second attempt finds the setup code ended and the setup concluded,
    // perhaps by the first one, whose answer was lost on the way.
    api.use(getConcludeSetupMockHandler422(problem(422, "setup_code_invalid")))
    setupStateBecomes(true, false)
    api.use(
      http.post("*/api/setup/conclude", () => HttpResponse.error(), {
        once: true,
      })
    )
    const conclude = screen.getByRole("button", { name: "Conclude the setup" })
    await conclude.click()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The server did not answer/)

    await conclude.click()

    const notice = screen.getByRole("alert")
    await expect.element(notice).toMatchTextContent(/The setup is concluded/)
    await expect
      .element(notice)
      .toMatchTextContent(/An earlier attempt to conclude it got no answer/)
    await expect
      .element(notice)
      .not.toMatchTextContent(/nothing of what you typed was saved/)
    expect(address()).toBe("/login?notice=setup_concluded_without_answer")
  })

  it("tells a person whose setup code was issued after the setup was concluded", async () => {
    const { screen, address } = await reachTheReview()
    api.use(
      getConcludeSetupMockHandler409(problem(409, "setup_already_concluded"))
    )
    setupStateBecomes(true, true)

    await screen.getByRole("button", { name: "Conclude the setup" }).click()

    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The setup was already concluded/)
    expect(address()).toBe("/login?notice=setup_already_concluded")
  })
})
