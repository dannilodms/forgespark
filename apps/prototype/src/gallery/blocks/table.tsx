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
import { useTranslation } from "react-i18next"

import { useFormat } from "#stores/language"

import { dataDisplay } from "../explanations/data-display"
import { freeText, sampleRequests } from "../free-text"
import { Frame, Part, Section } from "../parts"

type SampleRequest = (typeof sampleRequests)[number]

function Columns() {
  const { t } = useTranslation()

  return (
    <TableHeader>
      <TableRow>
        <TableHead>{t("gallery.table.request")}</TableHead>
        <TableHead>{t("gallery.table.title")}</TableHead>
        <TableHead>{t("gallery.common.client")}</TableHead>
        <TableHead>{t("gallery.table.status")}</TableHead>
        <TableHead className="text-right">
          {t("gallery.common.places.hours")}
        </TableHead>
      </TableRow>
    </TableHeader>
  )
}

/** The cells of one sample request. The hours are written for the language. */
function RequestCells({ request }: { request: SampleRequest }) {
  const { t } = useTranslation()
  const { number } = useFormat()

  return (
    <>
      <TableCell className="font-mono text-muted-foreground">
        {request.id}
      </TableCell>
      <TableCell>{request.title}</TableCell>
      <TableCell>{request.client}</TableCell>
      <TableCell>
        <StatusIndicator status={request.status}>
          {t(`gallery.common.status.${request.status}`)}
        </StatusIndicator>
      </TableCell>
      <TableCell className="text-right font-mono">
        {number(request.hours, 1)}
      </TableCell>
    </>
  )
}

export function TableBlock() {
  const { t } = useTranslation()
  const { number } = useFormat()
  const text = dataDisplay.table
  const billed = sampleRequests.slice(0, 2)

  return (
    <Section {...text.about}>
      <Part {...text.typical}>
        <Frame className="p-0">
          <Table>
            <Columns />
            <TableBody>
              {sampleRequests.map((request, index) => (
                <TableRow
                  key={request.id}
                  data-state={index === 1 ? "selected" : undefined}
                >
                  <RequestCells request={request} />
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Frame>
      </Part>

      <Part {...text.withTotal}>
        <Frame className="p-0 pb-3">
          <Table>
            <TableCaption>{t("gallery.table.caption")}</TableCaption>
            <Columns />
            <TableBody>
              {billed.map((request) => (
                <TableRow key={request.id}>
                  <RequestCells request={request} />
                </TableRow>
              ))}
            </TableBody>
            <TableFooter>
              <TableRow>
                <TableCell colSpan={4}>{t("gallery.table.total")}</TableCell>
                <TableCell className="text-right font-mono">
                  {number(
                    billed.reduce((sum, request) => sum + request.hours, 0),
                    1
                  )}
                </TableCell>
              </TableRow>
            </TableFooter>
          </Table>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="p-0">
          <Table>
            <Columns />
            <TableBody>
              <TableRow>
                <TableCell className="font-mono text-muted-foreground">
                  REQ-0118
                </TableCell>
                <TableCell className="max-w-56 truncate">
                  {freeText.longTitle}
                </TableCell>
                <TableCell className="max-w-40 truncate">
                  {freeText.longClient}
                </TableCell>
                <TableCell>
                  <StatusIndicator status="attention" className="max-w-40">
                    {t("gallery.table.longStatus")}
                  </StatusIndicator>
                </TableCell>
                <TableCell className="text-right font-mono">
                  {number(1240.5, 1)}
                </TableCell>
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
                      <EmptyTitle>{t("gallery.table.emptyTitle")}</EmptyTitle>
                      <EmptyDescription>
                        {t("gallery.table.emptyDescription")}
                      </EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                      <Button>
                        <PlusIcon data-icon="inline-start" />
                        {t("gallery.common.newRequest")}
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
