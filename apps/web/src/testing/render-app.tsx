import { createMemoryHistory } from "@tanstack/react-router"
import { render } from "vitest-browser-react"

import { createApp } from "../app"

type Visit = {
  /** The address the person opens. */
  at?: string
  /** The languages their browser asks for. */
  languages?: readonly string[]
}

/**
 * Opens the whole app at an address, as a person does in a new tab: a new
 * router, nothing in memory, the routes and the gate in force. The API
 * answers as the test told `api` to.
 */
export async function renderApp({ at = "/", languages = ["en-US"] }: Visit) {
  const history = createMemoryHistory({ initialEntries: [at] })
  const App = createApp({ languages, history })
  const screen = await render(<App />)

  return {
    screen,
    /** The address in the bar of the browser: path, search and hash. */
    address: () => history.location.href,
  }
}
