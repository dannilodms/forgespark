import { afterEach, describe, expect, it, vi } from "vite-plus/test"

import {
  getAskForLoginCodeMockHandler,
  getCheckLoginCodeMockHandler,
  getLogOutMockHandler,
  getTellLoggedInUserMockHandler,
  getTellLoggedInUserMockHandler500,
  getTellSetupStateMockHandler,
} from "../../shared/api/generated/client.msw"
import type { LoggedInUser } from "../../shared/api/generated/model"
import {
  api,
  noLoginProviderIsOn,
  nobodyIsLoggedIn,
  problem,
  traceId,
} from "../../testing/api"
import { renderApp } from "../../testing/render-app"

type Screen = Awaited<ReturnType<typeof renderApp>>["screen"]

/** The administrator of the tests, as the API answers who is logged in. */
const ana: LoggedInUser = {
  name: "Ana Example",
  email: "ana@consultancy.example",
  preferred_language: "en-US",
  role_keys: ["administrator"],
}

/** The setup is concluded, and no setup code is waiting. */
function setupConcluded(setupCodeWaiting = false) {
  api.use(
    getTellSetupStateMockHandler({
      concluded: true,
      setup_code_waiting: setupCodeWaiting,
    })
  )
}

describe("the gate, without a session", () => {
  it.for(["/", "/login"])("leads %s to the login screen", async (at) => {
    setupConcluded()
    api.use(nobodyIsLoggedIn(), noLoginProviderIsOn())

    const { screen, address } = await renderApp({ at })

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(address()).toBe("/login")
  })

  it("leads any other address to the login screen, which keeps the address to return to", async () => {
    setupConcluded()
    api.use(nobodyIsLoggedIn(), noLoginProviderIsOn())

    const { screen, address } = await renderApp({ at: "/projects/42?tab=2" })

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(address()).toBe("/login?return_to=%2Fprojects%2F42%3Ftab%3D2")
    await expect.element(screen.getByRole("alert")).not.toBeInTheDocument()
  })

  it("leads the address of the setup to the login screen while no setup code is waiting", async () => {
    setupConcluded()
    api.use(nobodyIsLoggedIn(), noLoginProviderIsOn())

    const { address, screen } = await renderApp({ at: "/setup" })

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(address()).toBe("/login")
  })
})

describe("the gate, with a session", () => {
  it("opens the app at the root", async () => {
    setupConcluded()
    api.use(getTellLoggedInUserMockHandler(ana))

    const { screen, address } = await renderApp({ at: "/" })

    await expect
      .element(screen.getByText("You are logged in as Ana Example."))
      .toBeVisible()
    expect(address()).toBe("/")
  })

  it("takes a logged-in user at the login address to the app", async () => {
    setupConcluded()
    api.use(getTellLoggedInUserMockHandler(ana))

    const { screen, address } = await renderApp({
      at: "/login?notice=setup_concluded",
    })

    await expect
      .element(screen.getByText("You are logged in as Ana Example."))
      .toBeVisible()
    expect(address()).toBe("/")
  })

  it("takes a logged-in user at the login address to the address to return to", async () => {
    setupConcluded()
    api.use(getTellLoggedInUserMockHandler(ana))

    const { screen, address } = await renderApp({
      at: "/login?return_to=%2F%3Fview%3Dboard",
    })

    await expect
      .element(screen.getByText("You are logged in as Ana Example."))
      .toBeVisible()
    expect(address()).toBe("/?view=board")
  })

  it.for(["//elsewhere.example/", "https://elsewhere.example/", "projects"])(
    "opens the root instead of %s, which is no path of the app",
    async (returnTo) => {
      setupConcluded()
      api.use(getTellLoggedInUserMockHandler(ana))

      const { screen, address } = await renderApp({
        at: `/login?return_to=${encodeURIComponent(returnTo)}`,
      })

      await expect
        .element(screen.getByText("You are logged in as Ana Example."))
        .toBeVisible()
      expect(address()).toBe("/")
    }
  )

  it("takes a logged-in user at an address the app does not have to the root", async () => {
    setupConcluded()
    api.use(getTellLoggedInUserMockHandler(ana))

    const { screen, address } = await renderApp({ at: "/projects/42" })

    await expect
      .element(screen.getByText("You are logged in as Ana Example."))
      .toBeVisible()
    expect(address()).toBe("/")
  })

  it("takes a logged-in user at the address of the setup to the app while no setup code is waiting", async () => {
    setupConcluded()
    api.use(getTellLoggedInUserMockHandler(ana))

    const { screen, address } = await renderApp({ at: "/setup" })

    await expect
      .element(screen.getByText("You are logged in as Ana Example."))
      .toBeVisible()
    expect(address()).toBe("/")
  })

  it("leads the address of the setup to the setup while a setup code is waiting", async () => {
    setupConcluded(true)

    const { screen, address } = await renderApp({ at: "/setup" })

    await expect
      .element(screen.getByRole("textbox", { name: "Setup code" }))
      .toBeVisible()
    expect(address()).toBe("/setup")
  })

  it("shows the error screen when who is logged in cannot be read", async () => {
    setupConcluded()
    api.use(
      getTellLoggedInUserMockHandler500(problem(500, "internal_server_error"))
    )

    const { screen } = await renderApp({ at: "/" })

    await expect
      .element(
        screen.getByRole("heading", { name: "This screen could not open" })
      )
      .toBeVisible()
    await expect.element(screen.getByRole("alert")).toMatchTextContent(traceId)
  })
})

