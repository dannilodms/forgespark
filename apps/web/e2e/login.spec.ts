import { expect, test, type Locator, type Page } from "playwright/test"

import {
  administrator,
  appOrigin,
  developer,
  fileStorage,
  localLoginProvider,
  mailboxOf,
  type Person,
} from "./system"

// The login, end to end, on the database the setup left (`setup.spec.ts`),
// whose first administrator, Ada Example, has the e-mail of the
// administrator Dex knows: through the local login provider, which is Dex,
// and with a login code the worker sends to Mailpit. Every name and address
// is fictitious (ADR-0021).

/** The first administrator and the consultancy the setup registered. */
const ada = { name: "Ada Example", email: administrator.email }
const consultancy = "Example Consultancy"

test.describe.serial("the login", () => {
  test("through the local login provider leads to the shell of the consultancy, and the logout to the login with its notice", async ({
    page,
  }) => {
    await page.goto("/")
    await expect(page.getByRole("heading", { name: "Log in" })).toBeVisible()

    await logInAtTheProvider(page, administrator)

    await expect(page).toHaveURL(`${appOrigin}/`)
    await expectTheShell(page)

    await page.getByRole("button", { name: `Menu of ${ada.name}` }).click()
    const menu = page.getByRole("menu")
    await expect(menu).toContainText(ada.name)
    await expect(menu).toContainText(ada.email)
    await menu.getByRole("menuitem", { name: "Log out" }).click()

    await expect(page.getByRole("heading", { name: "Log in" })).toBeVisible()
    await expect(page.getByRole("alert")).toContainText("You logged out")
    expect(page.url()).toMatch(/\/login\?notice=logged_out$/)
  })

  test("with a login code read from Mailpit leads to the shell again", async ({
    page,
  }) => {
    const mailbox = await mailboxOf(ada.email)
    await page.goto("/login")

    await page
      .getByRole("textbox", { name: "E-mail", exact: true })
      .fill(ada.email)
    await page.getByRole("button", { name: "Send a code" }).click()
    await expect(
      page.getByText(`If ${ada.email} belongs to a user`)
    ).toBeVisible()

    // The worker sends it, in the preferred language of the administrator.
    const code = await mailbox.nextLoginCode()
    await page.getByRole("textbox", { name: "Login code" }).fill(code)
    await page.getByRole("button", { name: "Log in", exact: true }).click()

    await expect(page).toHaveURL(`${appOrigin}/`)
    await expectTheShell(page)
  })

  test("through the local login provider as a person who is no user leads back to the login, with the notice and their e-mail", async ({
    page,
  }) => {
    await page.goto("/login")

    await logInAtTheProvider(page, developer)

    await expect(page.getByRole("heading", { name: "Log in" })).toBeVisible()
    const notice = page.getByRole("alert")
    await expect(notice).toContainText("No active user has this e-mail")
    await expect(notice).toContainText(developer.email)
    await expect(notice).toContainText("Ask an administrator for access.")
    // The e-mail comes from the login state, never from the address.
    expect(page.url()).toMatch(/\/login\?notice=no_active_user$/)
  })
})

// Last, because it spends what the tests above left of the limit of the
// login code, which counts the requests of the whole machine for a minute,
// apart from the limit of the setup. `task e2e` empties it before the next
// run.
test("past the limit, an operation of the login code answers 429 in the error format of the API, and the screen says to wait", async ({
  page,
  request,
}) => {
  // Ten POST a minute over the two operations of the login code, counted
  // together (Rate limit in CODING_STANDARDS.md). An e-mail of no user is
  // answered 202 and sends nothing; each ask names an address of its own,
  // since the API limits the asks for each address too.
  const limit = 10
  const nobody = (sent: number) => ({
    data: { email: `nobody+${sent}@consultancy.example` },
  })
  let answer = await request.post("/api/session/login_code", nobody(0))
  for (let sent = 1; sent <= limit && answer.status() !== 429; sent++) {
    expect(answer.status()).toBe(202)
    answer = await request.post("/api/session/login_code", nobody(sent))
  }

  expect(answer.status()).toBe(429)
  expect(answer.headers()).toMatchObject({
    "content-type": "application/problem+json",
    "cache-control": "no-store",
    "retry-after": expect.stringMatching(/^[1-9][0-9]*$/),
  })
  expect(await answer.json()).toEqual({
    title: "Too Many Requests",
    status: 429,
    code: "too_many_requests",
    params: {},
    trace_id: expect.stringMatching(/^[0-9a-f]{32}$/),
  })

  // Checking a code shares the limit.
  const check = await request.post("/api/session/login_code/check", {
    data: { code: "000000" },
  })
  expect(check.status()).toBe(429)

  // The screen shows the answer, with its trace identifier.
  await page.goto("/login")
  await page
    .getByRole("textbox", { name: "E-mail", exact: true })
    .fill(ada.email)
  const refused = page.waitForResponse("**/api/session/login_code")
  await page.getByRole("button", { name: "Send a code" }).click()
  const { trace_id: traceID } = await (await refused).json()
  const alert = page.getByRole("alert")
  await expect(alert).toContainText("Too many attempts from this address")
  await expect(alert).toContainText(traceID)
})

/**
 * Logs in through the local login provider: its button on the login screen
 * leads to the login page of Dex, where the person types their e-mail and
 * password, and Dex sends the browser back to the API.
 */
async function logInAtTheProvider(page: Page, person: Person) {
  await page
    .getByRole("link", { name: `Log in with ${localLoginProvider}` })
    .click()
  await page.locator("input[name=login]").fill(person.email)
  await page.locator("input[name=password]").fill(person.password)
  await page.locator("button[type=submit]").click()
}

/**
 * The shell, as Ada sees it once logged in: the menu of the user, whose
 * button is her photo; the sidebar with the name and the logo of the
 * consultancy; and the sentence of the work surface. The two images are the
 * ones sent in the setup, read from the storage of the test through the URLs
 * the API signed.
 */
async function expectTheShell(page: Page) {
  const menu = page.getByRole("button", { name: `Menu of ${ada.name}` })
  await expect(menu).toBeVisible()
  await expect(page.getByText(consultancy, { exact: true })).toBeVisible()
  await expect(
    page.getByText(
      "The areas of work, such as clients and projects, arrive in the next versions of the system."
    )
  ).toBeVisible()

  const user = (await (await page.request.get("/api/session")).json()) as {
    photo_url?: string
  }
  const shown = (await (await page.request.get("/api/consultancy")).json()) as {
    name: string
    logo_url?: string
  }
  expect(shown.name).toBe(consultancy)
  const photo = objectOf(user.photo_url)
  const logo = objectOf(shown.logo_url)
  await expectShown(menu.locator(`img[src^="${photo}?"]`))
  await expectShown(page.locator(`img[src^="${logo}?"]`))
}

/**
 * The object of the bucket of the test a URL signed by the API reads: the
 * URL without its signature, since the API signs a new one at each read.
 */
function objectOf(signedURL: string | undefined) {
  expect(signedURL).toBeDefined()
  const { origin, pathname } = new URL(signedURL ?? "")
  const object = `${origin}${pathname}`
  expect(object.startsWith(`${fileStorage}/`)).toBe(true)
  return object
}

/** An image the page shows, whose bytes the browser read. */
async function expectShown(image: Locator) {
  await expect(image).toBeVisible()
  expect(
    await image.evaluate((element: HTMLImageElement) => element.naturalWidth)
  ).toBeGreaterThan(0)
}
