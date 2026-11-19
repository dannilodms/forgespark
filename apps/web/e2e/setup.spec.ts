import { join } from "node:path"

import { expect, test, type Page } from "playwright/test"

import { administrator, issueSetupCode } from "./system"

// The setup, end to end, as whoever installs the system does it: the command
// of the API prints a setup code, and the app, opened through Caddy at the
// address it printed, takes it. The first administrator has the e-mail of the
// administrator Dex knows, so the login (`login.spec.ts`) continues on this
// database. Every name, number and address is fictitious (ADR-0021).

const images = join(import.meta.dirname, "images")

test.describe.serial("the setup", () => {
  test("concludes with a logo and a photo, and leads to the login screen", async ({
    page,
  }) => {
    const first = await issueSetupCode()
    expect(first.opens).toMatch(/^the first setup/)

    await page.goto(first.address)
    await page.getByRole("textbox", { name: "Setup code" }).fill(first.code)
    await page.getByRole("button", { name: "Continue" }).click()

    await expect(
      page.getByRole("heading", { name: "The consultancy" })
    ).toBeVisible()
    await fill(page, {
      Name: "Example Consultancy",
      "Legal name": "Example Consultancy Ltd.",
      "Tax ID": "00.000.000/0001-00",
    })
    await choose(page, "Country", "Braz", "Brazil")
    await sendImage(page, "Logo", "logo.png")
    await page.getByRole("button", { name: "Continue" }).click()

    await expect(
      page.getByRole("heading", { name: "The first administrator" })
    ).toBeVisible()
    await fill(page, {
      Name: "Ada Example",
      // Typed in capitals in part, and kept in lowercase, as Dex gives it.
      "E-mail": "Seed+Admin@Fabrica.example",
      "National ID": "000.000.000-00",
    })
    await choose(page, "Country", "Braz", "Brazil")
    await sendImage(page, "Photo", "photo.jpg")
    await page.getByRole("button", { name: "Continue" }).click()

    await expect(page.getByRole("heading", { name: "Review" })).toBeVisible()
    await expect(page.getByRole("img", { name: "logo.png" })).toBeVisible()
    await expect(page.getByRole("img", { name: "photo.jpg" })).toBeVisible()
    await page.getByRole("button", { name: "Conclude the setup" }).click()

    await expect(page.getByRole("heading", { name: "Log in" })).toBeVisible()
    await expect(page.getByRole("alert")).toContainText(
      "The setup is concluded"
    )
    expect(page.url()).toMatch(/\/login\?notice=setup_concluded$/)
  })

  test("registers a new administrator with a second setup code", async ({
    page,
  }) => {
    const second = await issueSetupCode()
    expect(second.opens).toMatch(/^the registration of a new administrator/)

    await page.goto(second.address)
    await page.getByRole("textbox", { name: "Setup code" }).fill(second.code)
    await page.getByRole("button", { name: "Continue" }).click()

    // The consultancy is left out: there is no step to go back to.
    await expect(
      page.getByRole("heading", { name: "The new administrator" })
    ).toBeVisible()
    await expect(page.getByRole("button", { name: "Back" })).toHaveCount(0)
    await fill(page, {
      Name: "Bruno Example",
      // The e-mail of the first administrator, typed in another case.
      "E-mail": administrator.email.toUpperCase(),
      "National ID": "000 000 000",
    })
    await choose(page, "Country", "Portu", "Portugal")
    await page.getByRole("button", { name: "Continue" }).click()
    await expect(page.getByRole("heading", { name: "Review" })).toBeVisible()
    await page
      .getByRole("button", { name: "Register the administrator" })
      .click()

    // A user has that e-mail: the step shows the refusal on its field.
    const email = page.getByRole("textbox", { name: "E-mail", exact: true })
    await expect(email).toHaveAttribute("aria-invalid", "true")
    await expect(
      page.getByText(
        "This e-mail belongs to a user of the system. Type another one."
      )
    ).toBeVisible()
    await email.fill("Bruno@Consultancy.example")
    await page.getByRole("button", { name: "Continue" }).click()
    await page
      .getByRole("button", { name: "Register the administrator" })
      .click()

    await expect(page.getByRole("heading", { name: "Log in" })).toBeVisible()
    await expect(page.getByRole("alert")).toContainText(
      "The administrator is registered"
    )
    expect(page.url()).toMatch(/\/login\?notice=administrator_registered$/)
  })
})

// Last, because it spends what the tests above left of the limit, which
// counts the requests of the whole machine for a minute. `task e2e` empties
// it before the next run.
test("past the limit, an operation of the setup answers 429 in the error format of the API, and the screen says to wait", async ({
  page,
  request,
}) => {
  // Ten POST a minute under /api/setup, counted together (Rate limit in
  // CODING_STANDARDS.md).
  const limit = 10
  const wrongCode = { data: { code: "00000-00000-00000-00000" } }
  let answer = await request.post("/api/setup/open", wrongCode)
  for (let sent = 1; sent <= limit && answer.status() !== 429; sent++) {
    expect(answer.status()).toBe(422)
    answer = await request.post("/api/setup/open", wrongCode)
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

  // A code that works meets the limit too, and the screen shows the answer.
  const issued = await issueSetupCode()
  await page.goto(issued.address)
  await page.getByRole("textbox", { name: "Setup code" }).fill(issued.code)
  const refused = page.waitForResponse("**/api/setup/open")
  await page.getByRole("button", { name: "Continue" }).click()
  const { trace_id: traceID } = await (await refused).json()
  const alert = page.getByRole("alert")
  await expect(alert).toContainText("Too many attempts from this address")
  await expect(alert).toContainText(traceID)
})

/** Types into each text field named by its label. */
async function fill(page: Page, fields: Record<string, string>) {
  for (const [label, text] of Object.entries(fields)) {
    await page.getByRole("textbox", { name: label, exact: true }).fill(text)
  }
}

/** Searches a list by part of a name, and chooses the item. */
async function choose(page: Page, label: string, typed: string, item: string) {
  await page.getByRole("combobox", { name: label, exact: true }).fill(typed)
  await page.getByRole("option", { name: item, exact: true }).click()
}

/** Chooses an image for its field, and waits until it is sent. */
async function sendImage(page: Page, label: string, file: string) {
  await page
    .getByLabel(label, { exact: true })
    .setInputFiles(join(images, file))
  await expect(
    page.getByRole("status").filter({ hasText: file })
  ).toContainText("Sent")
}
