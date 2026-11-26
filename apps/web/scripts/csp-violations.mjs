// Fails when the built app, served with the Content-Security-Policy of the
// servers, makes the browser report a violation of the policy, and names each
// one. The policy comes from its one home, the snippet `security-headers` of
// deploy/delivery.caddyfile (Security headers in CODING_STANDARDS.md).
//
// It opens the login screen of the build in apps/web/dist, which `task
// web:build` writes, in the Chromium of the Playwright in the catalog. Every
// request is answered here, as Caddy and the API would answer it: the files
// of the build, a route of the app with index.html, and an API whose setup is
// concluded and where nobody is logged in. Nothing listens on a port, and no
// request leaves the machine.
//
// Run from `apps/web`, after the build: `vp node scripts/csp-violations.mjs`.
import { existsSync, readFileSync, statSync } from "node:fs"
import { extname, join, resolve } from "node:path"

import { chromium } from "playwright"

const appDir = resolve(import.meta.dirname, "..")
const builtApp = join(appDir, "dist")
const snippets = resolve(appDir, "../../deploy/delivery.caddyfile")

// Fictitious origins (ADR-0021): every request to them is answered below.
const appOrigin = "https://app.example"
const fileStorageOrigin = "https://files.example"

/** The policy the servers send, with the host name of the file storage. */
function policyOfTheServers() {
  const policy = /Content-Security-Policy "([^"]+)"/.exec(
    readFileSync(snippets, "utf8")
  )?.[1]
  if (!policy) {
    throw new Error(`${snippets} sets no Content-Security-Policy.`)
  }
  return policy.replaceAll("{args[0]}", fileStorageOrigin)
}

const contentTypes = {
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".json": "application/json",
  ".svg": "image/svg+xml",
  ".woff": "font/woff",
  ".woff2": "font/woff2",
}

/** What the API answers while the setup is concluded and nobody logged in. */
const answersOfTheAPI = {
  "/api/setup": {
    status: 200,
    body: { concluded: true, setup_code_waiting: false },
  },
  "/api/session": {
    status: 401,
    body: {
      status: 401,
      code: "unauthorized",
      title: "Unauthorized",
      params: {},
      trace_id: "0".repeat(32),
    },
  },
  "/api/session/login": { status: 200, body: { login_providers: [] } },
}

/** The file of the build a path names, or index.html for a route. */
function fileOf(pathname) {
  const file = join(builtApp, pathname)
  if (pathname !== "/" && existsSync(file) && statSync(file).isFile()) {
    return file
  }
  return join(builtApp, "index.html")
}

if (!existsSync(join(builtApp, "index.html"))) {
  console.error("apps/web/dist holds no build: run task web:build first.")
  process.exit(1)
}

const policy = policyOfTheServers()
const browser = await chromium.launch()
let violations = []
try {
  const page = await browser.newPage()
  // Recorded from the first script of the page on. Playwright adds this one
  // past the policy, which it does not govern.
  await page.addInitScript(() => {
    window.violations = []
    document.addEventListener(
      "securitypolicyviolation",
      (event) =>
        window.violations.push(
          `${event.effectiveDirective} refused ${event.blockedURI || "inline"}, in ${event.sourceFile}:${event.lineNumber}`
        ),
      true
    )
  })
  await page.route("**/*", async (route) => {
    const url = new URL(route.request().url())
    if (url.origin !== appOrigin) {
      await route.abort()
      return
    }
    const answer = answersOfTheAPI[url.pathname]
    if (answer) {
      await route.fulfill({
        status: answer.status,
        contentType:
          answer.status < 400 ? "application/json" : "application/problem+json",
        headers: { "Content-Security-Policy": policy },
        body: JSON.stringify(answer.body),
      })
      return
    }
    const file = fileOf(url.pathname)
    await route.fulfill({
      status: 200,
      contentType: contentTypes[extname(file)] ?? "application/octet-stream",
      headers: { "Content-Security-Policy": policy },
      body: readFileSync(file),
    })
  })
  await page.goto(`${appOrigin}/`)
  // The login screen, once the gate of the app has read the setup and the
  // session: the app ran, so the policy was met by all of it.
  await page.locator('input[type="email"]').waitFor()
  await page.waitForLoadState("networkidle")
  violations = await page.evaluate(() => window.violations)
} finally {
  await browser.close()
}

if (violations.length > 0) {
  console.error(
    `The built app made the browser report ${violations.length} violation(s) of the Content-Security-Policy of the servers:\n` +
      violations.map((violation) => `  ${violation}`).join("\n")
  )
  process.exit(1)
}

console.log(
  "Content-Security-Policy: the login screen of the built app opened with no violation of the policy of the servers."
)
