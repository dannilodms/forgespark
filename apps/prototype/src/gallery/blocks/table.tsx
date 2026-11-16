import { Button } from "@workspace/ui/components/button"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { Skeleton } from "@workspace/ui/components/skeleton"
import { StatusIndicator } from "@workspace/ui/components/status-indicator"
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import { PlusIcon } from "lucide-react"

import { Frame, Part, Section } from "../parts"

const requests = [
  {
    id: "REQ-0142",
    title: "Export leaves out archived projects",
    client: "Northwind Traders",
    status: "working",
    label: "In progress",
    hours: "6.0",
  },
  {
    id: "REQ-0139",
    title: "New logo on the invoice",
    client: "Northwind Traders",
    status: "attention",
    label: "Waiting for the client",
    hours: "1.5",
  },
  {
    id: "REQ-0131",
    title: "Login by e-mail code",
    client: "Contoso",
    status: "done",
    label: "Delivered",
    hours: "24.0",
  },
  {
    id: "REQ-0127",
    title: "Nightly import of payments",
    client: "Fabrikam",
    status: "error",
    label: "Failed",
    hours: "11.0",
  },
] as const

function Columns() {
  return (
    <TableHeader>
      <TableRow>
        <TableHead>Request</TableHead>
        <TableHead>Title</TableHead>
        <TableHead>Client</TableHead>
        <TableHead>Status</TableHead>
        <TableHead className="text-right">Hours</TableHead>
      </TableRow>
    </TableHeader>
  )
}

export function TableBlock() {
  return (
    <Section
      title="Table"
      note="Records of one kind, one per row, compared column by column. An identifier and a number are set in mono, and a number is aligned to the right."
    >
      <Part
        title="Typical"
        note="The second row is selected. A row shows the hover wash under the pointer."
      >
        <Frame className="p-0">
          <Table>
            <Columns />
            <TableBody>
              {requests.map((request, index) => (
                <TableRow
                  key={request.id}
                  data-state={index === 1 ? "selected" : undefined}
                >
                  <TableCell className="font-mono text-muted-foreground">
                    {request.id}
                  </TableCell>
                  <TableCell>{request.title}</TableCell>
                  <TableCell>{request.client}</TableCell>
                  <TableCell>
                    <StatusIndicator status={request.status}>
                      {request.label}
                    </StatusIndicator>
                  </TableCell>
                  <TableCell className="text-right font-mono">
                    {request.hours}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Frame>
      </Part>

      <Part
        title="With a total and a caption"
        note="The footer holds what sums the column. The caption says what the table is, for everyone."
      >
        <Frame className="p-0 pb-3">
          <Table>
            <TableCaption>Requests billed in March.</TableCaption>
            <Columns />
            <TableBody>
              {requests.slice(0, 2).map((request) => (
                <TableRow key={request.id}>
                  <TableCell className="font-mono text-muted-foreground">
                    {request.id}
                  </TableCell>
                  <TableCell>{request.title}</TableCell>
                  <TableCell>{request.client}</TableCell>
                  <TableCell>
                    <StatusIndicator status={request.status}>
                      {request.label}
                    </StatusIndicator>
                  </TableCell>
                  <TableCell className="text-right font-mono">
                    {request.hours}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
            <TableFooter>
              <TableRow>
                <TableCell colSpan={4}>Total</TableCell>
                <TableCell className="text-right font-mono">7.5</TableCell>
              </TableRow>
            </TableFooter>
          </Table>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long text is cut inside its cell, and the table scrolls sideways inside its frame when the columns do not fit. With no rows, the table keeps its header and shows an empty state. While it loads, each cell is a skeleton."
      >
        <Frame className="p-0">
          <Table>
            <Columns />
            <TableBody>
              <TableRow>
                <TableCell className="font-mono text-muted-foreground">
                  REQ-0118
                </TableCell>
                <TableCell className="max-w-56 truncate">
                  A exportação das solicitações não inclui os projetos
                  arquivados nem os anexos de cada solicitação
                </TableCell>
                <TableCell className="max-w-40 truncate">
                  Comércio e Distribuição de Alimentos Vale do Rio Doce Ltda.
                </TableCell>
                <TableCell>
                  <StatusIndicator status="attention" className="max-w-40">
                    Aguardando a aprovação do cliente
                  </StatusIndicator>
                </TableCell>
                <TableCell className="text-right font-mono">1,240.5</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </Frame>
        <Frame className="p-0">
          <Table>
            <Columns />
            <TableBody>
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={5} className="whitespace-normal">
                  <Empty>
                    <EmptyHeader>
                      <EmptyTitle>No request yet</EmptyTitle>
                      <EmptyDescription>
                        A request opened in this project appears here.
                      </EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                      <Button>
                        <PlusIcon data-icon="inline-start" />
                        New request
                      </Button>
                    </EmptyContent>
                  </Empty>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </Frame>
        <Frame className="p-0">
          <Table aria-busy>
            <Columns />
            <TableBody>
              {(["w-4/5", "w-3/5", "w-2/3"] as const).map((width) => (
                <TableRow key={width} className="hover:bg-transparent">
                  <TableCell>
                    <Skeleton className="h-4 w-16" />
                  </TableCell>
                  <TableCell className="w-2/5">
                    <Skeleton className={`h-4 ${width}`} />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-24" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-20" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="ml-auto h-4 w-8" />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Frame>
      </Part>
    </Section>
  )
}
