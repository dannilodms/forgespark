import { Button } from "@workspace/ui/components/button"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { Skeleton } from "@workspace/ui/components/skeleton"
import { InboxIcon, PlusIcon, SearchXIcon } from "lucide-react"
import { Trans, useTranslation } from "react-i18next"

import { feedback } from "../explanations/feedback"
import { partTitle } from "../explanations/shared"
import { freeText } from "../free-text"
import { Frame, Part, Section } from "../parts"

export function SkeletonBlock() {
  const text = feedback.skeleton

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm">
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <div className="flex min-w-0 flex-1 flex-col gap-2">
            <Skeleton className="h-4 w-2/5" />
            <Skeleton className="h-3 w-4/5" />
          </div>
        </Frame>
      </Part>

      <Part {...text.list}>
        <Frame className="max-w-md flex-col flex-nowrap items-stretch gap-0 p-0">
          {(["w-3/5", "w-4/5", "w-2/5", "w-3/4"] as const).map((width) => (
            <div
              key={width}
              className="flex h-7 items-center gap-3 border-b border-border-subtle px-3 last:border-b-0"
            >
              <Skeleton className="h-3 w-14 shrink-0" />
              <Skeleton className={`h-3 ${width}`} />
            </div>
          ))}
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-40">
          <Skeleton className="h-4 w-80 max-w-full" />
          <Skeleton className="h-3 w-64 max-w-full" />
        </Frame>
      </Part>
    </Section>
  )
}

export function EmptyBlock() {
  const { t } = useTranslation()
  const text = feedback.empty

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <InboxIcon />
              </EmptyMedia>
              <EmptyTitle>{t("gallery.empty.nothingWaiting.title")}</EmptyTitle>
              <EmptyDescription>
                {t("gallery.empty.nothingWaiting.description")}
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button>
                <PlusIcon data-icon="inline-start" />
                {t("gallery.common.newRequest")}
              </Button>
            </EmptyContent>
          </Empty>
        </Frame>
      </Part>

      <Part {...text.afterSearch}>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-3">
          <Frame>
            <Empty className="border">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <SearchXIcon />
                </EmptyMedia>
                <EmptyTitle>{t("gallery.empty.noMatch.title")}</EmptyTitle>
                <EmptyDescription>
                  <Trans
                    i18nKey="gallery.empty.noMatch.description"
                    values={{ search: freeText.search }}
                    components={{ term: <span className="font-mono" /> }}
                  />
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button variant="outline">
                  {t("gallery.empty.clearSearch")}
                </Button>
              </EmptyContent>
            </Empty>
          </Frame>
          <Frame>
            <Empty className="border">
              <EmptyHeader>
                <EmptyTitle>{t("gallery.empty.noProject.title")}</EmptyTitle>
                <EmptyDescription>
                  {t("gallery.empty.noProject.description")}
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent className="flex-row justify-center">
                <Button>{t("gallery.common.newProject")}</Button>
                <Button variant="ghost">{t("gallery.empty.import")}</Button>
              </EmptyContent>
            </Empty>
          </Frame>
        </div>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <Empty>
            <EmptyHeader>
              <EmptyTitle>{t("gallery.empty.long.title")}</EmptyTitle>
              <EmptyDescription>
                {t("gallery.empty.long.description")}
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button variant="outline" className="max-w-full">
                <span className="truncate">
                  {t("gallery.empty.long.action")}
                </span>
              </Button>
            </EmptyContent>
          </Empty>
        </Frame>
      </Part>
    </Section>
  )
}
