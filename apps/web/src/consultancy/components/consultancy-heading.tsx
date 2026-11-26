import { useQuery } from "@tanstack/react-query"

import { ImageOrInitials } from "../../shared/components/image-or-initials"
import { consultancyQuery } from "../data/consultancy"

/**
 * The consultancy at the top of the sidebar of the shell: its logo, or its
 * initials, and its name, on two lines at most. A longer name ends in an
 * ellipsis, and the browser shows it whole under the pointer.
 */
export function ConsultancyHeading() {
  const { data: consultancy } = useQuery(consultancyQuery)

  if (!consultancy) {
    return null
  }

  return (
    <div className="flex min-h-10 items-center gap-2.5 px-2">
      <ImageOrInitials
        name={consultancy.name}
        imageURL={consultancy.logo_url}
        shape="square"
      />
      <span
        className="line-clamp-2 min-w-0 font-medium wrap-anywhere"
        title={consultancy.name}
      >
        {consultancy.name}
      </span>
    </div>
  )
}
