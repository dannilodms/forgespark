import {
  queryOptions,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"

import {
  concludeSetup,
  openSetup,
  registerAdministrator,
  startSetupUpload,
  tellSetupState,
} from "../../shared/api/generated/client"
import {
  type ConcludeSetupRequest,
  type OpenedSetup,
  type SetupAddress,
  type SetupState,
  StartSetupUploadRequestMediaType,
  type StartSetupUploadRequestPurpose,
} from "../../shared/api/generated/model"
import {
  ConcludeSetupBody,
  OpenSetupBody,
  startSetupUploadBodySizeMax,
} from "../../shared/api/generated/schemas"
import { ApiError } from "../../shared/api/request"
import { sendToTheStorage } from "../../shared/api/storage"
import { type FieldProblem, answerCodeOf } from "../../shared/errors/problems"
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
    /** The file of the logo once it is sent, or an empty text. */
    logo_file_id: string
  }
  administrator: {
    name: string
    email: string
    country: string
    national_id: string
    telephone: string
    /** The file of the photo once it is sent, or an empty text. */
    photo_file_id: string
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
      logo_file_id: "",
    },
    administrator: {
      name: "",
      email: "",
      country: "",
      national_id: "",
      telephone: "",
      photo_file_id: "",
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
    logo_file_id: optional(typed.logo_file_id),
  }
}

/**
 * The administrator as the body of concluding, or of registering a new
 * administrator, takes it, but for the preferred language, which is the
 * language of the screen when the request is sent.
 */
