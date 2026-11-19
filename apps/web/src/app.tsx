import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import {
  type RouterHistory,
  RouterProvider,
  createRouter,
} from "@tanstack/react-router"

import { TooltipProvider } from "#ui/components/tooltip"

import { openTheLoginWhenTheSessionEnds } from "./access"
import type accessTexts from "./access/locales/en-US.json"
import type consultancyTexts from "./consultancy/locales/en-US.json"
import { routeTree } from "./routeTree.gen"
import type sharedTexts from "./shared/locales/en-US.json"
import { startLanguage } from "./shared/locales/i18n"

// The app as a whole: the texts of every area, the router and the client of
// the data. `main.tsx` starts it in the browser, and the screen tests start
// it at an address of their own (see src/testing).

/** Every translation file: `./<namespace>/locales/<language>.json`. */
const translationFiles = import.meta.glob<Record<string, unknown>>(
  "./*/locales/*.json",
  { eager: true, import: "default" }
)

// The keys of the English files are the keys of the app: `t` accepts no
// other. A new area adds its namespace here.
declare module "i18next" {
  interface CustomTypeOptions {
    defaultNS: "shared"
    resources: {
      shared: typeof sharedTexts
      consultancy: typeof consultancyTexts
      access: typeof accessTexts
    }
  }
}

function createAppRouter(queryClient: QueryClient, history?: RouterHistory) {
  return createRouter({ routeTree, history, context: { queryClient } })
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof createAppRouter>
  }
}

type AppOptions = {
  /** The languages the browser asks for, most wanted first. */
  languages: readonly string[]
  /** Where the router keeps the address: the browser's when left out. */
  history?: RouterHistory
}

/** Starts the texts in the language of the browser, and builds the app. */
export function createApp({ languages, history }: AppOptions) {
  startLanguage(languages, translationFiles)
  const queryClient = new QueryClient({
    // A failure is shown at once, with its trace identifier, instead of
    // being tried again behind the person's back.
    defaultOptions: { queries: { retry: false } },
  })
  const router = createAppRouter(queryClient, history)
  // A 401 past the login, from any request, opens the login: the session
  // ended while the person worked.
  openTheLoginWhenTheSessionEnds(queryClient, router)

  return function App() {
    return (
      <QueryClientProvider client={queryClient}>
        {/* Lets a second tooltip open at once after the first. */}
        <TooltipProvider>
          <RouterProvider router={router} />
        </TooltipProvider>
      </QueryClientProvider>
    )
  }
}
