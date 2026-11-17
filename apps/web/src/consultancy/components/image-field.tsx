import { ImageIcon } from "lucide-react"
import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"
import { Field, FieldDescription, FieldLabel } from "#ui/components/field"
import { StatusIndicator } from "#ui/components/status-indicator"

import { FieldProblemMessage } from "../../shared/errors/field-problem"
import type { FieldProblem } from "../../shared/errors/problems"
import { largestImage, refusalOf } from "../data/setup"
import type { ChosenImage, SetupImage, SetupImageField } from "./setup-image"

/** The words of one image field, which name what the image is. */
type ImageTexts = {
  /** The label of the field. */
  label: string
  /** The button that chooses an image when there is none. */
  choose: string
  /** The name of the button that chooses another image in its place. */
  replace: string
}

/**
 * The field of TanStack Form that holds the file of an image, as far as the
 * image field reads it: its name, and the problem the API named on it.
 */
type ImageFieldApi = {
  name: string
  state: { meta: { errors: unknown[] } }
}

/**
 * The field of an image of the setup: the image chosen, shown from the
 * browser's own copy, how far it is on its way to the storage, and the
 * buttons that choose it. The image is optional.
 *
 * The control is the button that opens the browser's file picker; the input
 * of the file behind it is hidden. A file that is not an image the setup
 * accepts is refused here, with the reason, and is not sent: the image
 * chosen before, if any, stays. A problem the API named on the field shows
 * the same way.
 */
export function ImageField({
  field,
  image,
  texts,
}: {
  field: ImageFieldApi
  image: SetupImageField
  texts: ImageTexts
}) {
  const { t } = useTranslation("consultancy")
  const { t: tShared } = useTranslation()
  const input = useRef<HTMLInputElement>(null)
  const [refusal, setRefusal] = useState<FieldProblem>()
  const id = `field-${field.name.replaceAll(".", "-")}`
  const problem =
    refusal ?? (field.state.meta.errors[0] as FieldProblem | undefined)
  const { upload } = image
  const chosen = upload.step === "none" ? undefined : upload.image

  function take(file: File) {
    const refused = refusalOf(file)
    setRefusal(refused)
    if (!refused) {
      image.choose(file)
    }
  }

  return (
    <Field aria-labelledby={`${id}-label`}>
      <FieldLabel id={`${id}-label`}>
        {tShared("form.optional", { label: texts.label })}
      </FieldLabel>
      {/* The browser's file picker, which the button opens. */}
      <input
        ref={input}
        type="file"
        accept="image/png,image/jpeg,image/webp"
        hidden
        aria-label={texts.choose}
        onChange={(event) => {
          const file = event.target.files?.[0]
          // The same file can be chosen again after it was removed.
          event.target.value = ""
          if (file) {
            take(file)
          }
        }}
      />
      <div className="flex items-center gap-3">
        <ImageThumbnail image={chosen} />
        <div className="flex min-w-0 flex-col items-start gap-1.5">
          <p
            role="status"
            className="flex max-w-full min-w-0 items-center gap-2 empty:hidden"
          >
            {chosen ? (
              <>
                <span className="truncate type-caption text-muted-foreground">
                  {chosen.file.name}
                </span>
                <ImageStatus upload={upload} />
              </>
            ) : null}
          </p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            data-field={field.name}
            aria-label={chosen ? texts.replace : undefined}
            aria-describedby={[
              `${id}-description`,
              problem ? `${id}-problem` : "",
            ]
              .filter(Boolean)
              .join(" ")}
            onClick={() => input.current?.click()}
          >
            {chosen ? t("setup.image.replace") : texts.choose}
          </Button>
        </div>
      </div>
      <FieldDescription id={`${id}-description`}>
        {t("setup.image.description", { max_size: largestImage })}
      </FieldDescription>
      {problem ? (
        <FieldProblemMessage id={`${id}-problem`} problem={problem} />
      ) : null}
    </Field>
  )
}

/** The image chosen, small, or the place where it shows once chosen. */
export function ImageThumbnail({ image }: { image: ChosenImage | undefined }) {
  return (
    <span className="flex size-16 shrink-0 items-center justify-center overflow-hidden rounded-md border bg-background">
      {image ? (
        <img
          src={image.preview}
          alt={image.file.name}
          className="size-full object-contain"
        />
      ) : (
        <ImageIcon
          aria-hidden="true"
          className="size-4 text-faint-foreground"
        />
      )}
    </span>
  )
}

/** How far an image is on its way to the storage: a sign and a word. */
export function ImageStatus({ upload }: { upload: SetupImage }) {
  const { t } = useTranslation("consultancy")

  switch (upload.step) {
    case "none":
      return null
    case "sending":
      return (
        <StatusIndicator status="working">
          {t("setup.image.sending")}
        </StatusIndicator>
      )
    case "sent":
      return (
        <StatusIndicator status="done">{t("setup.image.sent")}</StatusIndicator>
      )
    case "failed":
      return (
        <StatusIndicator status="error">
          {t("setup.image.notSent")}
        </StatusIndicator>
      )
    case "waitingForCode":
      return (
        <StatusIndicator status="attention">
          {t("setup.image.waiting")}
        </StatusIndicator>
      )
  }
}
