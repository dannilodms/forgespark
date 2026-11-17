import { Button, ButtonShortcut } from "@workspace/ui/components/button"
import {
  ArrowRightIcon,
  LoaderCircleIcon,
  PlusIcon,
  SettingsIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { actions } from "../explanations/actions"
import { partTitle, pinnedStates } from "../explanations/shared"
import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

const text = actions.button

/** Each variant, with the key of its sample label in the translation files. */
const variants = [
  { variant: "default", label: "gallery.button.createRequest" },
  { variant: "secondary", label: "gallery.common.saveDraft" },
  { variant: "outline", label: "gallery.button.export" },
  { variant: "ghost", label: "gallery.common.cancel" },
  { variant: "destructive", label: "gallery.common.deleteProject" },
  { variant: "link", label: "gallery.button.seeHistory" },
] as const

const sizes = ["xs", "sm", "default", "lg"] as const

export function ButtonBlock() {
  const { t } = useTranslation()

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Button>{t("gallery.common.saveChanges")}</Button>
          <Button variant="ghost">{t("gallery.common.cancel")}</Button>
        </Frame>
      </Part>

      <Part title={partTitle.variants}>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(15rem,1fr))] gap-3">
          {variants.map(({ variant, label }) => (
            <Frame key={variant} className="flex-col items-start gap-2">
              <Button variant={variant}>{t(label)}</Button>
              <span className="type-code text-muted-foreground">{variant}</span>
              <span className="type-caption text-muted-foreground">
                {text.when[variant]}
              </span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part title={partTitle.sizes}>
        <Frame className="items-end">
          {sizes.map((size) => (
            <div key={size} className="flex flex-col items-start gap-1">
              <Button variant="secondary" size={size}>
                {t("gallery.button.openRequest")}
              </Button>
              <span className="type-code text-muted-foreground">{size}</span>
            </div>
          ))}
        </Frame>
      </Part>

      <Part {...text.withIcon}>
        <Frame>
          <Button>
            <PlusIcon data-icon="inline-start" />
            {t("gallery.common.newRequest")}
          </Button>
          <Button variant="secondary">
            {t("gallery.button.nextStep")}
            <ArrowRightIcon data-icon="inline-end" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            aria-label={t("gallery.common.settings")}
          >
            <SettingsIcon />
          </Button>
          <Button>
            {t("gallery.common.saveChanges")}
            <ButtonShortcut>Ctrl S</ButtonShortcut>
          </Button>
          <Button variant="secondary">
            {t("gallery.common.search")}
            <ButtonShortcut>Ctrl K</ButtonShortcut>
          </Button>
        </Frame>
      </Part>

      <Part {...pinnedStates}>
        <Frame>
          <StateMatrix
            rows={variants.map(({ variant }) => variant)}
            states={[
              "default",
              "hover",
              "focus",
              "active",
              "disabled",
              "loading",
            ]}
            render={(variant, state) =>
              state === "loading" ? (
                <Button variant={variant} disabled>
                  <LoaderCircleIcon
                    data-icon="inline-start"
                    className="animate-spin"
                  />
                  {t("gallery.common.saveChanges")}
                </Button>
              ) : (
                <Button variant={variant} {...previewState(state)}>
                  {t("gallery.common.saveChanges")}
                </Button>
              )
            }
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <Button className="max-w-full">
            <span className="truncate">{t("gallery.button.longSend")}</span>
          </Button>
          <Button variant="secondary" className="max-w-full">
            <PlusIcon data-icon="inline-start" />
            <span className="truncate">{t("gallery.button.longAdd")}</span>
          </Button>
        </Frame>
      </Part>
    </Section>
  )
}
