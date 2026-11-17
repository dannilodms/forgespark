import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vite-plus/test"
import { userEvent } from "vite-plus/test/browser"

import {
  getConcludeSetupMockHandler,
  getOpenSetupMockHandler,
  getStartSetupUploadMockHandler,
  getTellSetupStateMockHandler,
} from "../../shared/api/generated/client.msw"
import type {
  ConcludeSetupRequest,
  StartSetupUploadRequest,
} from "../../shared/api/generated/model"
import { api } from "../../testing/api"
import { renderApp } from "../../testing/render-app"

type Screen = Awaited<ReturnType<typeof renderApp>>["screen"]

/** The code the person types, as the command printed it. */
const typedCode = "ABCDE-FGHJK-MNPQR-STVWX"

/**
 * Opens the setup screen of an install not yet set up and types a setup code
 * that opens the first setup.
 */
async function openTheFirstSetup() {
  api.use(
    getTellSetupStateMockHandler({ concluded: false, setup_code_waiting: true })
  )
  api.use(
    getOpenSetupMockHandler({
      opens: "first_setup",
      expires_at: "2026-11-17T16:00:00Z",
    })
  )
  const opened = await renderApp({ at: "/setup", languages: ["en-US"] })
  const { screen } = opened
  await screen.getByRole("textbox", { name: "Setup code" }).fill(typedCode)
  await screen.getByRole("button", { name: "Continue" }).click()
  await expect
    .element(screen.getByRole("heading", { name: "The consultancy" }))
    .toBeVisible()
  return opened
}

/** An image drawn in the browser, as a file a person chooses. */
async function imageFile(
  name: string,
  type: "image/png" | "image/jpeg" | "image/webp" = "image/png"
) {
  const canvas = new OffscreenCanvas(48, 48)
  const context = canvas.getContext("2d")
  if (context) {
    context.fillStyle = "#1348dc"
    context.fillRect(8, 8, 32, 32)
  }
  const blob = await canvas.convertToBlob({ type })
  return new File([blob], name, { type })
}

/**
 * Chooses a file in the browser's file picker that the button of an image
 * field opens, by the name of that picker.
 */
async function chooseImage(screen: Screen, picker: string, file: File) {
  await userEvent.upload(screen.getByLabelText(picker), file)
}

/** Where the storage is reached, as the API signs a URL for it. */
const storageAddress = "http://files.localhost/software-factory-files"

/**
 * The API starts an upload for each image, answering the identifier given
 * for its purpose and a URL of the storage. Returns every body it received.
 */
function uploadsStart(fileIds: Partial<Record<string, string>>) {
  const started: StartSetupUploadRequest[] = []
  api.use(
    getStartSetupUploadMockHandler(async ({ request }) => {
      const body = await request.json()
      started.push(body)
      const fileId = fileIds[body.purpose] ?? "fil_unknown0"
      return {
        file_id: fileId,
        upload_url: `${storageAddress}/${fileId}?X-Amz-Signature=signed`,
        expires_at: "2026-11-17T15:05:00Z",
      }
    })
  )
  return started
}

/** What the storage received in one PUT. */
type Stored = { url: string; type: string | null; size: number }

/** The file storage, which keeps every image sent to it. */
function theStorage() {
  const stored: Stored[] = []
  api.use(
    http.put(`${storageAddress}/*`, async ({ request }) => {
      stored.push({
        url: request.url,
        type: request.headers.get("Content-Type"),
        size: (await request.arrayBuffer()).byteLength,
      })
      return new HttpResponse(null, { status: 200, headers: { ETag: '"1"' } })
    })
  )
  return stored
}

/** A field of text, by its label. */
function textField(screen: Screen, label: string) {
  return screen.getByRole("textbox", { name: label, exact: true })
}

/** Chooses Brazil in the list of countries of the step. */
async function chooseBrazil(screen: Screen) {
  await screen
    .getByRole("combobox", { name: "Country", exact: true })
    .fill("Braz")
  await screen.getByRole("option", { name: "Brazil", exact: true }).click()
}

/** Types what the consultancy requires, and moves on. */
async function fillTheConsultancy(screen: Screen) {
  await textField(screen, "Name").fill("Acme Software")
  await textField(screen, "Legal name").fill("Acme Software Ltda.")
  await textField(screen, "Tax ID").fill("12.345.678/0001-90")
  await chooseBrazil(screen)
  await screen.getByRole("button", { name: "Continue" }).click()
  await expect
    .element(screen.getByRole("heading", { name: "The first administrator" }))
    .toBeVisible()
}

/** Types what the first administrator requires, and moves on. */
async function fillTheAdministrator(screen: Screen) {
  await textField(screen, "Name").fill("Ana Souza")
  await textField(screen, "E-mail").fill("ana@acme.example")
  await textField(screen, "National ID").fill("123.456.789-09")
  await chooseBrazil(screen)
  await screen.getByRole("button", { name: "Continue" }).click()
  await expect
    .element(screen.getByRole("heading", { name: "Review" }))
    .toBeVisible()
}

/** Every body the API received to conclude the setup, in order. */
function concludingRequests() {
  const sent: ConcludeSetupRequest[] = []
  api.use(
    getConcludeSetupMockHandler(async ({ request }) => {
      sent.push(await request.json())
    })
  )
  return sent
}

/** Concludes the setup from the review, and waits for the login screen. */
async function conclude(screen: Screen) {
  await screen.getByRole("button", { name: "Conclude the setup" }).click()
  await expect
    .element(screen.getByRole("heading", { name: "Log in" }))
    .toBeVisible()
}

