import { StatusIndicator } from "@workspace/ui/components/status-indicator"
import { cn } from "@workspace/ui/lib/utils"
import { useTranslation } from "react-i18next"

import { feedback } from "../explanations/feedback"
import { partTitle } from "../explanations/shared"
import { sampleRequest, sampleRequests } from "../free-text"
import { Frame, Part, Section } from "../parts"

const statuses = ["working", "attention", "done", "error"] as const

export function StatusIndicatorBlock() {
  const { t } = useTranslation()
  const text = feedback.statusIndicator

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-md justify-between">
          <span className="min-w-0 truncate">{sampleRequest.title}</span>
          <StatusIndicator status={sampleRequest.status}>
            {t(`gallery.common.status.${sampleRequest.status}`)}
          </StatusIndicator>
        </Frame>
      </Part>

      <Part {...text.meanings}>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(13rem,1fr))] gap-3">
          {statuses.map((status) => (
            <Frame key={status} className="flex-col items-start gap-2">
              <StatusIndicator status={status}>
                {t(`gallery.common.status.${status}`)}
              </StatusIndicator>
              <span className="type-code text-muted-foreground">{status}</span>
              <span className="type-caption text-muted-foreground">
                {text.when[status]}
              </span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part {...text.variants}>
        <Frame className="flex-col items-start">
          <div className="flex flex-wrap items-center gap-4">
            {statuses.map((status) => (
              <StatusIndicator key={status} status={status}>
                {t(`gallery.common.status.${status}`)}
              </StatusIndicator>
            ))}
          </div>
          <div className="flex flex-wrap items-center gap-3">
            {statuses.map((status) => (
              <StatusIndicator key={status} status={status} variant="badge">
                {t(`gallery.common.status.${status}`)}
              </StatusIndicator>
            ))}
          </div>
        </Frame>
      </Part>

      <Part {...text.inList}>
        <Frame className="max-w-xl flex-col flex-nowrap items-stretch gap-0 p-0">
          {sampleRequests.map(({ id, title, status }) => (
            <div
              key={id}
              className="flex min-h-7 items-center gap-3 border-b border-border-subtle px-3 py-1 last:border-b-0"
            >
              <span className="shrink-0 type-code text-muted-foreground">
                {id}
              </span>
              <span className="min-w-0 flex-1 truncate">{title}</span>
              <StatusIndicator status={status} className="shrink">
                {t(`gallery.common.status.${status}`)}
              </StatusIndicator>
            </div>
          ))}
        </Frame>
      </Part>

      <Part {...text.onRow}>
        <Frame className="max-w-xl flex-col flex-nowrap items-stretch gap-0 p-0">
          {(["bg-accent", "bg-selected"] as const).map((fill) =>
            sampleRequests.map(({ id, title, status }) => (
              <div
                key={`${fill}-${id}`}
                className={cn(
                  "flex min-h-7 items-center gap-3 border-b border-border-subtle px-3 py-1 last:border-b-0",
                  fill
                )}
              >
                <span className="shrink-0 type-code text-muted-foreground">
                  {id}
                </span>
                <span className="min-w-0 flex-1 truncate">{title}</span>
                <StatusIndicator status={status} className="shrink">
                  {t(`gallery.common.status.${status}`)}
                </StatusIndicator>
              </div>
            ))
          )}
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs flex-col items-start">
          <StatusIndicator status="attention" className="shrink">
            {t("gallery.statusIndicator.longAttention")}
          </StatusIndicator>
          <StatusIndicator status="error" variant="badge" className="shrink">
            {t("gallery.statusIndicator.longError")}
          </StatusIndicator>
        </Frame>
      </Part>
    </Section>
  )
}
