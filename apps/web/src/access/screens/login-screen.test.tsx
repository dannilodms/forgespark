import { afterEach, describe, expect, it, vi } from "vite-plus/test"
import { userEvent } from "vite-plus/test/browser"
import type { Locator } from "vite-plus/test/browser"

import {
  getAskForLoginCodeMockHandler,
  getAskForLoginCodeMockHandler422,
  getAskForLoginCodeMockHandler429,
  getCheckLoginCodeMockHandler,
  getCheckLoginCodeMockHandler422,
  getCheckLoginCodeMockHandler429,
  getTellLoggedInUserMockHandler,
  getTellLoginStateMockHandler,
  getTellLoginStateMockHandler500,
  getTellSetupStateMockHandler,
} from "../../shared/api/generated/client.msw"
import type {
  LoggedInUser,
  LoginState,
  TurnedOnLoginProvider,
} from "../../shared/api/generated/model"
import { api, nobodyIsLoggedIn, problem, traceId } from "../../testing/api"
import { renderApp } from "../../testing/render-app"

/** The administrator of the tests, as the API answers who is logged in. */
const ana: LoggedInUser = {
  name: "Ana Example",
  email: "ana@consultancy.example",
  preferred_language: "en-US",
  role_keys: ["administrator"],
}

/** The local login provider, as `.env.example` turns it on. */
const local: TurnedOnLoginProvider = {
  key: "local",
  display_name: "Local login provider",
}

type Screen = Awaited<ReturnType<typeof renderApp>>["screen"]

/**
 * The login screen, as a browser with no session opens it. The API tells
 * the login state it is given: by default, no login provider is on and no
 * login through one was refused.
 */
async function openTheLogin(
  at = "/login",
  loginState: LoginState = { login_providers: [] }
) {
  api.use(
    getTellSetupStateMockHandler({
      concluded: true,
      setup_code_waiting: false,
    }),
    nobodyIsLoggedIn(),
    getTellLoginStateMockHandler(loginState)
  )
  const app = await renderApp({ at })
  await expect
    .element(app.screen.getByRole("heading", { name: "Log in" }))
    .toBeVisible()
  return app
}

function emailField(screen: Screen) {
  return screen.getByRole("textbox", { name: "E-mail" })
}

function codeField(screen: Screen) {
  return screen.getByRole("textbox", { name: "Login code" })
}

/** Types an e-mail and asks for a code: the API takes it. */
async function askForACode(screen: Screen, email = "Ana@Consultancy.example") {
  await emailField(screen).fill(email)
  await screen.getByRole("button", { name: "Send a code" }).click()
  await expect.element(codeField(screen)).toBeVisible()
}

/**
 * The API takes the code it is sent, and from then on answers that the
 * user is logged in.
 */
function theCodeLogsIn(user = ana) {
  const sent: unknown[] = []
  api.use(
    getCheckLoginCodeMockHandler(async ({ request }) => {
      sent.push(await request.json())
      api.use(getTellLoggedInUserMockHandler(user))
    })
  )
  return sent
}

/** The API answers a code that does not work, with the attempts left. */
function theCodeIsInvalid(attemptsLeft: number) {
  api.use(
    getCheckLoginCodeMockHandler422(
      problem(422, "login_code_invalid", {
        params: { attempts_left: attemptsLeft },
      }),
      { once: true }
    )
  )
}

/** The text of the alert that holds `text`: several may be on the screen. */
function alertWith(screen: Screen, text: string | RegExp): Locator {
  return screen.getByRole("alert").filter({ hasText: text })
}

afterEach(() => {
  vi.useRealTimers()
})

