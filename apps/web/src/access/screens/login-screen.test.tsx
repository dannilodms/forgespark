import { describe, expect, it } from "vite-plus/test"

import { getTellSetupStateMockHandler } from "../../shared/api/generated/client.msw"
import { api } from "../../testing/api"
import { renderApp } from "../../testing/render-app"

describe("the provisional login screen", () => {
  it("says the login is not built yet", async () => {
    api.use(
      getTellSetupStateMockHandler({
        concluded: true,
        setup_code_waiting: false,
      })
    )

    const { screen } = await renderApp({ at: "/login" })

    await expect
      .element(screen.getByText(/The login is not built yet/))
      .toBeVisible()
    await expect
      .element(screen.getByText("The setup is concluded"))
      .not.toBeInTheDocument()
  })

  it("says the setup is concluded to whoever arrives from it", async () => {
    api.use(
      getTellSetupStateMockHandler({
        concluded: true,
        setup_code_waiting: false,
      })
    )

    const { screen } = await renderApp({
      at: "/login?notice=setup_concluded",
    })

    await expect
      .element(screen.getByRole("alert"))
      .toMatchTextContent(/The setup is concluded/)
  })
})
