import { afterAll, afterEach, beforeAll } from "vite-plus/test"

import "../ui/styles/globals.css"
import { api } from "./api"

// Runs before every test file (setupFiles in vite.config.ts).

beforeAll(async () => {
  await api.start({ quiet: true, onUnhandledFrame: "error" })
})

afterEach(() => {
  api.resetHandlers()
})

afterAll(async () => {
  await api.stop()
})
