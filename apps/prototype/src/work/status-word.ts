import type { Status } from "@workspace/ui/components/status-indicator"
import { useTranslation } from "react-i18next"

import type { Place } from "./work-items"

/** The word the screen has for a status. Each place has its own four words. */
export function useStatusWord() {
  const { t } = useTranslation()
  return (place: Place, status: Status) => t(`work.status.${place}.${status}`)
}
