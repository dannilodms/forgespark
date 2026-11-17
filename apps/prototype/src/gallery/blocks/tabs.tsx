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
import { useTranslation } from "react-i18next"

import { useFormat } from "#stores/language"

import { navigation } from "../explanations/navigation"
import { partTitle, pinnedStates } from "../explanations/shared"
import { freeText, sampleRequest } from "../free-text"
import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const variants = ["default", "line"] as const

const matrixRows = [
  "default",
  "default, active",
  "line",
  "line, active",
] as const

export function TabsBlock() {
  const { t } = useTranslation()
  const { number, ago } = useFormat()
  const text = navigation.tabs

  const details = t("gallery.common.tabs.details")
  const history = t("gallery.common.tabs.history")
  const hours = t("gallery.common.tabs.hours")

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Tabs defaultValue="details" className="w-full">
            <TabsList>
              <TabsTrigger value="details">{details}</TabsTrigger>
              <TabsTrigger value="history">{history}</TabsTrigger>
              <TabsTrigger value="hours">{hours}</TabsTrigger>
            </TabsList>
            <TabsContent value="details">{freeText.description}</TabsContent>
            <TabsContent value="history">
              {t("gallery.tabs.historyContent", {
                who: freeText.ana,
                when: ago(12, "day"),
                assignee: freeText.bruno,
              })}
            </TabsContent>
            <TabsContent value="hours">
              {t("gallery.tabs.hoursContent", {
                logged: number(6),
                estimated: number(12),
              })}
            </TabsContent>
          </Tabs>
        </Frame>
      </Part>

      <Part title={partTitle.variants}>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-3">
          {variants.map((variant) => (
            <Frame key={variant} className="flex-col items-start gap-2">
              <Tabs defaultValue="details">
                <TabsList variant={variant}>
                  <TabsTrigger value="details">{details}</TabsTrigger>
                  <TabsTrigger value="history">{history}</TabsTrigger>
                  <TabsTrigger value="hours">{hours}</TabsTrigger>
                </TabsList>
              </Tabs>
              <span className="type-code text-muted-foreground">{variant}</span>
              <span className="type-caption text-muted-foreground">
                {text.when[variant]}
              </span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part {...text.withIcon}>
        <Frame className="items-start gap-8">
          <Tabs defaultValue="details">
            <TabsList>
              <TabsTrigger value="details">
                <FileTextIcon data-icon="inline-start" />
                {details}
              </TabsTrigger>
              <TabsTrigger value="history">
                <HistoryIcon data-icon="inline-start" />
                {history}
                <Badge variant="secondary" className="font-mono">
                  {number(8)}
                </Badge>
              </TabsTrigger>
              <TabsTrigger value="hours">
                <ClockIcon data-icon="inline-start" />
                {hours}
              </TabsTrigger>
            </TabsList>
          </Tabs>
          <Tabs defaultValue="details" orientation="vertical">
            <TabsList variant="line">
              <TabsTrigger value="details">{details}</TabsTrigger>
              <TabsTrigger value="history">{history}</TabsTrigger>
              <TabsTrigger value="hours">{hours}</TabsTrigger>
            </TabsList>
            <TabsContent value="details">
              {t("gallery.tabs.detailsOf")}
            </TabsContent>
            <TabsContent value="history">
              {t("gallery.tabs.historyOf")}
            </TabsContent>
            <TabsContent value="hours">{t("gallery.tabs.hoursOf")}</TabsContent>
          </Tabs>
        </Frame>
      </Part>

      <Part {...pinnedStates}>
        <Frame>
          <StateMatrix
            rows={matrixRows}
            states={["default", "hover", "focus", "disabled"]}
            render={(row, state) => (
              <Tabs defaultValue={row.endsWith("active") ? "details" : null}>
                <TabsList variant={row.startsWith("line") ? "line" : "default"}>
                  <TabsTrigger value="details" {...previewState(state)}>
                    {details}
                  </TabsTrigger>
                </TabsList>
              </Tabs>
            )}
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-sm">
          <Tabs defaultValue="details" className="w-full">
            <TabsList className="flex-wrap">
              <TabsTrigger value="details">
                {t("gallery.tabs.long.details")}
              </TabsTrigger>
              <TabsTrigger value="history">
                {t("gallery.tabs.long.history")}
              </TabsTrigger>
              <TabsTrigger value="hours" className="max-w-full">
                <span className="truncate">{t("gallery.tabs.long.hours")}</span>
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
  const { t } = useTranslation()
  const text = navigation.breadcrumb
  const label = t("gallery.common.path")
  const clients = t("gallery.common.places.clients")

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Breadcrumb aria-label={label}>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Crumb>{clients}</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <Crumb>{freeText.client}</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <Crumb>{freeText.project}</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage className="font-mono">
                  {sampleRequest.id}
                </BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </Frame>
      </Part>

      <Part {...text.longPath}>
        <Frame className="flex-col items-start">
          <Breadcrumb aria-label={label}>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Crumb>{clients}</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbEllipsis />
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <Crumb>{freeText.project}</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage className="font-mono">
                  {sampleRequest.id}
                </BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <Breadcrumb aria-label={label}>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Crumb>{clients}</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator>/</BreadcrumbSeparator>
              <BreadcrumbItem>
                <Crumb>{freeText.client}</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator>/</BreadcrumbSeparator>
              <BreadcrumbItem>
                <BreadcrumbPage>{freeText.project}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-sm">
          <Breadcrumb aria-label={label}>
            <BreadcrumbList>
              <BreadcrumbItem>
                <Crumb>{clients}</Crumb>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem className="max-w-44">
                <BreadcrumbLink
                  href="#breadcrumb"
                  onClick={stay}
                  className="truncate"
                >
                  {freeText.longClient}
                </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem className="max-w-44">
                <BreadcrumbPage className="truncate">
                  {freeText.longProject}
                </BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </Frame>
      </Part>
    </Section>
  )
}