describe("the step of the e-mail", () => {
  it("sends the e-mail as typed, in the body, and says a code was sent if it belongs to a user", async () => {
    const sent: unknown[] = []
    api.use(
      getAskForLoginCodeMockHandler(async ({ request }) => {
        sent.push(await request.json())
      })
    )
    const { screen, address } = await openTheLogin()

    await askForACode(screen, "Ana@Consultancy.example")

    expect(sent).toEqual([{ email: "Ana@Consultancy.example" }])
    await expect
      .element(
        screen.getByText(
          "If Ana@Consultancy.example belongs to a user, a login code was sent to it. Type its six digits."
        )
      )
      .toBeVisible()
    await expect.element(codeField(screen)).toHaveFocus()
    expect(address()).toBe("/login")
  })

  it("sends nothing while the e-mail is empty or is not an e-mail, and says why on the field", async () => {
    const { screen } = await openTheLogin()
    const field = emailField(screen)

    await screen.getByRole("button", { name: "Send a code" }).click()
    await expect
      .element(field)
      .toHaveAccessibleDescription("Fill in this field.")
    await expect.element(field).toHaveFocus()

    await field.fill("ana")
    await screen.getByRole("button", { name: "Send a code" }).click()
    await expect
      .element(field)
      .toHaveAccessibleDescription(/This value is not accepted/)
  })

  it("says to wait past the limit of codes for the address", async () => {
    api.use(
      getAskForLoginCodeMockHandler422(problem(422, "login_code_limit_reached"))
    )
    const { screen } = await openTheLogin()

    await emailField(screen).fill("ana@consultancy.example")
    await screen.getByRole("button", { name: "Send a code" }).click()

    const alert = screen.getByRole("alert")
    await expect
      .element(alert)
      .toMatchTextContent(/Too many codes asked for this e-mail/)
    await expect
      .element(alert)
      .toMatchTextContent(/Wait a few minutes before asking for another one/)
    await expect.element(alert).toMatchTextContent(traceId)
    await expect
      .element(emailField(screen))
      .toHaveValue("ana@consultancy.example")
  })

  it("shows the limit of Caddy as the setup does", async () => {
    api.use(getAskForLoginCodeMockHandler429(problem(429, "too_many_requests")))
    const { screen } = await openTheLogin()

    await emailField(screen).fill("ana@consultancy.example")
    await screen.getByRole("button", { name: "Send a code" }).click()

    const alert = screen.getByRole("alert")
    await expect
      .element(alert)
      .toMatchTextContent(/Too many attempts from this address/)
    await expect.element(alert).toMatchTextContent(traceId)
  })

  it("starts again at the e-mail on a page loaded again, with nothing typed", async () => {
    api.use(getAskForLoginCodeMockHandler())
    const first = await openTheLogin()
    await askForACode(first.screen)

    expect(localStorage.length).toBe(0)
    expect(sessionStorage.length).toBe(0)
    await first.screen.unmount()

    const again = await renderApp({ at: "/login" })
    await expect.element(emailField(again.screen)).toHaveValue("")
    await expect.element(codeField(again.screen)).not.toBeInTheDocument()
  })
})

