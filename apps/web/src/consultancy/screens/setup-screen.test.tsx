import { describe, expect, it } from "vite-plus/test"
import { userEvent } from "vite-plus/test/browser"

import {
  getOpenSetupMockHandler,
  getOpenSetupMockHandler422,
  getOpenSetupMockHandler429,
  getOpenSetupMockHandler500,
  getTellSetupStateMockHandler,
  getTellSetupStateMockHandler500,
} from "../../shared/api/generated/client.msw"
import { api, problem, traceId } from "../../testing/api"
import { renderApp } from "../../testing/render-app"

/** The API answers the state of the setup with this. */
function setupState(concluded: boolean, setupCodeWaiting: boolean) {
  api.use(
    getTellSetupStateMockHandler({
      concluded,
      setup_code_waiting: setupCodeWaiting,
    })
  )
}

describe("the gate of the setup", () => {
  describe("before the setup is concluded", () => {
    it.for(["/", "/setup", "/login", "/projects/42"])(
      "leads %s to the step of the setup code",
      async (at) => {
        setupState(false, true)

        const { screen, address } = await renderApp({ at })

        await expect
          .element(screen.getByRole("textbox", { name: "Setup code" }))
          .toBeVisible()
        expect(address()).toBe("/setup")
      }
    )
  })

  describe("after the setup is concluded", () => {
    it.for(["/", "/setup", "/login", "/projects/42"])(
      "leads %s to the login screen while no setup code is waiting",
      async (at) => {
        setupState(true, false)

        const { screen, address } = await renderApp({ at })

        await expect
          .element(screen.getByRole("heading", { name: "Log in" }))
          .toBeVisible()
        expect(address()).toBe("/login")
      }
    )

    it("leads the address of the setup to the step of the setup code while a setup code is waiting", async () => {
      setupState(true, true)

      const { screen, address } = await renderApp({ at: "/setup" })

      await expect
        .element(screen.getByRole("textbox", { name: "Setup code" }))
        .toBeVisible()
      expect(address()).toBe("/setup")
    })

    it.for(["/", "/login", "/projects/42"])(
      "leads %s to the login screen even while a setup code is waiting",
      async (at) => {
        setupState(true, true)

        const { screen, address } = await renderApp({ at })

        await expect
          .element(screen.getByRole("heading", { name: "Log in" }))
          .toBeVisible()
        expect(address()).toBe("/login")
      }
    )
  })
})

describe("the step of the setup code", () => {
  it("sends the code as typed, in the body, and moves on to the step of the consultancy", async () => {
    setupState(false, true)
    let sent: unknown
    api.use(
      getOpenSetupMockHandler(async ({ request }) => {
        sent = await request.json()
        return { opens: "first_setup", expires_at: "2026-11-17T16:00:00Z" }
      })
    )
    const { screen } = await renderApp({ at: "/setup" })

    await screen
      .getByRole("textbox", { name: "Setup code" })
      .fill("abcde fghjk-mnpqr-stvwx")
    await screen.getByRole("button", { name: "Continue" }).click()

    await expect
      .element(screen.getByRole("heading", { name: "The consultancy" }))
      .toBeVisible()
    expect(sent).toEqual({ code: "abcde fghjk-mnpqr-stvwx" })
  })

  it("moves on to the registration of a new administrator when the code opens it", async () => {
    setupState(true, true)
    api.use(
      getOpenSetupMockHandler({
        opens: "new_administrator",
        expires_at: "2026-11-17T16:00:00Z",
      })
    )
    const { screen } = await renderApp({ at: "/setup" })

    await screen
      .getByRole("textbox", { name: "Setup code" })
      .fill("ABCDE-FGHJK-MNPQR-STVWX")
    await screen.getByRole("button", { name: "Continue" }).click()

    await expect
      .element(
        screen.getByText(
          /The setup code opened the registration of a new administrator/
        )
      )
      .toBeVisible()
  })

  it("shows a code that does not work on its field, as typed, and asks for a new one", async () => {
    setupState(false, true)
    api.use(
      getOpenSetupMockHandler422(
        problem(422, "setup_code_invalid", {
          detail: "the setup code is wrong, expired or was ended",
        })
      )
    )
    const { screen } = await renderApp({ at: "/setup" })
    const field = screen.getByRole("textbox", { name: "Setup code" })

    await field.fill("ABCDE-FGHJK-MNPQR-STVWZ")
    await screen.getByRole("button", { name: "Continue" }).click()

    await expect
      .element(field)
      .toHaveAccessibleDescription(
        /This setup code does not work.*Issue a new one/
      )
    await expect.element(field).toHaveAttribute("aria-invalid", "true")
    await expect.element(field).toHaveValue("ABCDE-FGHJK-MNPQR-STVWZ")
    await expect.element(field).toHaveFocus()
  })

  it("says to wait when the limit of attempts is reached", async () => {
    setupState(false, true)
    api.use(getOpenSetupMockHandler429(problem(429, "too_many_requests")))
    const { screen } = await renderApp({ at: "/setup" })

    await screen
      .getByRole("textbox", { name: "Setup code" })
      .fill("ABCDE-FGHJK-MNPQR-STVWX")
    await screen.getByRole("button", { name: "Continue" }).click()

    const alert = screen.getByRole("alert")
    await expect
      .element(alert)
      .toMatchTextContent(/Too many attempts from this address/)
    await expect
      .element(alert)
      .toMatchTextContent(/Wait a minute and try again/)
  })

  it("shows a failure it does not expect with its trace identifier", async () => {
    setupState(false, true)
    api.use(getOpenSetupMockHandler500(problem(500, "internal_server_error")))
    const { screen } = await renderApp({ at: "/setup" })

    await screen
      .getByRole("textbox", { name: "Setup code" })
      .fill("ABCDE-FGHJK-MNPQR-STVWX")
    await screen.getByRole("button", { name: "Continue" }).click()

    const alert = screen.getByRole("alert")
    await expect.element(alert).toMatchTextContent(/The server failed/)
    await expect.element(alert).toMatchTextContent(traceId)
  })

  it("sends nothing while the field is empty, and says it is required", async () => {
    setupState(false, true)
    const { screen } = await renderApp({ at: "/setup" })
    const field = screen.getByRole("textbox", { name: "Setup code" })

    await screen.getByRole("button", { name: "Continue" }).click()

    await expect
      .element(field)
      .toHaveAccessibleDescription("Fill in this field.")
  })

  it("keeps the code in memory only, and asks for it again on a page loaded again", async () => {
    setupState(false, true)
    api.use(
      getOpenSetupMockHandler({
        opens: "first_setup",
        expires_at: "2026-11-17T16:00:00Z",
      })
    )
    const code = "ABCDE-FGHJK-MNPQR-STVWX"
    const first = await renderApp({ at: "/setup" })
    await first.screen.getByRole("textbox", { name: "Setup code" }).fill(code)
    await first.screen.getByRole("button", { name: "Continue" }).click()
    await expect
      .element(first.screen.getByRole("heading", { name: "The consultancy" }))
      .toBeVisible()

    expect(first.address()).toBe("/setup")
    expect(localStorage.length).toBe(0)
    expect(sessionStorage.length).toBe(0)
    expect(document.cookie).toBe("")

    await first.screen.unmount()
    const again = await renderApp({ at: "/setup" })
    await expect
      .element(again.screen.getByRole("textbox", { name: "Setup code" }))
      .toHaveValue("")
  })
})

