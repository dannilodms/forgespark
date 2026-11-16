import { createFileRoute } from "@tanstack/react-router"
import { Button } from "@workspace/ui/components/button"

export const Route = createFileRoute("/")({ component: Home })

function Home() {
  return (
    <main className="mx-auto flex max-w-xl flex-col gap-4 p-6">
      <h1 className="text-xl font-medium">Prototype</h1>
      <p className="text-muted-foreground">
        The clickable prototype of My Software Factory Manager. Screens are
        designed and approved here, with fake data, before the product builds
        them.
      </p>
      <div>
        <Button>Button</Button>
      </div>
    </main>
  )
}
