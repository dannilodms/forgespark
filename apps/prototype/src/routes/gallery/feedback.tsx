import { createFileRoute } from "@tanstack/react-router"

import { AlertBlock } from "../../gallery/blocks/alert"
import { EmptyBlock, SkeletonBlock } from "../../gallery/blocks/empty"
import { StatusIndicatorBlock } from "../../gallery/blocks/status-indicator"
import { ToastBlock } from "../../gallery/blocks/toast"
import { feedback } from "../../gallery/explanations/feedback"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/feedback")({
  component: Feedback,
})

function Feedback() {
  return (
    <GalleryPage section="feedback" lead={feedback.lead}>
      <AlertBlock />
      <ToastBlock />
      <SkeletonBlock />
      <EmptyBlock />
      <StatusIndicatorBlock />
    </GalleryPage>
  )
}
