import { createFileRoute } from "@tanstack/react-router"

import { SidebarBlock } from "../../gallery/blocks/sidebar"
import { BreadcrumbBlock, TabsBlock } from "../../gallery/blocks/tabs"
import { navigation } from "../../gallery/explanations/navigation"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/navigation")({
  component: Navigation,
})

function Navigation() {
  return (
    <GalleryPage section="navigation" lead={navigation.lead}>
      <TabsBlock />
      <BreadcrumbBlock />
      <SidebarBlock />
    </GalleryPage>
  )
}
