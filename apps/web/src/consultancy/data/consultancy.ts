import { queryOptions } from "@tanstack/react-query"

import { tellConsultancy } from "../../shared/api/generated/client"
import { signedURLsRenewEvery } from "../../shared/api/storage"

// The read of the consultancy as the shell of the app shows it (Data module
// in CODING_STANDARDS.md): there is one consultancy, so it is no set, and it
// goes through TanStack Query.

/**
 * The name of the consultancy and the URL of its logo, which any logged-in
 * user reads. It is read again every ten minutes while the tab shows, and
 * when the person comes back to the tab a minute or more after the last
 * read, so the URL of the logo, which works for fifteen minutes, is always
 * one that works.
 */
export const consultancyQuery = queryOptions({
  queryKey: ["consultancy"],
  queryFn: () => tellConsultancy(),
  staleTime: 60_000,
  refetchInterval: signedURLsRenewEvery,
})
