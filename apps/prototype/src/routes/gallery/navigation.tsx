import { createFileRoute } from "@tanstack/react-router"

import { SidebarBlock } from "../../gallery/blocks/sidebar"
import { BreadcrumbBlock, TabsBlock } from "../../gallery/blocks/tabs"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/navigation")({
  component: Navigation,
})

function Navigation() {
  return (
    <GalleryPage
      title="Navigation"
      lead="What takes a person from one place to another, and shows where they are."
    >
      <TabsBlock />
      <BreadcrumbBlock />
      <SidebarBlock />
    </GalleryPage>
  )
}
