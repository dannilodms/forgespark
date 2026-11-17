import { queryOptions, useMutation } from "@tanstack/react-query"

import { openSetup, tellSetupState } from "../../shared/api/generated/client"
import type { OpenedSetup, SetupState } from "../../shared/api/generated/model"
import { OpenSetupBody } from "../../shared/api/generated/schemas"

// The reads and writes of the setup (Data module in CODING_STANDARDS.md).
// The setup has no collection: its one read goes through TanStack Query.

/**
 * Whether the setup is concluded and whether a setup code is waiting. The
 * gate of the app reads it before any screen opens.
 */
export const setupStateQuery = queryOptions({
  queryKey: ["setup"],
  queryFn: () => tellSetupState(),
})

/**
 * Whether an address leads to the setup (Setup in CODING_STANDARDS.md):
 * every address does until the setup is concluded, and afterwards only its
 * own address, while a setup code is waiting.
 */
export function leadsToTheSetup(state: SetupState, atTheSetup: boolean) {
  return !state.concluded || (atTheSetup && state.setup_code_waiting)
}

/**
 * A setup code that opened the setup: the code as typed, what it opens and
 * when it stops working. It is held in the memory of the screen only, and
 * each request of the setup sends the code in its body.
 */
export type OpenedSetupCode = OpenedSetup & { code: string }

/** What the screen sends to open the setup, as the contract checks it. */
export const setupCodeContract = OpenSetupBody

/**
 * Opens the setup with a setup code: it answers what the code opens and when
 * it expires, and saves nothing. The code travels in the body only.
 */
export function useOpenSetup() {
  return useMutation({
    mutationFn: async (code: string): Promise<OpenedSetupCode> => ({
      ...(await openSetup({ code })),
      code,
    }),
  })
}
