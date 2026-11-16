import { createFileRoute } from "@tanstack/react-router"

import { AlertBlock } from "../../gallery/blocks/alert"
import { EmptyBlock, SkeletonBlock } from "../../gallery/blocks/empty"
import { StatusIndicatorBlock } from "../../gallery/blocks/status-indicator"
import { ToastBlock } from "../../gallery/blocks/toast"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/feedback")({
  component: Feedback,
})

function Feedback() {
  return (
    <GalleryPage
      title="Feedback"
      lead="What tells a person the state of things: a message, a status, something loading, nothing to show."
    >
      <AlertBlock />
      <ToastBlock />
      <SkeletonBlock />
      <EmptyBlock />
      <StatusIndicatorBlock />
    </GalleryPage>
  )
}
