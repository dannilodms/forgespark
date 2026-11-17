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
import { useTranslation } from "react-i18next"

import { useFormat } from "#stores/language"

import { dataDisplay } from "../explanations/data-display"
import { partTitle } from "../explanations/shared"
import { freeText } from "../free-text"
import { Frame, Part, Section } from "../parts"

function Facts() {
  const { t } = useTranslation()
  const { number, ago } = useFormat()

  return (
    <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1">
      <dt className="text-muted-foreground">
        {t("gallery.common.openRequests")}
      </dt>
      <dd className="font-mono">{number(7)}</dd>
      <dt className="text-muted-foreground">
        {t("gallery.common.hoursThisMonth")}
      </dt>
      <dd className="font-mono">{number(42.5, 1)}</dd>
      <dt className="text-muted-foreground">
        {t("gallery.card.lastDelivery")}
      </dt>
      <dd>{ago(12, "day")}</dd>
    </dl>
  )
}

export function CardBlock() {
  const { t } = useTranslation()
  const text = dataDisplay.card

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Card className="w-full max-w-sm">
            <CardHeader>
              <CardTitle>{freeText.project}</CardTitle>
              <CardDescription>{freeText.client}</CardDescription>
              <CardAction>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={t("gallery.common.moreActions", {
                    name: freeText.project,
                  })}
                >
                  <MoreHorizontalIcon />
                </Button>
              </CardAction>
            </CardHeader>
            <CardContent>
              <Facts />
            </CardContent>
            <CardFooter className="justify-between gap-2">
              <StatusIndicator status="working">
                {t("gallery.common.status.working")}
              </StatusIndicator>
              <Button variant="secondary" size="sm">
                {t("gallery.card.openProject")}
              </Button>
            </CardFooter>
          </Card>
        </Frame>
      </Part>

      <Part title={partTitle.sizes}>
        <Frame className="items-start">
          {(["sm", "default"] as const).map((size) => (
            <div key={size} className="flex flex-col items-start gap-1">
              <Card size={size} className="w-56">
                <CardHeader>
                  <CardTitle>{freeText.project}</CardTitle>
                  <CardDescription>{freeText.client}</CardDescription>
                </CardHeader>
                <CardContent>
                  {t("gallery.common.openRequestsCount", { count: 7 })}
                </CardContent>
              </Card>
              <span className="type-code text-muted-foreground">{size}</span>
            </div>
          ))}
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="items-start">
          <Card className="w-full max-w-xs">
            <CardHeader>
              <CardTitle>{freeText.longProject}</CardTitle>
              <CardDescription>{freeText.longClient}</CardDescription>
              <CardAction>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={t("gallery.common.moreActions", {
                    name: freeText.longProject,
                  })}
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
  const { t } = useTranslation()
  const text = dataDisplay.avatar

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Avatar>
            <AvatarImage src={samplePicture} alt="" />
            <AvatarFallback>{freeText.anaInitials}</AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <span className="font-medium">{freeText.ana}</span>
            <span className="type-caption text-muted-foreground">
              {freeText.client}
            </span>
          </div>
        </Frame>
      </Part>

      <Part title={partTitle.sizes}>
        <Frame className="items-end">
          {(["sm", "default", "lg"] as const).map((size) => (
            <div key={size} className="flex flex-col items-start gap-1">
              <Avatar size={size}>
                <AvatarImage src={samplePicture} alt={freeText.ana} />
                <AvatarFallback>{freeText.anaInitials}</AvatarFallback>
              </Avatar>
              <span className="type-code text-muted-foreground">{size}</span>
            </div>
          ))}
        </Frame>
      </Part>

      <Part {...text.group}>
        <Frame>
          <AvatarGroup>
            <Avatar>
              <AvatarImage src={samplePicture} alt={freeText.ana} />
              <AvatarFallback>{freeText.anaInitials}</AvatarFallback>
            </Avatar>
            <Avatar>
              <AvatarFallback aria-label={freeText.bruno}>
                {freeText.brunoInitials}
              </AvatarFallback>
            </Avatar>
            <Avatar>
              <AvatarFallback aria-label={freeText.carla}>
                {freeText.carlaInitials}
              </AvatarFallback>
            </Avatar>
            <AvatarGroupCount
              aria-label={t("gallery.avatar.morePeople", { count: 3 })}
            >
              +3
            </AvatarGroupCount>
          </AvatarGroup>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs flex-col flex-nowrap items-stretch">
          <div className="flex items-center gap-3">
            <Avatar>
              <AvatarFallback>{freeText.brunoInitials}</AvatarFallback>
            </Avatar>
            <span>{freeText.bruno}</span>
          </div>
          <div className="flex items-center gap-3">
            <Avatar>
              <AvatarImage src={brokenPicture} alt="" />
              <AvatarFallback>{freeText.carlaInitials}</AvatarFallback>
            </Avatar>
            <span>{freeText.carla}</span>
          </div>
          <div className="flex items-center gap-3">
            <Avatar>
              <AvatarFallback>{freeText.longNameInitials}</AvatarFallback>
            </Avatar>
            <span className="min-w-0 truncate">{freeText.longName}</span>
          </div>
        </Frame>
      </Part>
    </Section>
  )
}
