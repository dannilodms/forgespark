import {
  Avatar,
  AvatarFallback,
  AvatarGroup,
  AvatarGroupCount,
  AvatarImage,
} from "@workspace/ui/components/avatar"
import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Skeleton } from "@workspace/ui/components/skeleton"
import { StatusIndicator } from "@workspace/ui/components/status-indicator"
import { MoreHorizontalIcon } from "lucide-react"

import { Frame, Part, Section } from "../parts"

function Facts() {
  return (
    <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1">
      <dt className="text-muted-foreground">Open requests</dt>
      <dd className="font-mono">7</dd>
      <dt className="text-muted-foreground">Hours this month</dt>
      <dd className="font-mono">42.5</dd>
      <dt className="text-muted-foreground">Last delivery</dt>
      <dd>12 days ago</dd>
    </dl>
  )
}

export function CardBlock() {
  return (
    <Section
      title="Card"
      note="A group of facts and actions about one thing, on a page that shows several. Panes of a work screen are not cards: they are flush and split by lines."
    >
      <Part title="Typical">
        <Frame>
          <Card className="w-full max-w-sm">
            <CardHeader>
              <CardTitle>Billing portal</CardTitle>
              <CardDescription>Northwind Traders</CardDescription>
              <CardAction>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label="More actions for Billing portal"
                >
                  <MoreHorizontalIcon />
                </Button>
              </CardAction>
            </CardHeader>
            <CardContent>
              <Facts />
            </CardContent>
            <CardFooter className="justify-between gap-2">
              <StatusIndicator status="working">In progress</StatusIndicator>
              <Button variant="secondary" size="sm">
                Open project
              </Button>
            </CardFooter>
          </Card>
        </Frame>
      </Part>

      <Part title="Sizes">
        <Frame className="items-start">
          {(["sm", "default"] as const).map((size) => (
            <div key={size} className="flex flex-col items-start gap-1">
              <Card size={size} className="w-56">
                <CardHeader>
                  <CardTitle>Billing portal</CardTitle>
                  <CardDescription>Northwind Traders</CardDescription>
                </CardHeader>
                <CardContent>7 open requests</CardContent>
              </Card>
              <span className="type-code text-muted-foreground">{size}</span>
            </div>
          ))}
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long title and a long description wrap. A card that is loading keeps its shape with skeletons."
      >
        <Frame className="items-start">
          <Card className="w-full max-w-xs">
            <CardHeader>
              <CardTitle>
                Portal de cobrança e de acompanhamento das solicitações
              </CardTitle>
              <CardDescription>
                Comércio e Distribuição de Alimentos Vale do Rio Doce Ltda.
              </CardDescription>
              <CardAction>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label="Mais ações do projeto"
                >
                  <MoreHorizontalIcon />
                </Button>
              </CardAction>
            </CardHeader>
            <CardContent>
              <Facts />
            </CardContent>
          </Card>
          <Card className="w-full max-w-xs">
            <CardHeader>
              <Skeleton className="h-5 w-3/5" />
              <Skeleton className="h-4 w-2/5" />
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              <Skeleton className="h-4 w-4/5" />
              <Skeleton className="h-4 w-3/5" />
              <Skeleton className="h-4 w-2/3" />
            </CardContent>
          </Card>
        </Frame>
      </Part>
    </Section>
  )
}

/* A sample picture, drawn here so the gallery loads nothing from outside. */
const samplePicture = `data:image/svg+xml,${encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect width="64" height="64" fill="#c9d3e6"/><circle cx="32" cy="25" r="12" fill="#5b6b8c"/><path d="M8 64c2-16 12-23 24-23s22 7 24 23z" fill="#5b6b8c"/></svg>'
)}`

/* A picture that fails to load, without a request that would fail in the console. */
const brokenPicture = "data:image/png;base64,AAAA"

export function AvatarBlock() {
  return (
    <Section
      title="Avatar"
      note="The picture of a person, or the initials when there is no picture. It always sits beside the name or carries it for a screen reader."
    >
      <Part title="Typical">
        <Frame>
          <Avatar>
            <AvatarImage src={samplePicture} alt="" />
            <AvatarFallback>AL</AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <span className="font-medium">Ana Lima</span>
            <span className="type-caption text-muted-foreground">
              Northwind Traders
            </span>
          </div>
        </Frame>
      </Part>

      <Part title="Sizes">
        <Frame className="items-end">
          {(["sm", "default", "lg"] as const).map((size) => (
            <div key={size} className="flex flex-col items-start gap-1">
              <Avatar size={size}>
                <AvatarImage src={samplePicture} alt="Ana Lima" />
                <AvatarFallback>AL</AvatarFallback>
              </Avatar>
              <span className="type-code text-muted-foreground">{size}</span>
            </div>
          ))}
        </Frame>
      </Part>

      <Part
        title="A group"
        note="People on the same thing overlap. The ones that do not fit become a count."
      >
        <Frame>
          <AvatarGroup>
            <Avatar>
              <AvatarImage src={samplePicture} alt="Ana Lima" />
              <AvatarFallback>AL</AvatarFallback>
            </Avatar>
            <Avatar>
              <AvatarFallback aria-label="Bruno Tavares">BT</AvatarFallback>
            </Avatar>
            <Avatar>
              <AvatarFallback aria-label="Carla Mendes">CM</AvatarFallback>
            </Avatar>
            <AvatarGroupCount aria-label="3 more people">+3</AvatarGroupCount>
          </AvatarGroup>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="With no picture, or with one that fails to load, the initials show. A long name beside the avatar is cut with an ellipsis."
      >
        <Frame className="max-w-xs flex-col flex-nowrap items-stretch">
          <div className="flex items-center gap-3">
            <Avatar>
              <AvatarFallback>BT</AvatarFallback>
            </Avatar>
            <span>Bruno Tavares</span>
          </div>
          <div className="flex items-center gap-3">
            <Avatar>
              <AvatarImage src={brokenPicture} alt="" />
              <AvatarFallback>CM</AvatarFallback>
            </Avatar>
            <span>Carla Mendes</span>
          </div>
          <div className="flex items-center gap-3">
            <Avatar>
              <AvatarFallback>MF</AvatarFallback>
            </Avatar>
            <span className="min-w-0 truncate">
              Maria Fernanda de Albuquerque Cavalcanti e Vasconcelos
            </span>
          </div>
        </Frame>
      </Part>
    </Section>
  )
}
