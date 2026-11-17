import { Button } from "@workspace/ui/components/button"
import { Toaster, toast } from "@workspace/ui/components/toast"
import { useTranslation } from "react-i18next"

import { feedback } from "../explanations/feedback"
import { partTitle } from "../explanations/shared"
import { freeText, sampleRequest } from "../free-text"
import { Frame, Part, Section } from "../parts"

const text = feedback.toast

/** The identifier of a request that was just created. */
const createdId = "REQ-0143"

/**
 * A toast is written when it is shown, in the language of that moment: one
 * that is already on the screen keeps its text when the language changes.
 */
export function ToastBlock() {
  const { t } = useTranslation()

  const exportFailed = {
    title: t("gallery.common.exportFailed"),
    description: t("gallery.common.nothingChanged"),
  }

  const kinds = [
    {
      ...text.kind.plain,
      show: () =>
        toast.add({
          title: t("gallery.toast.draftSaved.title"),
          description: t("gallery.toast.draftSaved.description"),
        }),
    },
    {
      ...text.kind.success,
      show: () =>
        toast.add({
          type: "success",
          title: t("gallery.toast.requestCreated.title"),
          description: t("gallery.toast.requestCreated.description", {
            id: createdId,
            project: freeText.project,
          }),
        }),
    },
    {
      ...text.kind.warning,
      show: () =>
        toast.add({
          type: "warning",
          title: t("gallery.toast.savedWithoutClient.title"),
          description: t("gallery.toast.savedWithoutClient.description"),
        }),
    },
    {
      ...text.kind.error,
      show: () =>
        toast.add({ type: "error", priority: "high", ...exportFailed }),
    },
    {
      ...text.kind.info,
      show: () =>
        toast.add({
          type: "info",
          title: t("gallery.toast.newVersion.title"),
          description: t("gallery.toast.newVersion.description"),
        }),
    },
  ]

  const exportProject = () => {
    void toast.promise(
      new Promise<void>((resolve) => {
        window.setTimeout(resolve, 2000)
      }),
      {
        loading: {
          title: t("gallery.toast.exporting.title"),
          description: t("gallery.toast.exporting.description"),
        },
        success: {
          title: t("gallery.toast.exportReady.title"),
          description: t("gallery.toast.exportReady.description", {
            file: freeText.projectFile,
          }),
        },
        error: exportFailed,
      }
    )
  }

  const archiveRequest = () => {
    toast.add({
      title: t("gallery.toast.archived.title"),
      description: t("gallery.toast.archived.description", {
        id: sampleRequest.id,
      }),
      actionProps: {
        children: t("gallery.toast.archived.undo"),
        onClick: () =>
          toast.add({
            type: "success",
            title: t("gallery.toast.archived.restored"),
          }),
      },
    })
  }

  const showLongToast = () => {
    toast.add({
      type: "warning",
      title: t("gallery.toast.long.title"),
      description: t("gallery.toast.long.description"),
      actionProps: { children: t("gallery.toast.long.action") },
    })
  }

  return (
    <Section {...text.about}>
      {/* The viewport sits above the prototype bar, which takes the bottom of the screen. */}
      <Toaster className="bottom-16" closeLabel={t("gallery.toast.close")} />

      <Part title={partTitle.typical}>
        <Frame>
          <Button variant="secondary" onClick={kinds[1]?.show}>
            {text.kind.success.label}
          </Button>
        </Frame>
      </Part>

      <Part {...text.kinds}>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(15rem,1fr))] gap-3">
          {kinds.map(({ label, when, show }) => (
            <Frame key={label} className="flex-col items-start gap-2">
              <Button variant="outline" onClick={show}>
                {label}
              </Button>
              <span className="type-caption text-muted-foreground">{when}</span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part {...text.loading}>
        <Frame>
          <Button variant="outline" onClick={exportProject}>
            {text.showLoading}
          </Button>
          <Button variant="outline" onClick={archiveRequest}>
            {text.showAction}
          </Button>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame>
          <Button variant="outline" onClick={showLongToast}>
            {text.showLong}
          </Button>
        </Frame>
      </Part>
    </Section>
  )
}