function administratorOf(typed: TypedSetup["administrator"]) {
  return {
    name: typed.name,
    email: typed.email,
    country: typed.country,
    national_id: typed.national_id,
    telephone: optional(typed.telephone),
    photo_file_id: optional(typed.photo_file_id),
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

/**
 * What an image of the setup is for: the logo of the consultancy, or the
 * photo of a user, which the administrator is.
 */
export type ImagePurpose = StartSetupUploadRequestPurpose

/** The largest image the setup accepts, in bytes, as the contract says. */
export const largestImage = startSetupUploadBodySizeMax

/** The types of image the setup accepts, as the contract says. */
const imageTypes: readonly string[] = Object.values(
  StartSetupUploadRequestMediaType
)

function isImageType(type: string): type is StartSetupUploadRequestMediaType {
  return imageTypes.includes(type)
}

/**
 * What is wrong with a file chosen as an image of the setup, in the codes
 * the API answers with, or nothing when it may be sent: its type and its
 * size are what the contract declares, and nothing else is checked before it
 * is sent. Concluding checks its bytes.
 */
export function refusalOf(file: File): FieldProblem | undefined {
  if (!isImageType(file.type)) {
    return { code: "file_type_unsupported", params: {} }
  }
  if (file.size === 0) {
    return { code: "file_empty", params: {} }
  }
  if (file.size > largestImage) {
    return { code: "file_too_large", params: { max_size: largestImage } }
  }
  return undefined
}

/** An image the setup sends: what it is for, and the file chosen. */
type ImageToSend = {
  purpose: ImagePurpose
  file: File
  /** The setup code, as typed. */
  code: string
}

/**
 * Sends an image of the setup: the API checks the setup code and the type and
 * size declared and answers a URL of the storage, where the browser sends the
 * bytes itself. It resolves with the identifier of the file, which concluding
 * takes; the image is confirmed then. It rejects as the API or the storage
 * answered: `setup_code_invalid` when the setup code no longer works.
 */
export async function sendSetupImage({ purpose, file, code }: ImageToSend) {
  const upload = await startSetupUpload({
    code,
    purpose,
    // A file `refusalOf` accepted; the API refuses any other type anyway.
    media_type: file.type as StartSetupUploadRequestMediaType,
    size: file.size,
  })
  await sendToTheStorage(upload.upload_url, file)
  return upload.file_id
}

/**
 * What concluding the setup, or registering a new administrator, sends,
 * once.
 */
type Conclusion = {
  typed: TypedSetup
  /** The setup code, as typed. */
  code: string
  /** The language of the screen, the preferred one of the administrator. */
  language: Language
}

/**
 * How concluding ended: the setup is concluded by this request, someone
 * concluded it before, or it is concluded and an earlier attempt of this page
 * got no answer, so which of the two it was cannot be told. Any other answer
 * rejects.
 */
type Concluded = "concluded" | "concludedBefore" | "concludedWithoutAnswer"

/**
 * The attempts to conclude the setup, which the cache of mutations keeps for
 * the life of the page, so that a later attempt knows how the earlier ones
 * ended.
 */
const concludingKey = ["setup", "conclude"]

/**
 * Whether an attempt ended without an answer of the API, so that it may have
 * concluded the setup, or registered an administrator, all the same: the
 * network failed, the answer was out of the format of the API, or Caddy
 * answered for an API it did not reach, or that did not answer in time.
 */
function endedWithoutAnswer(error: unknown) {
  if (!(error instanceof ApiError)) {
    return true
  }
  return [502, 503, 504].includes(error.problem.status)
}

/**
 * Concludes the setup: the consultancy, its first administrator and the
 * first roles are saved together, and the setup code stops working. The
 * state the gate reads is updated, so the login screen opens.
 *
 * A setup code that no longer works may have been used by someone who
 * concluded first, which ends it: the state of the setup tells. When an
 * earlier attempt of this page got no answer, that attempt may be the one
 * that concluded, and the result says so.
 */
export function useConcludeSetup() {
  const queryClient = useQueryClient()

  /** Whether an earlier attempt of this page got no answer. */
  function anEarlierAttemptGotNoAnswer() {
    return queryClient
      .getMutationCache()
      .findAll({ mutationKey: concludingKey, status: "error" })
      .some((attempt) => endedWithoutAnswer(attempt.state.error))
  }

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
    mutationKey: concludingKey,
    // Kept while the page lives, whichever step is on the screen.
    gcTime: Number.POSITIVE_INFINITY,
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
          // Perhaps by the attempt whose answer was lost: what was typed
          // may be saved, and the screen must not say it was not.
          return anEarlierAttemptGotNoAnswer()
            ? "concludedWithoutAnswer"
            : "concludedBefore"
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

/**
 * How registering a new administrator ended: registered by this request, or
 * refused in a way that an earlier attempt of this page that got no answer
 * explains, so that attempt may have registered the administrator. Any other
 * answer rejects.
 */
type Registered = "registered" | "registeredWithoutAnswer"

/** The attempts to register an administrator, kept as those of concluding. */
const registeringKey = ["setup", "register"]

/** Whether two e-mails are one, as the API compares them: in lower case. */
function sameEmail(one: string, other: string) {
  return one.trim().toLowerCase() === other.trim().toLowerCase()
}

/**
 * Registers a new administrator, the way back in: a user with the
 * administrator role, from the administrator typed, and the setup code stops
 * working. No user that exists is changed. The state the gate reads is
 * updated, so the address of the setup leads to the login screen again.
 *
 * An earlier attempt of this page that got no answer may have registered the
 * administrator, which ends the setup code and puts the e-mail in use. So a
 * later attempt refused for a setup code that stopped working, or for the
 * e-mail of such an attempt in use, ends as `registeredWithoutAnswer`, and
 * the screen never says that nothing was saved.
 *
 * The consultancy typed in the form, if any, is not sent.
 */
export function useRegisterAdministrator() {
  const queryClient = useQueryClient()

  /**
   * The e-mails of the earlier attempts of this page that got no answer:
   * each may have registered an administrator.
   */
  function emailsWithoutAnswer() {
    return (
      queryClient
        .getMutationCache()
        .findAll({ mutationKey: registeringKey, status: "error" })
        .filter((attempt) => endedWithoutAnswer(attempt.state.error))
        // The key holds the attempts of this mutation only, each sent with
        // a Conclusion.
        .map(
          (attempt) =>
            (attempt.state.variables as Conclusion).typed.administrator.email
        )
    )
  }

  /** Whether an attempt that got no answer may explain this refusal. */
  function explainedByALostAnswer(error: unknown, email: string) {
    const lost = emailsWithoutAnswer()
    switch (answerCodeOf(error)) {
      case "setup_code_invalid":
        return lost.length > 0
      case "email_in_use":
        return lost.some((sent) => sameEmail(sent, email))
      default:
        return false
    }
  }

  return useMutation({
    mutationKey: registeringKey,
    // Kept while the page lives, whichever step is on the screen.
    gcTime: Number.POSITIVE_INFINITY,
    mutationFn: async ({
      typed,
      code,
      language,
    }: Conclusion): Promise<Registered> => {
      try {
        await registerAdministrator({
          code,
          administrator: {
            ...administratorOf(typed.administrator),
            preferred_language: language,
          },
        })
      } catch (error) {
        if (!explainedByALostAnswer(error, typed.administrator.email)) {
          throw error
        }
        // Whether a setup code is still waiting is read again for the gate.
        await queryClient.fetchQuery({ ...setupStateQuery, staleTime: 0 })
        return "registeredWithoutAnswer"
      }
      // The setup code that registered the administrator stopped working.
      queryClient.setQueryData(setupStateQuery.queryKey, {
        concluded: true,
        setup_code_waiting: false,
      })
      return "registered"
    },
  })
}