describe("an image chosen", () => {
  it("is sent to the storage at once, and shows on its field", async () => {
    const { screen } = await openTheFirstSetup()
    const started = uploadsStart({ consultancy_logo: "fil_logo0001" })
    const stored = theStorage()
    const logo = await imageFile("acme-logo.png")

    await chooseImage(screen, "Choose a logo", logo)

    const shown = screen.getByRole("img", { name: "acme-logo.png" })
    await expect.element(shown).toBeVisible()
    // The browser's own copy of the file.
    await expect
      .element(shown)
      .toHaveAttribute("src", expect.stringMatching(/^blob:/))
    await expect
      .element(screen.getByRole("status"))
      .toMatchTextContent(/acme-logo\.png\s*Sent/)
    expect(started).toEqual([
      {
        code: typedCode,
        purpose: "consultancy_logo",
        media_type: "image/png",
        size: logo.size,
      },
    ])
    expect(stored).toEqual([
      {
        url: `${storageAddress}/fil_logo0001?X-Amz-Signature=signed`,
        type: "image/png",
        size: logo.size,
      },
    ])
  })

  it("shows on the review, and concluding keeps the logo and the photo by their files", async () => {
    const { screen } = await openTheFirstSetup()
    const started = uploadsStart({
      consultancy_logo: "fil_logo0001",
      user_photo: "fil_photo001",
    })
    theStorage()
    await chooseImage(screen, "Choose a logo", await imageFile("acme-logo.png"))
    await fillTheConsultancy(screen)
    const photo = await imageFile("ana.jpg", "image/jpeg")
    await chooseImage(screen, "Choose a photo", photo)
    await fillTheAdministrator(screen)

    await expect
      .element(
        screen
          .getByRole("region", { name: "The consultancy" })
          .getByRole("img", { name: "acme-logo.png" })
      )
      .toBeVisible()
    await expect
      .element(
        screen
          .getByRole("region", { name: "The first administrator" })
          .getByRole("img", { name: "ana.jpg" })
      )
      .toBeVisible()
    const sent = concludingRequests()
    await conclude(screen)

    expect(started[1]).toEqual({
      code: typedCode,
      purpose: "user_photo",
      media_type: "image/jpeg",
      size: photo.size,
    })
    expect(sent).toHaveLength(1)
    expect(sent[0]?.consultancy.logo_file_id).toBe("fil_logo0001")
    expect(sent[0]?.administrator.photo_file_id).toBe("fil_photo001")
  })
})

describe("a file refused when it is chosen", () => {
  /** The button of the field of the logo, which its problem describes. */
  function logoButton(screen: Screen) {
    return screen.getByRole("button", { name: "Choose a logo" })
  }

  it("is a file of another type: the field says why, and nothing is sent", async () => {
    const { screen } = await openTheFirstSetup()
    const started = uploadsStart({ consultancy_logo: "fil_logo0001" })
    const stored = theStorage()

    await chooseImage(
      screen,
      "Choose a logo",
      new File(["%PDF-1.7"], "logo.pdf", { type: "application/pdf" })
    )

    await expect
      .element(logoButton(screen))
      .toHaveAccessibleDescription(
        /This file is not a PNG, a JPEG or a WebP image\. Choose one that is\./
      )
    await expect.element(screen.getByRole("img")).not.toBeInTheDocument()
    expect(started).toEqual([])
    expect(stored).toEqual([])
  })

  it("is an image over 2 MB: the field says why, and nothing is sent", async () => {
    const { screen } = await openTheFirstSetup()
    const started = uploadsStart({ consultancy_logo: "fil_logo0001" })
    const stored = theStorage()

    await chooseImage(
      screen,
      "Choose a logo",
      new File([new Uint8Array(2 * 1024 * 1024 + 1)], "huge.png", {
        type: "image/png",
      })
    )

    await expect
      .element(logoButton(screen))
      .toHaveAccessibleDescription(
        /This file is larger than 2 MB\. Choose a smaller image\./
      )
    expect(started).toEqual([])
    expect(stored).toEqual([])
  })

  it("is an empty file: the field says so, and nothing is sent", async () => {
    const { screen } = await openTheFirstSetup()
    const started = uploadsStart({ consultancy_logo: "fil_logo0001" })

    await chooseImage(
      screen,
      "Choose a logo",
      new File([], "empty.png", { type: "image/png" })
    )

    await expect
      .element(logoButton(screen))
      .toHaveAccessibleDescription(/This file is empty\. Choose another\./)
    expect(started).toEqual([])
  })

  it("leaves the image chosen before as it was", async () => {
    const { screen } = await openTheFirstSetup()
    const started = uploadsStart({ consultancy_logo: "fil_logo0001" })
    theStorage()
    await chooseImage(screen, "Choose a logo", await imageFile("acme-logo.png"))
    await expect.element(screen.getByRole("status")).toMatchTextContent(/Sent/)

    await chooseImage(
      screen,
      "Choose a logo",
      new File(["GIF89a"], "logo.gif", { type: "image/gif" })
    )

    await expect
      .element(screen.getByRole("button", { name: "Replace the logo" }))
      .toHaveAccessibleDescription(/This file is not a PNG/)
    await expect
      .element(screen.getByRole("img", { name: "acme-logo.png" }))
      .toBeVisible()
    expect(started).toHaveLength(1)
  })
})