describe("the step of the code", () => {
  it("sends the six digits, in the body, and opens the app", async () => {
    api.use(getAskForLoginCodeMockHandler())
    const sent = theCodeLogsIn()
    const { screen, address } = await openTheLogin()
    await askForACode(screen)

    await codeField(screen).fill("123456")
    await screen.getByRole("button", { name: "Log in" }).click()

    await expect
      .element(screen.getByText("You are logged in as Ana Example."))
      .toBeVisible()
    expect(sent).toEqual([{ code: "123456" }])
    expect(address()).toBe("/")
  })

  it("opens the address to return to after the login", async () => {
    api.use(getAskForLoginCodeMockHandler())
    theCodeLogsIn()
    const { screen, address } = await openTheLogin(
      "/login?notice=session_ended&return_to=%2F%3Fview%3Dboard"
    )
    await askForACode(screen)

    await codeField(screen).fill("123456")
    await screen.getByRole("button", { name: "Log in" }).click()

    await expect
      .element(screen.getByText("You are logged in as Ana Example."))
      .toBeVisible()
    expect(address()).toBe("/?view=board")
  })

  it("sends nothing while the code has fewer than six digits", async () => {
    api.use(getAskForLoginCodeMockHandler())
    const { screen } = await openTheLogin()
    await askForACode(screen)
    const field = codeField(screen)

    await screen.getByRole("button", { name: "Log in" }).click()
    await expect
      .element(field)
      .toHaveAccessibleDescription("Fill in this field.")

    await field.fill("123")
    await screen.getByRole("button", { name: "Log in" }).click()
    await expect
      .element(field)
      .toHaveAccessibleDescription(/This value is not accepted/)
    await expect.element(field).toHaveFocus()
  })

  it("shows a wrong code under the field with the attempts left, and keeps the step", async () => {
    api.use(getAskForLoginCodeMockHandler())
    theCodeIsInvalid(2)
    const { screen } = await openTheLogin()
    await askForACode(screen)
    const field = codeField(screen)

    await field.fill("000000")
    await screen.getByRole("button", { name: "Log in" }).click()

    await expect
      .element(field)
      .toHaveAccessibleDescription(/This code does not work.*Attempts left: 2/)
    await expect
      .element(field)
      .toHaveAccessibleDescription(
        new RegExp(`Identifier to pass on: ?${traceId}`)
      )
    await expect.element(field).toHaveAttribute("aria-invalid", "true")
    await expect.element(field).toHaveValue("000000")
    await expect.element(field).toHaveFocus()

    theCodeIsInvalid(1)
    await field.fill("000001")
    await screen.getByRole("button", { name: "Log in" }).click()
    await expect.element(field).toHaveAccessibleDescription(/Attempts left: 1/)
  })

  it("returns to the e-mail step with a notice at the third wrong attempt, which ends the code", async () => {
    api.use(
      getAskForLoginCodeMockHandler(),
      getCheckLoginCodeMockHandler422(problem(422, "login_code_ended"))
    )
    theCodeIsInvalid(1)
    theCodeIsInvalid(2)
    const { screen } = await openTheLogin()
    await askForACode(screen, "ana@consultancy.example")

    for (const code of ["000000", "000001", "000002"]) {
      await codeField(screen).fill(code)
      await screen.getByRole("button", { name: "Log in" }).click()
    }

    await expect
      .element(alertWith(screen, "This code no longer works"))
      .toMatchTextContent(/Ask for a new code/)
    await expect.element(codeField(screen)).not.toBeInTheDocument()
    await expect
      .element(emailField(screen))
      .toHaveValue("ana@consultancy.example")
    await expect.element(emailField(screen)).toHaveFocus()
  })

  it("offers a new code a minute after the last one, counting the seconds", async () => {
    vi.useFakeTimers({ toFake: ["Date", "setInterval", "clearInterval"] })
    const sent: unknown[] = []
    api.use(
      getAskForLoginCodeMockHandler(async ({ request }) => {
        sent.push(await request.json())
      })
    )
    theCodeIsInvalid(2)
    const { screen } = await openTheLogin()
    await askForACode(screen, "ana@consultancy.example")
    const newCode = screen.getByRole("button", { name: "Send a new code" })

    await expect.element(newCode).toBeDisabled()
    await expect
      .element(newCode)
      .toHaveAccessibleDescription("You can ask for a new code in 60 seconds.")
    vi.advanceTimersByTime(59_000)
    await expect
      .element(newCode)
      .toHaveAccessibleDescription("You can ask for a new code in 1 second.")
    await expect.element(newCode).toBeDisabled()

    // A wrong code meanwhile, which the new code leaves behind.
    await codeField(screen).fill("000000")
    await screen.getByRole("button", { name: "Log in" }).click()
    await expect
      .element(codeField(screen))
      .toHaveAccessibleDescription(/Attempts left: 2/)

    vi.advanceTimersByTime(1_000)
    await expect.element(newCode).toBeEnabled()
    await newCode.click()

    await expect.element(newCode).toBeDisabled()
    await expect
      .element(newCode)
      .toHaveAccessibleDescription("You can ask for a new code in 60 seconds.")
    expect(sent).toEqual([
      { email: "ana@consultancy.example" },
      { email: "ana@consultancy.example" },
    ])
    await expect.element(codeField(screen)).toHaveValue("")
    await expect.element(codeField(screen)).not.toHaveAttribute("aria-invalid")
    await expect.element(codeField(screen)).toHaveFocus()
  })

  it("says to wait when a new code meets the limit of codes for the address", async () => {
    vi.useFakeTimers({ toFake: ["Date", "setInterval", "clearInterval"] })
    api.use(getAskForLoginCodeMockHandler(undefined, { once: true }))
    const { screen } = await openTheLogin()
    await askForACode(screen)
    api.use(
      getAskForLoginCodeMockHandler422(problem(422, "login_code_limit_reached"))
    )

    vi.advanceTimersByTime(60_000)
    await screen.getByRole("button", { name: "Send a new code" }).click()

    await expect
      .element(alertWith(screen, "Too many codes asked for this e-mail"))
      .toMatchTextContent(traceId)
    await expect.element(codeField(screen)).toBeVisible()
  })

  it("shows the limit of Caddy on a check as the setup does, and keeps the code", async () => {
    api.use(
      getAskForLoginCodeMockHandler(),
      getCheckLoginCodeMockHandler429(problem(429, "too_many_requests"))
    )
    const { screen } = await openTheLogin()
    await askForACode(screen)

    await codeField(screen).fill("123456")
    await screen.getByRole("button", { name: "Log in" }).click()

    await expect
      .element(alertWith(screen, "Too many attempts from this address"))
      .toMatchTextContent(traceId)
    await expect.element(codeField(screen)).toHaveValue("123456")
  })

  it("goes back to the e-mail step, which holds the e-mail typed", async () => {
    api.use(getAskForLoginCodeMockHandler())
    const { screen } = await openTheLogin()
    await askForACode(screen, "ana@consultancy.example")

    await screen.getByRole("button", { name: "Use another e-mail" }).click()

    await expect
      .element(emailField(screen))
      .toHaveValue("ana@consultancy.example")
    await expect.element(codeField(screen)).not.toBeInTheDocument()
  })
})

