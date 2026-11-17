import { defineConfig } from "playwright/test"

import { apiBinary, appOrigin, databaseURL, required } from "./e2e/system"

// The end-to-end test: the real system, through Caddy, as whoever installs it
// uses it. `task e2e` runs it, after it starts the services and creates the
// empty database the test runs on (Tests in CODING_STANDARDS.md). Playwright
// starts the API and Vite where Caddy reaches them, and stops them at the end.
export default defineConfig({
  testDir: "e2e",
  // The limit of Caddy counts every request of the machine together, so the
  // tests run one at a time, in the order they are written, and a failure is
  // not run again.
  workers: 1,
  retries: 0,
  forbidOnly: Boolean(process.env.CI),
  // Vite bundles the dependencies of the app on its first load.
  timeout: 5 * 60_000,
  expect: { timeout: 30_000 },
  reporter: "list",
  // What a failure leaves to read: the trace of the test, which the pipeline
  // keeps.
  outputDir: "test-results",
  use: {
    baseURL: appOrigin,
    locale: "en-US",
    timezoneId: "America/Sao_Paulo",
    trace: "retain-on-failure",
    navigationTimeout: 3 * 60_000,
    actionTimeout: 2 * 60_000,
  },
  projects: [{ name: "chromium", use: { browserName: "chromium" } }],
  // Each is ready once Caddy reaches it. A server already listening there,
  // such as one of `task dev`, fails the run instead of being used.
  webServer: [
    {
      name: "API",
      command: `"${apiBinary}"`,
      env: { DATABASE_URL: databaseURL },
      url: `${appOrigin}/up`,
      reuseExistingServer: false,
    },
    {
      name: "Vite",
      // As `task web` starts it.
      command: `vp dev --host "${required("WEB_LISTEN_HOST")}" --port 5173 --strictPort`,
      url: `${appOrigin}/`,
      reuseExistingServer: false,
    },
  ],
})