/** Asks for a code and types it: the API logs `user` in. */
async function logInWithACode(screen: Screen, user: LoggedInUser) {
  api.use(
    getAskForLoginCodeMockHandler(),
    getCheckLoginCodeMockHandler(() => {
      api.use(getTellLoggedInUserMockHandler(user))
    })
  )
  await screen.getByRole("textbox", { name: /E-mail|Correo/ }).fill(user.email)
  await screen
    .getByRole("button", { name: /Send a code|Enviar un código/ })
    .click()
  await screen
    .getByRole("textbox", { name: /Login code|Código/ })
    .fill("123456")
  await screen
    .getByRole("button", { name: /^(Log in|Iniciar sesión)$/ })
    .click()
}

/**
 * The person leaves the tab and comes back to it, `minutes` later: the app
 * reads again what it read before then.
 */
function comeBackToTheTab(minutes: number) {
  vi.setSystemTime(Date.now() + minutes * 60_000)
  window.dispatchEvent(new Event("visibilitychange"))
}

describe("the session", () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it("speaks the preferred language of the user once logged in, and the language of the selector after the logout", async () => {
    setupConcluded()
    api.use(nobodyIsLoggedIn(), noLoginProviderIsOn(), getLogOutMockHandler())
    const { screen, address } = await renderApp({
      at: "/login",
      languages: ["en-US"],
    })
    await screen.getByRole("combobox", { name: "Language" }).click()
    await screen.getByRole("option", { name: "Español" }).click()
    await expect
      .element(screen.getByRole("heading", { name: "Iniciar sesión" }))
      .toBeVisible()

    await logInWithACode(screen, { ...ana, preferred_language: "pt-BR" })

    await expect
      .element(screen.getByText("Você entrou como Ana Example."))
      .toBeVisible()
    expect(document.documentElement.lang).toBe("pt-BR")
    await expect
      .element(screen.getByRole("combobox", { name: "Idioma" }))
      .not.toBeInTheDocument()

    await screen.getByRole("button", { name: "Sair" }).click()

    await expect
      .element(screen.getByRole("heading", { name: "Iniciar sesión" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent("Cerraste la sesión")
    expect(address()).toBe("/login?notice=logged_out")
    expect(document.documentElement.lang).toBe("es")
  })

  it("opens the login with a notice when the session ends while the person works, and returns them afterwards", async () => {
    vi.useFakeTimers({ toFake: ["Date"] })
    setupConcluded()
    const user: LoggedInUser = { ...ana, preferred_language: "pt-BR" }
    api.use(getTellLoggedInUserMockHandler(user))
    const { screen, address } = await renderApp({
      at: "/?view=board",
      languages: ["en-US"],
    })
    await expect
      .element(screen.getByText("Você entrou como Ana Example."))
      .toBeVisible()

    // The session ends: idle for 7 days, or the user was deactivated.
    api.use(nobodyIsLoggedIn(), noLoginProviderIsOn())
    comeBackToTheTab(2)

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent("Your session ended")
    expect(address()).toBe(
      "/login?notice=session_ended&return_to=%2F%3Fview%3Dboard"
    )

    await logInWithACode(screen, user)

    await expect
      .element(screen.getByText("Você entrou como Ana Example."))
      .toBeVisible()
    expect(address()).toBe("/?view=board")
  })
})
