import { execFile } from "node:child_process"
import { join } from "node:path"
import { promisify } from "node:util"

// What the end-to-end test drives besides the browser: the API binary, which
// `task e2e` builds, on the database and the bucket `task e2e` gives the
// test; Mailpit, where the worker sends the login code; and the people Dex,
// the local login provider, knows.

/**
 * The API binary: it serves the API, its `worker` command sends the e-mail,
 * and its `issue-setup-code` command issues setup codes.
 */
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
 * The connection of the API role to the database of the test. The API, the
 * worker and the command get it as `DATABASE_URL`, so none of them touches
 * the development database.
 */
export const databaseURL = required("E2E_DATABASE_URL")

/**
 * The bucket of the test, which the API and the worker get as
 * `FILE_STORAGE_BUCKET`. The worker cleans up the bucket it is given, so the
 * test never shares the one of development (Files in CODING_STANDARDS.md).
 */
export const fileStorageBucket = required("E2E_FILE_STORAGE_BUCKET")

/** Where the browser reads a file of the test: the bucket, at its host name. */
export const fileStorage = `${required("FILE_STORAGE_PUBLIC_URL")}/${fileStorageBucket}`

/** The name the login screen gives the local login provider. */
export const localLoginProvider = required("LOCAL_LOGIN_PROVIDER_DISPLAY_NAME")

/** A person the local login provider knows, with their password. */
export type Person = { email: string; password: string }

/**
 * The people Dex knows (`compose.yaml`, listed in the README): the
 * administrator of the seed, whom the setup of the test registers, and the
 * developer of the seed, who is no user. Fictitious, as everything here.
 */
export const administrator: Person = {
  email: "seed+admin@fabrica.example",
  password: "local-provider-admin",
}
export const developer: Person = {
  email: "seed+developer@fabrica.example",
  password: "local-provider-developer",
}

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

/** The API of Mailpit, which keeps every e-mail of development. */
const mailpit = `http://localhost:${required("MAILPIT_PORT")}/api/v1`

/**
 * The mailbox of an address in Mailpit, from now on: the e-mails it already
 * holds, from the use of development or a run before, are not new. Mailpit
 * keeps them all, so the test takes the first e-mail that arrives after.
 */
export async function mailboxOf(address: string) {
  const held = new Set(await messagesTo(address))
  return {
    /** The six digits of the next login code that arrives, once it does. */
    async nextLoginCode(): Promise<string> {
      const message = await nextMessage(address, held)
      held.add(message.ID)
      if (message.Subject !== "Your login code") {
        throw new Error(`The e-mail that arrived is "${message.Subject}"`)
      }
      const code = /^\s*(\d{6})\s*$/m.exec(message.Text)?.[1]
      if (code === undefined) {
        throw new Error(
          `The e-mail of the login code holds no code:\n${message.Text}`
        )
      }
      return code
    },
  }
}

/** The e-mail Mailpit answers, with its plain-text part. */
type Message = { ID: string; Subject: string; Text: string }

/** Waits for an e-mail to the address that is none of `held`. */
async function nextMessage(
  address: string,
  held: Set<string>
): Promise<Message> {
  const deadline = Date.now() + 30_000
  while (Date.now() < deadline) {
    const arrived = (await messagesTo(address)).find((id) => !held.has(id))
    if (arrived !== undefined) {
      return (await readMailpit(`/message/${arrived}`)) as Message
    }
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
  throw new Error(`No new e-mail reached ${address} in 30 seconds`)
}

/** The identifiers of the e-mails to the address, the newest first. */
async function messagesTo(address: string): Promise<string[]> {
  const query = new URLSearchParams({ query: `to:"${address}"` })
  const { messages } = (await readMailpit(`/search?${query}`)) as {
    messages: { ID: string }[]
  }
  return messages.map((message) => message.ID)
}

async function readMailpit(path: string): Promise<unknown> {
  const answer = await fetch(`${mailpit}${path}`)
  if (!answer.ok) {
    throw new Error(`Mailpit answered ${answer.status} to ${path}`)
  }
  return answer.json()
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
