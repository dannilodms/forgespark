import { Badge } from "@workspace/ui/components/badge"
import { Kbd, KbdGroup } from "@workspace/ui/components/kbd"
import { ArrowUpRightIcon, GitBranchIcon, TagIcon } from "lucide-react"
import { Trans, useTranslation } from "react-i18next"

import { useFormat } from "#stores/language"

import { actions } from "../explanations/actions"
import { partTitle } from "../explanations/shared"
import { freeText } from "../free-text"
import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const variants = [
  "default",
  "secondary",
  "outline",
  "destructive",
  "ghost",
  "link",
] as const

/** A version and a branch are identifiers: they are the same in every language. */
const version = "v2.4"
const branch = "release/2.4"

export function BadgeBlock() {
  const { t } = useTranslation()
  const { number } = useFormat()
  const text = actions.badge

  const sample = {
    default: t("gallery.badge.new"),
    secondary: t("gallery.common.tags.internal"),
    outline: version,
    destructive: t("gallery.badge.blocked"),
    ghost: t("gallery.badge.draft"),
    link: t("gallery.badge.linked", { count: 3 }),
  }

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <span className="type-title">{freeText.project}</span>
          <Badge variant="secondary">{t("gallery.common.tags.internal")}</Badge>
          <Badge variant="outline" className="font-mono">
            {version}
          </Badge>
        </Frame>
      </Part>

      <Part title={partTitle.variants}>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(15rem,1fr))] gap-3">
          {variants.map((variant) => (
            <Frame key={variant} className="flex-col items-start gap-2">
              <Badge variant={variant}>{sample[variant]}</Badge>
              <span className="type-code text-muted-foreground">{variant}</span>
              <span className="type-caption text-muted-foreground">
                {text.when[variant]}
              </span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part {...text.withIcon}>
        <Frame>
          <Badge variant="secondary">
            <TagIcon data-icon="inline-start" />
            {t("gallery.common.tags.billing")}
          </Badge>
          <Badge variant="outline">
            <GitBranchIcon data-icon="inline-start" />
            <span className="font-mono">{branch}</span>
          </Badge>
          <Badge variant="secondary">
            {t("gallery.badge.changelog")}
            <ArrowUpRightIcon data-icon="inline-end" />
          </Badge>
          <Badge variant="secondary" className="font-mono">
            {number(12)}
          </Badge>
        </Frame>
      </Part>

      <Part {...text.states}>
        <Frame>
          <StateMatrix
            rows={variants}
            states={["default", "hover", "focus"]}
            render={(variant, state) => (
              <Badge
                variant={variant}
                render={
                  <a
                    href="#badge"
                    onClick={(event) => event.preventDefault()}
                  />
                }
                {...previewState(state)}
              >
                {t("gallery.common.tags.billing")}
              </Badge>
            )}
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <Badge variant="secondary">
            <span className="truncate">{t("gallery.badge.longWaiting")}</span>
          </Badge>
          <Badge variant="outline">
            <TagIcon data-icon="inline-start" />
            <span className="truncate">{t("gallery.badge.longPending")}</span>
          </Badge>
        </Frame>
      </Part>
    </Section>
  )
}

/** The names of keys are the ones printed on a keyboard, in every language. */
export function KbdBlock() {
  const text = actions.kbd

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <span>
            <Trans i18nKey="gallery.kbd.press" components={{ kbd: <Kbd /> }} />
          </span>
        </Frame>
      </Part>

      <Part {...text.severalKeys}>
        <Frame className="gap-6">
          <KbdGroup>
            <Kbd>Ctrl</Kbd>
            <Kbd>K</Kbd>
          </KbdGroup>
          <KbdGroup>
            <Kbd>Ctrl</Kbd>
            <Kbd>Shift</Kbd>
            <Kbd>P</Kbd>
          </KbdGroup>
          <KbdGroup>
            <Kbd>Alt</Kbd>
            <Kbd>←</Kbd>
          </KbdGroup>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <KbdGroup>
            <Kbd>Ctrl</Kbd>
            <Kbd>Backspace</Kbd>
          </KbdGroup>
          <KbdGroup>
            <Kbd>Shift</Kbd>
            <Kbd>Page Down</Kbd>
          </KbdGroup>
        </Frame>
      </Part>
    </Section>
  )
}
