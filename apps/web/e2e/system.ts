import { execFile } from "node:child_process"
import { join } from "node:path"
import { promisify } from "node:util"

// What the end-to-end test drives besides the browser: the API binary, which
// `task e2e` builds, on the database `task e2e` creates empty for the test.

/** The API binary: it serves the API, and its command issues setup codes. */
export const apiBinary = join(
  import.meta.dirname,
  "..",
  "..",
  "api",
  "bin",
  process.platform === "win32" ? "api.exe" : "api"
)

/** Where the test opens the app: Caddy, at the origin of development. */
export const appOrigin = required("APP_ORIGIN")

/**
 * The connection of the API role to the database of the test. The API and
 * its command get it as `DATABASE_URL`, so neither touches the development
 * database.
 */
export const databaseURL = required("E2E_DATABASE_URL")

/** What `issue-setup-code` printed: the code, what it opens, the address. */
type IssuedSetupCode = { code: string; opens: string; address: string }

/** Issues a setup code with the command, as whoever installs the system does. */
export async function issueSetupCode(): Promise<IssuedSetupCode> {
  const { stdout } = await promisify(execFile)(
    apiBinary,
    ["issue-setup-code"],
    { env: { ...process.env, DATABASE_URL: databaseURL } }
  )
  return {
    code: printed(stdout, "Setup code"),
    opens: printed(stdout, "It opens"),
    address: printed(stdout, "Address"),
  }
}

/** The value the command printed after a label, on the same line. */
function printed(output: string, label: string): string {
  const line = output.split("\n").find((text) => text.startsWith(`${label}:`))
  if (line === undefined) {
    throw new Error(`issue-setup-code printed no "${label}":\n${output}`)
  }
  return line.slice(label.length + 1).trim()
}

/** A variable `task e2e` gives the test, from `.env.example` or its own. */
export function required(name: string): string {
  const value = process.env[name]
  if (!value) {
    throw new Error(
      `${name} is not set. Run the end-to-end test with \`task e2e\`, which sets it.`
    )
  }
  return value
}
