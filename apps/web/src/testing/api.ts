import { setupWorker } from "msw/browser"

import { getTellLoggedInUserMockHandler401 } from "../shared/api/generated/client.msw"
import type { Problem } from "../shared/api/generated/model"

/**
 * The API as a screen test simulates it, at the network, with MSW. It
 * answers nothing until a test hands it the answers it needs, built from the
 * handlers orval generates: `api.use(getTellSetupStateMockHandler({...}))`.
 * A request no answer was given for fails the test. `setup.ts` starts it for
 * every test file and forgets the answers after each test.
 */
export const api = setupWorker()

/** A trace identifier, as the API writes it: 32 hexadecimal characters. */
export const traceId = "4bf92f3577b34da6a3ce929d0e0e4736"

/**
 * A failure in the error format of the API, for the handler of the status
 * that answers it: `getOpenSetupMockHandler422(problem(422, "..."))`.
 */
export function problem(
  status: number,
  code: string,
  more: Partial<Problem> = {}
): Problem {
  return { title: "", status, code, params: {}, trace_id: traceId, ...more }
}

/**
 * The answer to who is logged in in a browser with no session, or whose
 * session ended: 401 `unauthorized`. The gate reads it on every address but
 * the setup's, so a test that reaches the login screen hands it to the API:
 * `api.use(nobodyIsLoggedIn())`.
 */
export function nobodyIsLoggedIn() {
  return getTellLoggedInUserMockHandler401(problem(401, "unauthorized"))
}
