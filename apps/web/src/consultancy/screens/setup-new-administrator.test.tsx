import { HttpResponse, http } from "msw"
import { describe, expect, it } from "vite-plus/test"
import { userEvent } from "vite-plus/test/browser"

import {
  getOpenSetupMockHandler,
  getRegisterAdministratorMockHandler,
  getRegisterAdministratorMockHandler409,
  getRegisterAdministratorMockHandler422,
  getStartSetupUploadMockHandler,
  getTellSetupStateMockHandler,
} from "../../shared/api/generated/client.msw"
import type { RegisterAdministratorRequest } from "../../shared/api/generated/model"
import { api, problem, traceId } from "../../testing/api"
import { renderApp } from "../../testing/render-app"

type Screen = Awaited<ReturnType<typeof renderApp>>["screen"]

/** The code the person types, as the command printed it. */
const typedCode = "ABCDE-FGHJK-MNPQR-STVWX"

/**
 * Opens the address of the setup of an install whose setup is concluded,
 * while a setup code is waiting, and types the code, which opens the
 * registration of a new administrator: the way back in.
 */
async function openTheWayBackIn(languages: readonly string[] = ["en-US"]) {
  api.use(
    getTellSetupStateMockHandler({ concluded: true, setup_code_waiting: true })
  )
  api.use(
    getOpenSetupMockHandler({
      opens: "new_administrator",
      expires_at: "2026-11-17T16:00:00Z",
    })
  )
  const opened = await renderApp({ at: "/setup", languages })
  const { screen } = opened
  await screen
    .getByRole("textbox", { name: /Setup code|Código de configuraç/ })
    .fill(typedCode)
  await screen.getByRole("button", { name: /Continue|Continuar/ }).click()
  return opened
}

/** A field of text, by its label. */
function textField(screen: Screen, label: string) {
  return screen.getByRole("textbox", { name: label, exact: true })
}

/** Every field of the new administrator as typed. */
const administrator = {
  Name: "Bruno Lima",
  "E-mail": "Bruno@Acme.example",
  "National ID": "123 456 789",
  "Telephone (optional)": "+351 21 000 1234",
}

/** Types the whole new administrator, of Portugal, and moves on. */
async function fillTheAdministrator(screen: Screen) {
  for (const [label, text] of Object.entries(administrator)) {
    await textField(screen, label).fill(text)
  }
  await screen
    .getByRole("combobox", { name: "Country", exact: true })
    .fill("Portu")
  await screen.getByRole("option", { name: "Portugal", exact: true }).click()
  await screen.getByRole("button", { name: "Continue" }).click()
}

/** Where the storage is reached, as the API signs a URL for it. */
const storageAddress = "http://files.localhost/software-factory-files"

/** An image drawn in the browser, as a file a person chooses. */
async function imageFile(name: string) {
  const canvas = new OffscreenCanvas(48, 48)
  const context = canvas.getContext("2d")
  if (context) {
    context.fillStyle = "#1348dc"
    context.fillRect(8, 8, 32, 32)
  }
  const blob = await canvas.convertToBlob({ type: "image/png" })
  return new File([blob], name, { type: "image/png" })
}

/** Every body the API received to register an administrator, in order. */
function registeringRequests() {
  const sent: unknown[] = []
  api.use(
    getRegisterAdministratorMockHandler(async ({ request }) => {
      sent.push(await request.json())
    })
  )
  return sent
}

/** The body the administrator typed above is registered with. */
function registeredAdministrator(
  language: RegisterAdministratorRequest["administrator"]["preferred_language"]
): RegisterAdministratorRequest {
  return {
    code: typedCode,
    administrator: {
      name: administrator.Name,
      email: administrator["E-mail"],
      country: "PT",
      national_id: administrator["National ID"],
      telephone: administrator["Telephone (optional)"],
      preferred_language: language,
    },
  }
}