/** The button of a login provider, which leads the browser to its login. */
function providerButton(screen: Screen, name: string) {
  return screen.getByRole("link", { name: `Log in with ${name}` })
}

describe("the login providers", () => {
  it("draws the button of the one provider turned on, above the divider, leading to its login", async () => {
    const { screen } = await openTheLogin("/login", {
      login_providers: [local],
    })

    const button = providerButton(screen, "Local login provider")
    await expect.element(button).toBeVisible()
    await expect
      .element(button)
      .toHaveAttribute("href", "/api/session/login_providers/local/start")
    await expect.element(screen.getByText("or", { exact: true })).toBeVisible()
    await expect.element(emailField(screen)).toBeVisible()
  })

  it("draws one button for each provider turned on, in the order of the login state", async () => {
    const { screen } = await openTheLogin("/login", {
      login_providers: [
        { key: "google", display_name: "Google" },
        { key: "microsoft", display_name: "Microsoft" },
        local,
      ],
    })

    const buttons = screen.getByRole("link", { name: /^Log in with / })
    await expect.element(buttons.first()).toBeVisible()
    expect(buttons.elements().map((button) => button.textContent)).toEqual([
      "Log in with Google",
      "Log in with Microsoft",
      "Log in with Local login provider",
    ])
    expect(
      buttons.elements().map((button) => button.getAttribute("href"))
    ).toEqual([
      "/api/session/login_providers/google/start",
      "/api/session/login_providers/microsoft/start",
      "/api/session/login_providers/local/start",
    ])
  })

  it("draws neither a button nor the divider while no provider is on", async () => {
    const { screen } = await openTheLogin()

    await expect.element(emailField(screen)).toBeVisible()
    await expect
      .element(screen.getByRole("link", { name: /^Log in with / }))
      .not.toBeInTheDocument()
    await expect
      .element(screen.getByText("or", { exact: true }))
      .not.toBeInTheDocument()
  })

  it("passes the address to return to on to the login through the provider", async () => {
    const { screen } = await openTheLogin(
      "/login?notice=session_ended&return_to=%2Fprojects%2F42%3Ftab%3D2",
      { login_providers: [local] }
    )

    await expect
      .element(providerButton(screen, "Local login provider"))
      .toHaveAttribute(
        "href",
        "/api/session/login_providers/local/start?return_to=%2Fprojects%2F42%3Ftab%3D2"
      )
  })

  it.for(["//elsewhere.example/", "https://elsewhere.example/", "/"])(
    "passes no address to return to for %s, which leads to the root",
    async (returnTo) => {
      const { screen } = await openTheLogin(
        `/login?return_to=${encodeURIComponent(returnTo)}`,
        { login_providers: [local] }
      )

      await expect
        .element(providerButton(screen, "Local login provider"))
        .toHaveAttribute("href", "/api/session/login_providers/local/start")
    }
  )

  it("passes no address to return to longer than the contract takes", async () => {
    const { screen } = await openTheLogin(
      `/login?return_to=${encodeURIComponent(`/projects?q=${"a".repeat(2048)}`)}`,
      { login_providers: [local] }
    )

    await expect
      .element(providerButton(screen, "Local login provider"))
      .toHaveAttribute("href", "/api/session/login_providers/local/start")
  })

  it("shows the error screen when the login state cannot be read", async () => {
    api.use(
      getTellLoginStateMockHandler500(problem(500, "internal_server_error"))
    )
    api.use(
      getTellSetupStateMockHandler({
        concluded: true,
        setup_code_waiting: false,
      }),
      nobodyIsLoggedIn()
    )

    const { screen } = await renderApp({ at: "/login" })

    await expect
      .element(
        screen.getByRole("heading", { name: "This screen could not open" })
      )
      .toBeVisible()
    await expect.element(screen.getByRole("alert")).toMatchTextContent(traceId)
  })
})

