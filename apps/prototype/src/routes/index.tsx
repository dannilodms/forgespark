import { Link, createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/")({ component: Home })

const destinations = [
  {
    to: "/login",
    label: "Login screen",
    about: "The first screen of the product: login providers or a login code.",
  },
  {
    to: "/work",
    label: "Sample work screen",
    about:
      "A full work surface over sample data, built from the four recipes: where the login leads.",
  },
  {
    to: "/gallery",
    label: "Gallery",
    about: "Every style variable and component of the design system.",
  },
] as const

function Home() {
  return (
    <main className="mx-auto flex max-w-xl flex-col gap-6 p-6">
      <header className="flex flex-col gap-2">
        <h1 className="type-display">Prototype</h1>
        <p className="text-muted-foreground">
          The clickable prototype of My Software Factory Manager. Screens are
          designed and approved here, with fake data, before the product builds
          them.
        </p>
      </header>
      <nav aria-label="Prototype" className="flex flex-col border-t">
        {destinations.map((destination) => (
          <Link
            key={destination.to}
            to={destination.to}
            className="flex flex-col gap-0.5 border-b px-2 py-3 hover:bg-accent"
          >
            <span className="type-title">{destination.label}</span>
            <span className="type-caption text-muted-foreground">
              {destination.about}
            </span>
          </Link>
        ))}
      </nav>
    </main>
  )
}