describe("the way back in", () => {
  it("leaves out the consultancy: from the code to the new administrator, the review and the login screen", async () => {
    const { screen, address } = await openTheWayBackIn()

    await expect
      .element(screen.getByRole("heading", { name: "The new administrator" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("button", { name: "Back" }))
      .not.toBeInTheDocument()
    await fillTheAdministrator(screen)

    await expect
      .element(screen.getByRole("heading", { name: "Review" }))
      .toBeVisible()
    await expect
      .element(screen.getByText(/A new administrator will be registered/))
      .toMatchTextContent(/the users that exist are not changed/)
    await expect
      .element(screen.getByRole("region", { name: "The consultancy" }))
      .not.toBeInTheDocument()
    const ofTheAdministrator = screen.getByRole("region", {
      name: "The new administrator",
    })
    for (const typed of [...Object.values(administrator), "Portugal"]) {
      await expect.element(ofTheAdministrator).toMatchTextContent(typed)
    }
    const sent = registeringRequests()

    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The administrator is registered/)
    expect(address()).toBe("/login?notice=administrator_registered")
    expect(sent).toEqual([registeredAdministrator("en-US")])
  })
})

describe("the e-mail of a user that exists", () => {
  it("shows on its field, on the step of the new administrator, as typed", async () => {
    const { screen } = await openTheWayBackIn()
    await fillTheAdministrator(screen)
    api.use(
      getRegisterAdministratorMockHandler409(problem(409, "email_in_use"))
    )

    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    await expect
      .element(screen.getByRole("heading", { name: "The new administrator" }))
      .toBeVisible()
    const email = textField(screen, "E-mail")
    await expect
      .element(email)
      .toHaveAccessibleDescription(/This e-mail belongs to a user/)
    await expect.element(email).toHaveAttribute("aria-invalid", "true")
    await expect.element(email).toHaveFocus()
    await expect.element(email).toHaveValue(administrator["E-mail"])
    await expect
      .element(email)
      .toHaveAccessibleDescription(
        new RegExp(`Identifier to pass on: ?${traceId}`)
      )
  })

  it("is changed, and the administrator is registered with the new one", async () => {
    const { screen } = await openTheWayBackIn()
    await fillTheAdministrator(screen)
    api.use(
      getRegisterAdministratorMockHandler409(problem(409, "email_in_use"), {
        once: true,
      })
    )
    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()
    const sent = registeringRequests()

    await textField(screen, "E-mail").fill("bruno.lima@acme.example")
    await screen.getByRole("button", { name: "Continue" }).click()
    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(sent).toHaveLength(1)
    expect(sent[0]).toMatchObject({
      administrator: { email: "bruno.lima@acme.example" },
    })
  })
})

describe("the language of the way back in", () => {
  it("is the preferred language of the new administrator", async () => {
    const { screen } = await openTheWayBackIn(["pt-BR"])
    await expect
      .element(screen.getByRole("heading", { name: "O novo administrador" }))
      .toBeVisible()
    await screen.getByRole("combobox", { name: "Idioma" }).click()
    await screen.getByRole("option", { name: "English" }).click()
    await fillTheAdministrator(screen)
    const sent = registeringRequests()

    await screen.getByRole("combobox", { name: "Language" }).click()
    await screen.getByRole("option", { name: "Español" }).click()
    await expect
      .element(screen.getByText(/Se registrará un nuevo administrador/))
      .toMatchTextContent(/los usuarios que existen no se modifican/)
    await screen
      .getByRole("button", { name: "Registrar al administrador" })
      .click()

    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/El administrador está registrado/)
    expect(sent).toEqual([registeredAdministrator("es")])
  })
})

describe("the photo of the new administrator", () => {
  it("is sent under the setup code and registered with the administrator", async () => {
    const { screen } = await openTheWayBackIn()
    const started: unknown[] = []
    api.use(
      getStartSetupUploadMockHandler(async ({ request }) => {
        started.push(await request.json())
        return {
          file_id: "fil_photo001",
          upload_url: `${storageAddress}/fil_photo001?X-Amz-Signature=signed`,
          expires_at: "2026-11-17T15:05:00Z",
        }
      })
    )
    api.use(
      http.put(
        `${storageAddress}/*`,
        () => new HttpResponse(null, { status: 200, headers: { ETag: '"1"' } })
      )
    )
    const photo = await imageFile("bruno.png")
    await userEvent.upload(screen.getByLabelText("Photo"), photo)
    await fillTheAdministrator(screen)
    await expect
      .element(
        screen
          .getByRole("region", { name: "The new administrator" })
          .getByRole("img", { name: "bruno.png" })
      )
      .toBeVisible()
    const sent = registeringRequests()

    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(started).toEqual([
      {
        code: typedCode,
        purpose: "user_photo",
        media_type: "image/png",
        size: photo.size,
      },
    ])
    expect(sent).toEqual([
      {
        ...registeredAdministrator("en-US"),
        administrator: {
          ...registeredAdministrator("en-US").administrator,
          photo_file_id: "fil_photo001",
        },
      },
    ])
  })
})

describe("a setup code that stops working on the way back in", () => {
  it("returns to the step of the code, and a new code resumes at the review of the new administrator", async () => {
    const { screen } = await openTheWayBackIn()
    await fillTheAdministrator(screen)
    api.use(
      getRegisterAdministratorMockHandler422(
        problem(422, "setup_code_invalid"),
        { once: true }
      )
    )

    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    const code = textField(screen, "Setup code")
    await expect.element(code).toHaveFocus()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The setup code stopped working/)
    const sent = registeringRequests()
    await code.fill("QRSTV-WXYZ0-12345-6789A")
    await screen.getByRole("button", { name: "Continue" }).click()
    await expect
      .element(screen.getByRole("heading", { name: "Review" }))
      .toBeVisible()
    await expect
      .element(screen.getByRole("region", { name: "The consultancy" }))
      .not.toBeInTheDocument()
    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()
    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    expect(sent).toEqual([
      { ...registeredAdministrator("en-US"), code: "QRSTV-WXYZ0-12345-6789A" },
    ])
  })
})

