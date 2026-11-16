import { Badge } from "@workspace/ui/components/badge"
import {
  Breadcrumb,
  BreadcrumbEllipsis,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@workspace/ui/components/breadcrumb"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@workspace/ui/components/tabs"
import { ClockIcon, FileTextIcon, HistoryIcon } from "lucide-react"
import type { MouseEvent, ReactNode } from "react"

import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const variants = [
  {
    variant: "default",
    when: "Views of one thing inside a pane. The active tab is a raised control.",
  },
  {
    variant: "line",
    when: "Views that span the width of a pane, under its header.",
  },
] as const

const matrixRows = [
  "default",
  "default, active",
  "line",
  "line, active",
] as const

export function TabsBlock() {
  return (
    <Section
      title="Tabs"
      note="Views of the same thing, one at a time. The arrow keys move between the tabs, and Tab goes into the view."
    >
      <Part title="Typical">
        <Frame>
          <Tabs defaultValue="details" className="w-full">
            <TabsList>
              <TabsTrigger value="details">Details</TabsTrigger>
              <TabsTrigger value="history">History</TabsTrigger>
              <TabsTrigger value="hours">Hours</TabsTrigger>
            </TabsList>
            <TabsContent value="details">
              The client asked for the export to include archived projects.
            </TabsContent>
            <TabsContent value="history">
              Opened by Ana Lima 12 days ago, and assigned to Bruno Tavares.
            </TabsContent>
            <TabsContent value="hours">
              6 hours logged, of the 12 that were estimated.
            </TabsContent>
          </Tabs>
        </Frame>
      </Part>

      <Part title="Variants">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-3">
          {variants.map(({ variant, when }) => (
            <Frame key={variant} className="flex-col items-start gap-2">
              <Tabs defaultValue="details">
                <TabsList variant={variant}>
                  <TabsTrigger value="details">Details</TabsTrigger>
                  <TabsTrigger value="history">History</TabsTrigger>
                  <TabsTrigger value="hours">Hours</TabsTrigger>
                </TabsList>
              </Tabs>
              <span className="type-code text-muted-foreground">{variant}</span>
              <span className="type-caption text-muted-foreground">{when}</span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part
        title="With an icon or a count, and vertical"
        note="An icon leads the label and a count trails it, in mono. A vertical list sits beside its view."
      >
        <Frame className="items-start gap-8">
          <Tabs defaultValue="details">
            <TabsList>
              <TabsTrigger value="details">
                <FileTextIcon data-icon="inline-start" />
                Details
              </TabsTrigger>
              <TabsTrigger value="history">
                <HistoryIcon data-icon="inline-start" />
                History
                <Badge variant="secondary" className="font-mono">
                  8
                </Badge>
              </TabsTrigger>
              <TabsTrigger value="hours">
                <ClockIcon data-icon="inline-start" />
                Hours
              </TabsTrigger>
            </TabsList>
          </Tabs>
          <Tabs defaultValue="details" orientation="vertical">
            <TabsList variant="line">
              <TabsTrigger value="details">Details</TabsTrigger>
              <TabsTrigger value="history">History</TabsTrigger>
              <TabsTrigger value="hours">Hours</TabsTrigger>
            </TabsList>
            <TabsContent value="details">Details of the request.</TabsContent>
            <TabsContent value="history">History of the request.</TabsContent>
            <TabsContent value="hours">Hours of the request.</TabsContent>
          </Tabs>
        </Frame>
      </Part>

      <Part
        title="States"
        note="Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={matrixRows}
            states={["default", "hover", "focus", "disabled"]}
            render={(row, state) => (
              <Tabs defaultValue={row.endsWith("active") ? "details" : null}>
                <TabsList variant={row.startsWith("line") ? "line" : "default"}>
                  <TabsTrigger value="details" {...previewState(state)}>
                    Details
                  </TabsTrigger>
                </TabsList>
              </Tabs>
            )}
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="Long labels wrap onto a second line of tabs before they push the layout, and a single label can be cut with an ellipsis."
      >
        <Frame className="max-w-sm">
          <Tabs defaultValue="details" className="w-full">
            <TabsList className="flex-wrap">
              <TabsTrigger value="details">Dados da solicitação</TabsTrigger>
              <TabsTrigger value="history">Histórico de alterações</TabsTrigger>
              <TabsTrigger value="hours" className="max-w-full">
                <span className="truncate">
                  Horas lançadas pela equipe neste mês e no mês anterior
                </span>
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </Frame>
      </Part>
    </Section>
  )
}

/** A crumb of the gallery leads nowhere: it only shows how a link looks. */
function stay(event: MouseEvent) {
  event.preventDefault()
}

function Crumb({ children }: { children: ReactNode }) {
  return (
    <BreadcrumbLink href="#breadcrumb" onClick={stay}>
      {children}
    </BreadcrumbLink>
  )
}

export function BreadcrumbBlock() {
  return (
    <Section
      title="Breadcrumb"
      note="Where the current page sits, as a path of links back up. The last item is the page itself and is not a link."
    >
      <Part title="Typical">
        <Frame>
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Crumb>Clients</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <Crumb>Northwind Traders</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <Crumb>Billing portal</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage className="font-mono">REQ-0142</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </Frame>
      </Part>

      <Part
        title="A long path, and another separator"
        note="The middle of a long path collapses into an ellipsis. A slash can stand in for the chevron."
      >
        <Frame className="flex-col items-start">
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Crumb>Clients</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbEllipsis />
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <Crumb>Billing portal</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage className="font-mono">REQ-0142</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Crumb>Clients</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator>/</BreadcrumbSeparator>
              <BreadcrumbItem>
                <Crumb>Northwind Traders</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator>/</BreadcrumbSeparator>
              <BreadcrumbItem>
                <BreadcrumbPage>Billing portal</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long name is cut with an ellipsis, and a path wider than its container wraps onto a second line."
      >
        <Frame className="max-w-sm">
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Crumb>Clientes</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem className="max-w-44">
                <BreadcrumbLink
                  href="#breadcrumb"
                  onClick={stay}
                  className="truncate"
                >
                  Comércio e Distribuição de Alimentos Vale do Rio Doce Ltda.
                </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem className="max-w-44">
                <BreadcrumbPage className="truncate">
                  Portal de cobrança e de acompanhamento das solicitações
                </BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </Frame>
      </Part>
    </Section>
  )
}
