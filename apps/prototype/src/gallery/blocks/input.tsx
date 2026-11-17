import { Input } from "@workspace/ui/components/input"
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSeparator,
  InputOTPSlot,
} from "@workspace/ui/components/input-otp"
import { useState } from "react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { inputs } from "../explanations/inputs"
import { partTitle, pinnedStates } from "../explanations/shared"
import { freeText } from "../free-text"
import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

export function InputBlock() {
  const { t } = useTranslation()
  const text = inputs.input

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm">
          <Input
            aria-label={t("gallery.common.email")}
            placeholder={t("gallery.common.emailPlaceholder")}
          />
        </Frame>
      </Part>

      <Part {...pinnedStates}>
        <Frame>
          <StateMatrix
            rows={["empty", "filled"] as const}
            states={["default", "hover", "focus", "disabled", "invalid"]}
            render={(row, state) => (
              <Input
                aria-label={`${t("gallery.common.projectName")}, ${row}, ${state}`}
                className="w-40"
                placeholder={t("gallery.common.projectName")}
                defaultValue={row === "filled" ? freeText.project : undefined}
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <Input
            aria-label={t("gallery.input.address")}
            defaultValue={freeText.longAddress}
          />
        </Frame>
      </Part>
    </Section>
  )
}

/**
 * A code field that starts filled. It holds its own value, because the
 * library under InputOTP warns in the console when it is given a default one.
 */
export function FilledInputOTP({
  initial,
  children,
  ...props
}: {
  initial: string
  maxLength: number
  children: ReactNode
  id?: string
  disabled?: boolean
  "aria-label"?: string
  "aria-invalid"?: boolean
}) {
  const [value, setValue] = useState(initial)
  return (
    <InputOTP value={value} onChange={setValue} {...props}>
      {children}
    </InputOTP>
  )
}

function Slots({ from }: { from: number }) {
  return (
    <InputOTPGroup>
      <InputOTPSlot index={from} />
      <InputOTPSlot index={from + 1} />
      <InputOTPSlot index={from + 2} />
    </InputOTPGroup>
  )
}

export function InputOTPBlock() {
  const { t } = useTranslation()
  const text = inputs.inputOTP
  const label = t("gallery.common.loginCode")

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <InputOTP aria-label={label} maxLength={6} inputMode="numeric">
            <Slots from={0} />
            <InputOTPSeparator />
            <Slots from={3} />
          </InputOTP>
        </Frame>
      </Part>

      <Part {...text.filled}>
        <Frame className="gap-6">
          <FilledInputOTP
            aria-label={`${label}, filled`}
            maxLength={6}
            initial="481027"
          >
            <Slots from={0} />
            <InputOTPSeparator />
            <Slots from={3} />
          </FilledInputOTP>
          <FilledInputOTP
            aria-label={`${label}, invalid`}
            maxLength={6}
            initial="481027"
            aria-invalid
          >
            <Slots from={0} />
            <InputOTPSeparator />
            <Slots from={3} />
          </FilledInputOTP>
          <FilledInputOTP
            aria-label={`${label}, disabled`}
            maxLength={6}
            initial="481"
            disabled
          >
            <Slots from={0} />
            <InputOTPSeparator />
            <Slots from={3} />
          </FilledInputOTP>
        </Frame>
      </Part>
    </Section>
  )
}