describe("a registration whose answer was lost", () => {
  /** Registers, and the network fails before any answer arrives. */
  async function registerWithoutAnswer(screen: Screen) {
    api.use(
      http.post("*/api/setup/administrators", () => HttpResponse.error(), {
        once: true,
      })
    )
    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The server did not answer/)
  }

  /** What the login screen says, read whole. */
  async function expectTheNoticeOfALostAnswer(screen: Screen) {
    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
    const notice = screen.getByRole("alert")
    await expect
      .element(notice)
      .toMatchTextContent(/The administrator may be registered/)
    await expect
      .element(notice)
      .toMatchTextContent(/try logging in here with that e-mail/)
    await expect.element(notice).not.toMatchTextContent(/not saved|nothing/i)
  }

  it("leads to the login screen, which says the administrator may be registered, when a later attempt finds the setup code ended", async () => {
    const { screen, address } = await openTheWayBackIn()
    await fillTheAdministrator(screen)
    await registerWithoutAnswer(screen)
    // The first attempt may have registered the administrator, which ends
    // the setup code.
    api.use(
      getRegisterAdministratorMockHandler422(problem(422, "setup_code_invalid"))
    )

    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    await expectTheNoticeOfALostAnswer(screen)
    expect(address()).toBe(
      "/login?notice=administrator_registered_without_answer"
    )
  })

  it("leads there too when a later attempt finds the e-mail typed then in use, in any case", async () => {
    const { screen, address } = await openTheWayBackIn()
    await fillTheAdministrator(screen)
    // Caddy answers for an API that did not answer in time.
    api.use(
      http.post(
        "*/api/setup/administrators",
        () =>
          HttpResponse.json(problem(504, "gateway_timeout"), { status: 504 }),
        { once: true }
      )
    )
    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()
    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The server took too long to answer/)
    await screen.getByRole("button", { name: "Back" }).click()
    await textField(screen, "E-mail").fill("bruno@acme.example")
    await screen.getByRole("button", { name: "Continue" }).click()
    api.use(
      getRegisterAdministratorMockHandler409(problem(409, "email_in_use"))
    )

    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    await expectTheNoticeOfALostAnswer(screen)
    expect(address()).toBe(
      "/login?notice=administrator_registered_without_answer"
    )
  })

  it("still shows an e-mail in use on its field when it is not the one of the attempt that got no answer", async () => {
    const { screen, address } = await openTheWayBackIn()
    await fillTheAdministrator(screen)
    await registerWithoutAnswer(screen)
    await screen.getByRole("button", { name: "Back" }).click()
    await textField(screen, "E-mail").fill("bruno.lima@acme.example")
    await screen.getByRole("button", { name: "Continue" }).click()
    api.use(
      getRegisterAdministratorMockHandler409(problem(409, "email_in_use"))
    )

    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    const email = textField(screen, "E-mail")
    await expect
      .element(email)
      .toHaveAccessibleDescription(/This e-mail belongs to a user/)
    await expect.element(email).toHaveValue("bruno.lima@acme.example")
    expect(address()).toBe("/setup")
  })

  it("is told in the language of the screen", async () => {
    const { screen } = await openTheWayBackIn(["pt-BR"])
    await screen.getByRole("combobox", { name: "Idioma" }).click()
    await screen.getByRole("option", { name: "English" }).click()
    await fillTheAdministrator(screen)
    await registerWithoutAnswer(screen)
    await screen.getByRole("combobox", { name: "Language" }).click()
    await screen.getByRole("option", { name: "Português (Brasil)" }).click()
    api.use(
      getRegisterAdministratorMockHandler422(problem(422, "setup_code_invalid"))
    )

    await screen
      .getByRole("button", { name: "Cadastrar o administrador" })
      .click()

    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/O administrador pode ter sido cadastrado/)
  })
})

describe("the address of the setup once the administrator is registered", () => {
  it("leads to the login screen: the setup code is used", async () => {
    const { screen, address, goTo } = await openTheWayBackIn()
    await fillTheAdministrator(screen)
    registeringRequests()
    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()
    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()

    goTo("/setup")

    await expect.poll(address).toBe("/login")
    await expect
      .element(screen.getByRole("heading", { name: "Log in" }))
      .toBeVisible()
  })
})

describe("an install whose setup is not concluded", () => {
  it("is told above the review, with the trace identifier", async () => {
    const { screen } = await openTheWayBackIn()
    await fillTheAdministrator(screen)
    api.use(
      getRegisterAdministratorMockHandler409(
        problem(409, "setup_not_concluded")
      )
    )

    await screen
      .getByRole("button", { name: "Register the administrator" })
      .click()

    const alert = screen.getByRole("alert")
    await expect.element(alert).toMatchTextContent(/The setup is not concluded/)
    await expect.element(alert).toMatchTextContent(traceId)
    await expect
      .element(screen.getByRole("heading", { name: "Review" }))
      .toBeVisible()
  })
})
