import { Input } from "@workspace/ui/components/input"
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSeparator,
  InputOTPSlot,
} from "@workspace/ui/components/input-otp"
import { useState } from "react"
import type { ReactNode } from "react"

import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

export function InputBlock() {
  return (
    <Section title="Input" note="One line of text typed by a person.">
      <Part title="Typical">
        <Frame className="max-w-sm">
          <Input aria-label="E-mail" placeholder="you@company.com" />
        </Frame>
      </Part>

      <Part
        title="States"
        note="Each state is pinned, so all of them are visible at once."
      >
        <Frame>
          <StateMatrix
            rows={["empty", "filled"] as const}
            states={["default", "hover", "focus", "disabled", "invalid"]}
            render={(row, state) => (
              <Input
                aria-label="Project name"
                className="w-40"
                placeholder="Project name"
                defaultValue={row === "filled" ? "Billing portal" : undefined}
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A value longer than the field scrolls inside it."
      >
        <Frame className="max-w-xs">
          <Input
            aria-label="Address"
            defaultValue="https://clientes.exemplo.com.br/projetos/portal-de-cobranca/solicitacoes"
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
  return (
    <Section
      title="InputOTP"
      note="A short code typed one character per box, such as the login code. Pasting the whole code fills every box."
    >
      <Part title="Typical">
        <Frame>
          <InputOTP aria-label="Login code" maxLength={6} inputMode="numeric">
            <Slots from={0} />
            <InputOTPSeparator />
            <Slots from={3} />
          </InputOTP>
        </Frame>
      </Part>

      <Part title="Filled, invalid and disabled">
        <Frame className="gap-6">
          <FilledInputOTP
            aria-label="Login code, filled"
            maxLength={6}
            initial="481027"
          >
            <Slots from={0} />
            <InputOTPSeparator />
            <Slots from={3} />
          </FilledInputOTP>
          <FilledInputOTP
            aria-label="Login code, wrong"
            maxLength={6}
            initial="481027"
            aria-invalid
          >
            <Slots from={0} />
            <InputOTPSeparator />
            <Slots from={3} />
          </FilledInputOTP>
          <FilledInputOTP
            aria-label="Login code, disabled"
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
