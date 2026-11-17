import { setupWorker } from "msw/browser"

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
