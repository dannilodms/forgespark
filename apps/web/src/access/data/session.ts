import {
  type QueryClient,
  queryOptions,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"
import type { RegisteredRouter } from "@tanstack/react-router"

import {
  askForLoginCode,
  checkLoginCode,
  getStartProviderLoginUrl,
  logOut,
  tellLoggedInUser,
  tellLoginState,
} from "../../shared/api/generated/client"
import type {
  LoggedInUser,
  TurnedOnLoginProvider,
} from "../../shared/api/generated/model"
import {
  AskForLoginCodeBody,
  CheckLoginCodeBody,
  StartProviderLoginQueryParams,
} from "../../shared/api/generated/schemas"
import { ApiError, whenUnauthorized } from "../../shared/api/request"

// The reads and writes of the session (Data module in CODING_STANDARDS.md):
// who is logged in, the login state, asking for a login code, checking it,
// the way to a login through a provider, and logging out. None is a set:
// they go through TanStack Query.

/**
 * Who is logged in to the session of this browser, or `null` when nobody is:
 * the API answers 401 then, which here is a value and not a failure, because
 * the gate decides with it. It is read again when the person comes back to
 * the tab a minute or more after the last read.
 */
export const loggedInUserQuery = queryOptions({
  queryKey: ["session"],
  queryFn: async (): Promise<LoggedInUser | null> => {
    try {
      return await tellLoggedInUser()
    } catch (error) {
      if (error instanceof ApiError && error.problem.status === 401) {
        return null
      }
      throw error
    }
  },
  staleTime: 60_000,
})

/**
 * Forgets everything the app read for the session that ended, so that
 * nothing of it shows to whoever logs in next, and keeps that nobody is
 * logged in, so the gate opens the login without asking the API again.
 */
function forgetTheSession(queryClient: QueryClient) {
  queryClient.removeQueries()
  queryClient.setQueryData(loggedInUserQuery.queryKey, null)
}

/**
 * Whether a path is a path of the app, which the login may open afterwards:
 * it starts with one slash, so it cannot lead to another origin
 * (`//elsewhere.example`, or a backslash that browsers read as a slash).
 */
function isPathOfTheApp(path: string) {
  return /^\/(?![/\\])/.test(path)
}

/**
 * Where the app goes once a user is logged in: the path to return to, when
 * it is a path of the app, and the root otherwise. It takes whatever the
 * address held.
 */
export function whereTheLoginLeads(returnTo: unknown) {
  return typeof returnTo === "string" && isPathOfTheApp(returnTo)
    ? returnTo
    : "/"
}

/**
 * The address to return to after the login, from the address the person is
 * at: its path, search and hash. The root, the login and the setup lead
 * nowhere the login would not.
 */
export function returnPathOf(location: { href: string; pathname: string }) {
  const leadsNowhere =
    location.href === "/" ||
    location.pathname === "/login" ||
    location.pathname === "/setup"
  return leadsNowhere ? undefined : location.href
}

/**
 * Opens the login when the API answers 401 while the person works past it:
 * the session ended, idle for 7 days, 30 days old, or of a user deactivated
 * since. The app forgets what it read, goes back to the language of the
 * selector (the gate does it), and the login says the session ended and
 * keeps where the person was, to return there afterwards. A 401 before any
 * screen opened, at the login or at the setup is no session that ended: the
 * gate reads who is logged in there, and nobody is.
 *
 * Set once, when the app is put together (`createApp`): `request`, the fetch
 * every generated function calls, tells of every 401.
 */
export function openTheLoginWhenTheSessionEnds(
  queryClient: QueryClient,
  router: RegisteredRouter
) {
  whenUnauthorized(() => {
    const at = router.state.resolvedLocation
    if (!at || at.pathname === "/login" || at.pathname === "/setup") {
      return
    }
    forgetTheSession(queryClient)
    void router.navigate({
      to: "/login",
      search: { notice: "session_ended", return_to: returnPathOf(at) },
      replace: true,
    })
  })
}

/**
 * The login state, which the login screen reads when it opens, before anyone
 * is logged in: the login providers turned on, one button each, in the
 * order the API gives them, and the e-mail the last login through a
 * provider in this browser was refused for, which the notice
 * `no_active_user` shows. The e-mail travels in this answer and in no
 * address. Reading it forgets nothing.
 */
export const loginStateQuery = queryOptions({
  queryKey: ["session", "login"],
  queryFn: () => tellLoginState(),
})

/** A login provider turned on, as the login state names it. */
export type LoginProvider = TurnedOnLoginProvider

/**
 * The address of the button of a login provider: the start of a login
 * through it, to which the browser navigates (the generated client never
 * calls it, since the API answers it with a redirect to the provider). It
 * carries the path to return to once logged in, when there is one other
 * than the root that the contract takes; the provider brings the browser
 * back there, or to the login with a notice that keeps it.
 */
export function providerLoginAddress(key: string, returnTo: unknown) {
  const path = whereTheLoginLeads(returnTo)
  const taken = StartProviderLoginQueryParams.safeParse({ return_to: path })
  return getStartProviderLoginUrl(
    key,
    path !== "/" && taken.success ? taken.data : {}
  )
}

/** What the screen sends to ask for a login code, as the contract checks it. */
export const emailContract = AskForLoginCodeBody

/** What the screen sends to check a login code, as the contract checks it. */
export const loginCodeContract = CheckLoginCodeBody

/**
 * Asks for a login code for an e-mail. The API answers the same for every
 * e-mail, and sends the code only when it belongs to an active user. It
 * rejects with `login_code_limit_reached` past the limit of codes for the
 * address, and with the 429 of Caddy past the limit of the address of the
 * browser.
 */
export function useAskForLoginCode() {
  return useMutation({
    mutationFn: (email: string) => askForLoginCode({ email }),
  })
}

/**
 * Checks the six digits of a login code, in the browser that asked for it:
 * a right code logs the user in, and who is logged in is read again, so the
 * gate opens the app. It rejects with `login_code_invalid`, with the
 * `attempts_left`, or with `login_code_ended` at the third wrong attempt.
 */
export function useCheckLoginCode() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (code: string) => {
      await checkLoginCode({ code })
      return queryClient.fetchQuery({ ...loggedInUserQuery, staleTime: 0 })
    },
  })
}

/**
 * Logs out: the API ends the session of this browser, whatever it held, and
 * the app forgets what it read for it.
 */
export function useLogOut() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      await logOut()
      forgetTheSession(queryClient)
    },
  })
}