describe("a failure while the screen opens", () => {
  it("shows the error screen, with the trace identifier and a reload", async () => {
    api.use(
      getTellSetupStateMockHandler500(problem(500, "internal_server_error"))
    )

    const { screen } = await renderApp({ at: "/setup" })

    await expect
      .element(
        screen.getByRole("heading", { name: "This screen could not open" })
      )
      .toBeVisible()
    await expect.element(screen.getByRole("alert")).toMatchTextContent(traceId)
    await expect
      .element(screen.getByRole("button", { name: "Reload the page" }))
      .toBeVisible()
  })
})

describe("the language of the setup screen", () => {
  it("is the language of the browser when it is supported", async () => {
    setupState(false, true)

    const { screen } = await renderApp({
      at: "/setup",
      languages: ["pt-BR", "en-US"],
    })

    await expect
      .element(screen.getByRole("textbox", { name: "Código de configuração" }))
      .toBeVisible()
    expect(document.documentElement.lang).toBe("pt-BR")
  })

  it("is the closest supported one for a regional variant", async () => {
    setupState(false, true)

    const { screen } = await renderApp({ at: "/setup", languages: ["es-AR"] })

    await expect
      .element(screen.getByRole("textbox", { name: "Código de configuración" }))
      .toBeVisible()
  })

  it("is English when the browser asks for no supported language", async () => {
    setupState(false, true)

    const { screen } = await renderApp({
      at: "/setup",
      languages: ["de-DE", "fr"],
    })

    await expect
      .element(screen.getByRole("textbox", { name: "Setup code" }))
      .toBeVisible()
  })

  it("changes at once when another language is chosen, keeping what was typed", async () => {
    setupState(false, true)
    const { screen } = await renderApp({ at: "/setup" })
    await screen.getByRole("textbox", { name: "Setup code" }).fill("ABCDE")

    await screen.getByRole("combobox", { name: "Language" }).click()
    await screen.getByRole("option", { name: "Português (Brasil)" }).click()

    await expect
      .element(screen.getByRole("heading", { name: "Configuração inicial" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("textbox", { name: "Código de configuração" }))
      .toHaveValue("ABCDE")
    await expect
      .element(screen.getByRole("combobox", { name: "Idioma" }))
      .toBeVisible()
    expect(document.documentElement.lang).toBe("pt-BR")
  })
})

describe("the setup screen by keyboard", () => {
  it("takes the code and moves on with the keys alone", async () => {
    setupState(false, true)
    api.use(
      getOpenSetupMockHandler({
        opens: "first_setup",
        expires_at: "2026-11-17T16:00:00Z",
      })
    )
    const { screen } = await renderApp({ at: "/setup" })
    const field = screen.getByRole("textbox", { name: "Setup code" })
    await expect.element(field).toBeVisible()

    await userEvent.tab()
    await expect.element(field).toHaveFocus()
    await userEvent.keyboard("ABCDE-FGHJK-MNPQR-STVWX{Enter}")

    await expect
      .element(screen.getByRole("heading", { name: "The consultancy" }))
      .toBeVisible()
  })

  it("chooses another language with the keys alone", async () => {
    setupState(false, true)
    const { screen } = await renderApp({ at: "/setup" })
    const language = screen.getByRole("combobox", { name: "Language" })
    await expect.element(language).toBeVisible()

    await userEvent.tab()
    await userEvent.tab()
    await userEvent.tab()
    await expect.element(language).toHaveFocus()
    await userEvent.keyboard("{Enter}")
    await expect
      .element(screen.getByRole("option", { name: "Español" }))
      .toBeVisible()
    await userEvent.keyboard("{ArrowDown}{Enter}")

    await expect
      .element(screen.getByRole("heading", { name: "Configuración inicial" }))
      .toBeVisible()
  })
})
