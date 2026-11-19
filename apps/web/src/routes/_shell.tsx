import { Outlet, createFileRoute } from "@tanstack/react-router"

import { UserMenu } from "../access"
import { ConsultancyHeading, consultancyQuery } from "../consultancy"
import { AppShell } from "../shared/components/app-shell"

// Every screen after the login, in the shell of the app. The gate in
// `__root.tsx` lets only a logged-in user reach it; the shell reads the
// consultancy before it opens. The pieces come from two areas, so this route
// puts them together: the menu of the user from access, the consultancy from
// consultancy.
export const Route = createFileRoute("/_shell")({
  loader: ({ context: { queryClient } }) =>
    queryClient.ensureQueryData(consultancyQuery),
  component: Shell,
})

function Shell() {
  return (
    <AppShell consultancy={<ConsultancyHeading />} userMenu={<UserMenu />}>
      <Outlet />
    </AppShell>
  )
}
