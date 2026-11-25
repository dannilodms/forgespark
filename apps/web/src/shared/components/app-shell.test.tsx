import { HttpResponse, http } from "msw"
import { afterEach, beforeEach, describe, expect, it, vi } from "vite-plus/test"
import { page, userEvent } from "vite-plus/test/browser"

import {
  getLogOutMockHandler,
  getLogOutMockHandler500,
  getTellConsultancyMockHandler,
  getTellLoggedInUserMockHandler,
  getTellSetupStateMockHandler,
} from "../api/generated/client.msw"
import type { Consultancy, LoggedInUser } from "../api/generated/model"
import {
  api,
  loggedIn,
  noLoginProviderIsOn,
  nobodyIsLoggedIn,
  problem,
  theConsultancy,
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

/** Where the storage is reached, as the API signs a URL for it. */
const storageAddress = "http://files.localhost/forgespark-files"

/** A URL the API signed to read an object, for fifteen minutes. */
function signedURL(key: string, signature = "first") {
  return `${storageAddress}/${key}?X-Amz-Expires=900&X-Amz-Signature=${signature}`
}

/** A PNG drawn in the browser, as the storage answers it. */
async function aPNG() {
  const canvas = new OffscreenCanvas(48, 48)
  const context = canvas.getContext("2d")
  if (context) {
    context.fillStyle = "#0073ea"
    context.fillRect(0, 0, 48, 48)
  }
  return (await canvas.convertToBlob({ type: "image/png" })).arrayBuffer()
}

/**
 * The file storage, which answers every signed URL with an image, as the
 * browser reads it. Returns each URL it was asked for.
 */
async function theStorage() {
  const image = await aPNG()
  const read: string[] = []
  api.use(
    http.get(`${storageAddress}/*`, ({ request }) => {
      read.push(request.url)
      return HttpResponse.arrayBuffer(image, {
        headers: { "Content-Type": "image/png" },
      })
    })
  )
  return read
}

/** The setup is concluded, and the API answers as `user` is logged in. */
async function openTheShell(
  user: LoggedInUser = ana,
  consultancy: Consultancy = theConsultancy,
  {
    at = "/",
    languages = ["en-US"],
  }: { at?: string; languages?: string[] } = {}
) {
  api.use(
    getTellSetupStateMockHandler({
      concluded: true,
      setup_code_waiting: false,
    }),
    ...loggedIn(user, consultancy)
  )
  const app = await renderApp({ at, languages })
  await expect.element(app.screen.getByRole("main")).toBeVisible()
  return app
}

function userMenuButton(screen: Screen, name = "Ana Example") {
  return screen.getByRole("button", { name: `Menu of ${name}` })
}

/** The image with this address, which is decorative and has no role. */
function imageAt(url: string) {
  return document.querySelector(`img[src="${CSS.escape(url)}"]`)
}

describe("the shell, at a wide width", () => {
  beforeEach(async () => {
    await page.viewport(1280, 800)
  })

  it("shows the mark and the product name, the consultancy, the menu of the user and the work surface", async () => {
    const { screen } = await openTheShell()

    await expect.element(screen.getByText("ForgeSpark")).toBeVisible()
    // The slogan is for the screens seen before a login, not the work.
    await expect
      .element(screen.getByText("From ideas to execution."))
      .not.toBeInTheDocument()
    await expect.element(screen.getByText("Example Consultancy")).toBeVisible()
    await expect.element(userMenuButton(screen)).toBeVisible()
    await expect
      .element(screen.getByRole("main"))
      .toMatchTextContent(
        "The areas of work, such as clients and projects, arrive in the next versions of the system."
      )
  })

  it("shows nothing that does nothing: no search, notifications, help, favourites or settings", async () => {
    const { screen } = await openTheShell()
    await expect.element(userMenuButton(screen)).toBeVisible()

    // The menu of the user is the one control, and the sidebar is no sheet.
    expect(screen.getByRole("button").all()).toHaveLength(1)
    expect(screen.getByRole("searchbox").all()).toHaveLength(0)
    expect(screen.getByRole("link").all()).toHaveLength(0)
  })

  it("shows the initials of the user and of the consultancy when they have no image", async () => {
    const { screen } = await openTheShell(ana, { name: "Fábrica Exemplo" })

    await expect.element(userMenuButton(screen)).toMatchTextContent("AE")
    await expect.element(screen.getByText("FE", { exact: true })).toBeVisible()
    expect(document.querySelector("img")).toBeNull()
  })

  it("shows the photo of the user and the logo of the consultancy, read from the storage", async () => {
    const read = await theStorage()
    const photo = signedURL("0123456789abcdef0123456789abcdef")
    const logo = signedURL("fedcba9876543210fedcba9876543210")

    const { screen } = await openTheShell(
      { ...ana, photo_url: photo },
      { ...theConsultancy, logo_url: logo }
    )

    await expect.poll(() => imageAt(photo)).not.toBeNull()
    await expect.poll(() => imageAt(logo)).not.toBeNull()
    await expect.element(userMenuButton(screen)).not.toMatchTextContent("AE")
    expect(read).toEqual(expect.arrayContaining([photo, logo]))
  })
})

describe("the menu of the user", () => {
  it("shows the name and the e-mail of the user, and the logout, which leads to the login with its notice", async () => {
    const { screen, address } = await openTheShell()
    api.use(getLogOutMockHandler(), nobodyIsLoggedIn(), noLoginProviderIsOn())

    await userMenuButton(screen).click()

    const menu = screen.getByRole("menu")
    await expect.element(menu).toMatchTextContent("Ana Example")
    await expect.element(menu).toMatchTextContent("ana@consultancy.example")
    expect(screen.getByRole("menuitem").all()).toHaveLength(1)

    await screen.getByRole("menuitem", { name: "Log out" }).click()

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent("You logged out")
    expect(address()).toBe("/login?notice=logged_out")
  })

  it("stays open and shows the failure when the logout fails", async () => {
    const { screen, address } = await openTheShell()
    api.use(getLogOutMockHandler500(problem(500, "internal_server_error")))

    await userMenuButton(screen).click()
    await screen.getByRole("menuitem", { name: "Log out" }).click()

    await expect
      .element(screen.getByRole("menu").getByRole("alert"))
      .toMatchTextContent(traceId)
    expect(address()).toBe("/")
  })

  it("opens and logs out by keyboard", async () => {
    await page.viewport(1280, 800)
    const { screen, address } = await openTheShell()
    api.use(getLogOutMockHandler(), nobodyIsLoggedIn(), noLoginProviderIsOn())

    await userEvent.tab()
    await expect.element(userMenuButton(screen)).toHaveFocus()
    await userEvent.keyboard("{Enter}")
    await expect
      .element(screen.getByRole("menuitem", { name: "Log out" }))
      .toHaveFocus()
    await userEvent.keyboard("{Enter}")

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(address()).toBe("/login?notice=logged_out")
  })
})

describe("the language of the shell", () => {
  it("is the preferred language of the user, whatever the browser asks for", async () => {
    const { screen } = await openTheShell(
      { ...ana, preferred_language: "pt-BR" },
      theConsultancy,
      { languages: ["es"] }
    )

    await expect
      .element(screen.getByRole("main"))
      .toMatchTextContent(
        "As áreas de trabalho, como clientes e projetos, chegam nas próximas versões do sistema."
      )
    await screen.getByRole("button", { name: "Menu de Ana Example" }).click()
    await expect
      .element(screen.getByRole("menuitem", { name: "Sair" }))
      .toBeVisible()
    expect(document.documentElement.lang).toBe("pt-BR")
  })
})

describe("the shell, at a narrow width", () => {
  beforeEach(async () => {
    await page.viewport(375, 740)
  })

  it.for([375, 959])(
    "opens the sidebar, %ipx wide, as a sheet from a button at the start of the header",
    async (width) => {
      await page.viewport(width, 740)
      const { screen } = await openTheShell()
      await expect.element(userMenuButton(screen)).toBeVisible()
      await expect
        .element(screen.getByText("Example Consultancy"))
        .not.toBeInTheDocument()

      await screen.getByRole("button", { name: "Open the sidebar" }).click()

      const sheet = screen.getByRole("dialog", { name: "Sidebar" })
      await expect.element(sheet).toBeVisible()
      await expect.element(sheet.getByText("Example Consultancy")).toBeVisible()
    }
  )

  it("keeps the sidebar beside the work surface from 960px", async () => {
    await page.viewport(960, 740)
    const { screen } = await openTheShell()

    await expect.element(screen.getByText("Example Consultancy")).toBeVisible()
    await expect
      .element(screen.getByRole("button", { name: "Open the sidebar" }))
      .not.toBeInTheDocument()
  })

  it("has no page wider than the screen", async () => {
    const { screen } = await openTheShell({
      ...ana,
      name: "Ana Example With A Name Much Longer Than The Header Holds",
    })
    await expect.element(screen.getByRole("main")).toBeVisible()

    expect(document.documentElement.scrollWidth).toBeLessThanOrEqual(375)
  })
})

describe("the images of the shell", () => {
  beforeEach(async () => {
    await page.viewport(1280, 800)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it("are read again before their URLs expire, so they never break", async () => {
    vi.useFakeTimers({ toFake: ["setInterval", "clearInterval"] })
    const read = await theStorage()
    const firstPhoto = signedURL("0123456789abcdef0123456789abcdef", "first")
    const firstLogo = signedURL("fedcba9876543210fedcba9876543210", "first")
    await openTheShell(
      { ...ana, photo_url: firstPhoto },
      { ...theConsultancy, logo_url: firstLogo }
    )
    await expect.poll(() => imageAt(firstPhoto)).not.toBeNull()
    await expect.poll(() => imageAt(firstLogo)).not.toBeNull()

    const renewedPhoto = signedURL(
      "0123456789abcdef0123456789abcdef",
      "renewed"
    )
    const renewedLogo = signedURL("fedcba9876543210fedcba9876543210", "renewed")
    api.use(
      getTellLoggedInUserMockHandler({ ...ana, photo_url: renewedPhoto }),
      getTellConsultancyMockHandler({
        ...theConsultancy,
        logo_url: renewedLogo,
      })
    )
    // Ten minutes go by, which is less than the fifteen a URL works for.
    vi.advanceTimersByTime(10 * 60_000)

    await expect.poll(() => imageAt(renewedPhoto)).not.toBeNull()
    await expect.poll(() => imageAt(renewedLogo)).not.toBeNull()
    expect(imageAt(firstPhoto)).toBeNull()
    expect(read).toEqual(expect.arrayContaining([renewedPhoto, renewedLogo]))
  })
})
