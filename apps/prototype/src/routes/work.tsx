import { createFileRoute } from "@tanstack/react-router"

import { type Place, places } from "../work/work-items"
import { WorkScreen } from "../work/work-screen"

type WorkSearch = {
  /** The place the screen shows. With none, the first one. */
  place?: Place
  /** The identifier of the item that is open. */
  item?: string
}

export const Route = createFileRoute("/work")({
  validateSearch: (search: Record<string, unknown>): WorkSearch => ({
    place: places.find((place) => place === search.place),
    item: typeof search.item === "string" ? search.item : undefined,
  }),
  component: Work,
})

/** The sample work screen: where the login screen leads. */
function Work() {
  const { place = places[0], item } = Route.useSearch()

  return <WorkScreen place={place} itemId={item} />
}
