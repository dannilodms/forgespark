import {
  queryOptions,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"

import {
  concludeSetup,
  openSetup,
  tellSetupState,
} from "../../shared/api/generated/client"
import type {
  ConcludeSetupRequest,
  OpenedSetup,
  SetupAddress,
  SetupState,
} from "../../shared/api/generated/model"
import {
  ConcludeSetupBody,
  OpenSetupBody,
} from "../../shared/api/generated/schemas"
import { answerCodeOf } from "../../shared/errors/problems"
import type { Language } from "../../shared/locales/languages"

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

/**
 * What the person types in the setup. Every field is a text, named as the
 * body names it, so that a field the API refuses is found by its place in
 * the body: `consultancy.address.city`. An optional field left empty is an
 * empty text, which the body leaves out.
 */
export type TypedSetup = {
  consultancy: {
    name: string
    legal_name: string
    country: string
    tax_id: string
    time_zone: string
    currency: string
    email: string
    telephone: string
    website: string
    address: Required<{ [Part in keyof SetupAddress]: string }>
  }
  administrator: {
    name: string
    email: string
    country: string
    national_id: string
    telephone: string
  }
}

/** A setup with nothing typed yet, in a time zone given at the start. */
export function blankSetup(timeZone: string): TypedSetup {
  return {
    consultancy: {
      name: "",
      legal_name: "",
      country: "",
      tax_id: "",
      time_zone: timeZone,
      currency: "",
      email: "",
      telephone: "",
      website: "",
      address: { line1: "", line2: "", city: "", region: "", postal_code: "" },
    },
    administrator: {
      name: "",
      email: "",
      country: "",
      national_id: "",
      telephone: "",
    },
  }
}

/** An optional text: left out of the body when nothing was typed. */
function optional(text: string) {
  return text === "" ? undefined : text
}

/** The address as the body takes it, or nothing when every part is empty. */
function addressOf(typed: TypedSetup["consultancy"]["address"]) {
  const address: SetupAddress = {
    line1: optional(typed.line1),
    line2: optional(typed.line2),
    city: optional(typed.city),
    region: optional(typed.region),
    postal_code: optional(typed.postal_code),
  }
  return Object.values(address).some((part) => part !== undefined)
    ? address
    : undefined
}

/** The consultancy as the body of concluding takes it. */
function consultancyOf(typed: TypedSetup["consultancy"]) {
  return {
    name: typed.name,
    legal_name: typed.legal_name,
    country: typed.country,
    tax_id: typed.tax_id,
    time_zone: typed.time_zone,
    currency: typed.currency,
    email: optional(typed.email),
    telephone: optional(typed.telephone),
    website: optional(typed.website),
    address: addressOf(typed.address),
  }
}

/**
 * The administrator as the body of concluding takes it, but for the
 * preferred language, which is the language of the screen when concluding.
 */
function administratorOf(typed: TypedSetup["administrator"]) {
  return {
    name: typed.name,
    email: typed.email,
    country: typed.country,
    national_id: typed.national_id,
    telephone: optional(typed.telephone),
  }
}

/**
 * The step of the consultancy, as the contract checks it: the schema of the
 * body, and how the typed fields become that part of it.
 */
export const consultancyContract = {
  schema: ConcludeSetupBody.shape.consultancy,
  toBody: consultancyOf,
}

/** The step of the administrator, as the contract checks it. */
export const administratorContract = {
  schema: ConcludeSetupBody.shape.administrator.omit({
    preferred_language: true,
  }),
  toBody: administratorOf,
}

/** What concluding the setup sends, once. */
type Conclusion = {
  typed: TypedSetup
  /** The setup code, as typed. */
  code: string
  /** The language of the screen, the preferred one of the administrator. */
  language: Language
}

/**
 * How concluding ended: the setup is concluded by this request, or someone
 * concluded it before. Any other answer rejects.
 */
type Concluded = "concluded" | "concludedBefore"

/**
 * Concludes the setup: the consultancy, its first administrator and the
 * first roles are saved together, and the setup code stops working. The
 * state the gate reads is updated, so the login screen opens.
 *
 * A setup code that no longer works may have been used by someone who
 * concluded first, which ends it: the state of the setup tells.
 */
export function useConcludeSetup() {
  const queryClient = useQueryClient()

  /** Whether the setup was concluded by someone else, read again. */
  async function concludedBefore(error: unknown) {
    const code = answerCodeOf(error)
    if (code !== "setup_already_concluded" && code !== "setup_code_invalid") {
      return false
    }
    const state = await queryClient.fetchQuery({
      ...setupStateQuery,
      staleTime: 0,
    })
    return state.concluded
  }

  return useMutation({
    mutationFn: async ({
      typed,
      code,
      language,
    }: Conclusion): Promise<Concluded> => {
      const body: ConcludeSetupRequest = {
        code,
        consultancy: consultancyOf(typed.consultancy),
        administrator: {
          ...administratorOf(typed.administrator),
          preferred_language: language,
        },
      }
      try {
        await concludeSetup(body)
      } catch (error) {
        if (await concludedBefore(error)) {
          return "concludedBefore"
        }
        throw error
      }
      // The setup code that concluded it stopped working.
      queryClient.setQueryData(setupStateQuery.queryKey, {
        concluded: true,
        setup_code_waiting: false,
      })
      return "concluded"
    },
  })
}
