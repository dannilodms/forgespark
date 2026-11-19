import { setupWorker } from "msw/browser"

import {
  getTellConsultancyMockHandler,
  getTellLoggedInUserMockHandler,
  getTellLoggedInUserMockHandler401,
  getTellLoginStateMockHandler,
} from "../shared/api/generated/client.msw"
import type {
  Consultancy,
  LoggedInUser,
  Problem,
} from "../shared/api/generated/model"

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

/**
 * The login state of a browser where no login provider is turned on and no
 * login through one was refused. The login screen reads it when it opens, so
 * a test that reaches the login screen hands it to the API beside
 * `nobodyIsLoggedIn()`. A test of the providers hands its own:
 * `getTellLoginStateMockHandler({ login_providers: [...] })`.
 */
export function noLoginProviderIsOn() {
  return getTellLoginStateMockHandler({ login_providers: [] })
}

/** The consultancy of the tests, as the API answers it: it has no logo. */
export const theConsultancy: Consultancy = { name: "Example Consultancy" }

/**
 * The answers of the API once `user` is logged in: who they are, and the
 * consultancy, which the shell of the app reads before it opens. A test that
 * reaches the shell hands both to the API: `api.use(...loggedIn(ana))`.
 */
export function loggedIn(user: LoggedInUser, consultancy = theConsultancy) {
  return [
    getTellLoggedInUserMockHandler(user),
    getTellConsultancyMockHandler(consultancy),
  ]
}