describe("the notices of the login screen", () => {
  it.for([
    ["setup_concluded", "The setup is concluded", /logs in below/],
    [
      "setup_already_concluded",
      "The setup was already concluded",
      /logs in below/,
    ],
    [
      "setup_concluded_without_answer",
      "The setup is concluded",
      /logs in below/,
    ],
    [
      "administrator_registered",
      "The administrator is registered",
      /logs in below/,
    ],
    [
      "administrator_registered_without_answer",
      "The administrator may be registered",
      /Try logging in below with that e-mail/,
    ],
    ["session_ended", "Your session ended", /Log in again/],
    ["logged_out", "You logged out", /log in below/],
  ] as const)("shows %s", async ([notice, title, todo]) => {
    const { screen } = await openTheLogin(`/login?notice=${notice}`)

    const alert = screen.getByRole("alert")
    await expect.element(alert).toMatchTextContent(title)
    await expect.element(alert).toMatchTextContent(todo)
  })

  it("shows login_not_concluded, and that the person can try again, with the button of the provider", async () => {
    const { screen } = await openTheLogin(
      "/login?notice=login_not_concluded&return_to=%2Fprojects",
      { login_providers: [local] }
    )

    const alert = screen.getByRole("alert")
    await expect.element(alert).toMatchTextContent("The login did not conclude")
    await expect.element(alert).toMatchTextContent(/You can try again below/)
    await expect
      .element(providerButton(screen, "Local login provider"))
      .toHaveAttribute(
        "href",
        "/api/session/login_providers/local/start?return_to=%2Fprojects"
      )
  })

  it("shows no_active_user with the e-mail of the login state, and to ask an administrator, the same for an unknown and a deactivated user", async () => {
    const { screen } = await openTheLogin(
      "/login?notice=no_active_user&email=someone%40else.example",
      {
        login_providers: [local],
        refused_email: "bruno@consultancy.example",
      }
    )

    const alert = screen.getByRole("alert")
    await expect
      .element(alert)
      .toMatchTextContent("No active user has this e-mail")
    await expect
      .element(alert)
      .toMatchTextContent(
        "The login provider gave the e-mail bruno@consultancy.example, and no active user of the system has it. Ask an administrator for access."
      )
    await expect
      .element(screen.getByText("someone@else.example"))
      .not.toBeInTheDocument()
  })

  it("shows no_active_user in the language of the screen", async () => {
    api.use(
      getTellSetupStateMockHandler({
        concluded: true,
        setup_code_waiting: false,
      }),
      nobodyIsLoggedIn(),
      getTellLoginStateMockHandler({
        login_providers: [local],
        refused_email: "bruno@consultancy.example",
      })
    )

    const { screen } = await renderApp({
      at: "/login?notice=no_active_user",
      languages: ["pt-BR"],
    })

    const alert = screen.getByRole("alert")
    await expect
      .element(alert)
      .toMatchTextContent("Nenhum usuário ativo tem este e-mail")
    await expect
      .element(alert)
      .toMatchTextContent(
        /bruno@consultancy\.example.*Peça acesso a um administrador/
      )
    await expect
      .element(providerButton(screen, "Local login provider"))
      .not.toBeInTheDocument()
    await expect
      .element(
        screen.getByRole("link", { name: "Entrar com Local login provider" })
      )
      .toBeVisible()
  })

  it("shows no notice for no_active_user when no login was refused in this browser", async () => {
    const { screen } = await openTheLogin("/login?notice=no_active_user", {
      login_providers: [local],
    })

    await expect.element(emailField(screen)).toBeVisible()
    await expect.element(screen.getByRole("alert")).not.toBeInTheDocument()
  })

  it("shows no notice for a value it does not know", async () => {
    const { screen } = await openTheLogin("/login?notice=nonsense")

    await expect.element(emailField(screen)).toBeVisible()
    await expect.element(screen.getByRole("alert")).not.toBeInTheDocument()
  })
})

describe("the login screen by keyboard", () => {
  it("asks for a code and logs in with the keys alone", async () => {
    api.use(getAskForLoginCodeMockHandler())
    theCodeLogsIn()
    const { screen } = await openTheLogin()

    await userEvent.tab()
    await expect.element(emailField(screen)).toHaveFocus()
    await userEvent.keyboard("ana@consultancy.example{Enter}")
    await expect.element(codeField(screen)).toHaveFocus()
    await userEvent.keyboard("123456{Enter}")

    await expect
      .element(screen.getByText("You are logged in as Ana Example."))
      .toBeVisible()
  })
})
