import { defineConfig } from "playwright/test"

import {
  apiBinary,
  appOrigin,
  databaseURL,
  fileStorageBucket,
  required,
} from "./e2e/system"

// The end-to-end test: the real system, through Caddy, as whoever installs it
// and its users use it. `task e2e` runs it, after it starts the services,
// Mailpit and Dex among them, and creates the empty database and the bucket
// the test runs on (Tests in CODING_STANDARDS.md). Playwright starts the API,
// the worker and Vite where Caddy reaches them, and stops them at the end.

/** What the API and the worker of the test run on in place of development. */
const ofTheTest = {
  DATABASE_URL: databaseURL,
  FILE_STORAGE_BUCKET: fileStorageBucket,
}

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
    browserName: "chromium",
    locale: "en-US",
    timezoneId: "America/Sao_Paulo",
    trace: "retain-on-failure",
    navigationTimeout: 3 * 60_000,
    actionTimeout: 2 * 60_000,
  },
  // The login continues on the database the setup leaves: its administrator
  // is the person Dex knows. It runs only once the setup passed. The files of
  // the brand need neither.
  projects: [
    { name: "setup", testMatch: /setup\.spec\.ts$/ },
    { name: "login", testMatch: /login\.spec\.ts$/, dependencies: ["setup"] },
    { name: "brand", testMatch: /brand\.spec\.ts$/ },
  ],
  // Each of the API and Vite is ready once Caddy reaches it. A server already
  // listening there, such as one of `task dev`, fails the run instead of
  // being used. The worker listens on nothing, and is ready once it says so.
  webServer: [
    {
      name: "API",
      command: `"${apiBinary}"`,
      env: ofTheTest,
      url: `${appOrigin}/up`,
      reuseExistingServer: false,
    },
    {
      name: "Worker",
      command: `"${apiBinary}" worker`,
      env: ofTheTest,
      wait: { stderr: /"msg":"the worker started"/ },
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
