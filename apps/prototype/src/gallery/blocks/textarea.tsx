import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupText,
  InputGroupTextarea,
} from "@workspace/ui/components/input-group"
import { Kbd } from "@workspace/ui/components/kbd"
import { Textarea } from "@workspace/ui/components/textarea"
import { CopyIcon, SearchIcon, SendIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useFormat } from "#stores/language"

import { inputs } from "../explanations/inputs"
import { partTitle, pinnedStates } from "../explanations/shared"
import { freeText, sampleRequest } from "../free-text"
import { Frame, Part, Section } from "../parts"
import { StateMatrix, previewState } from "../state-matrix"

export function TextareaBlock() {
  const { t } = useTranslation()
  const text = inputs.textarea

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm">
          <Textarea
            aria-label={t("gallery.common.description")}
            placeholder={t("gallery.textarea.placeholder")}
          />
        </Frame>
      </Part>

      <Part {...pinnedStates}>
        <Frame>
          <StateMatrix
            rows={["empty", "filled"] as const}
            states={["default", "hover", "focus", "disabled", "invalid"]}
            render={(row, state) => (
              <Textarea
                aria-label={`${t("gallery.common.description")}, ${row}, ${state}`}
                className="w-36"
                placeholder={t("gallery.common.description")}
                defaultValue={
                  row === "filled" ? freeText.shortDescription : undefined
                }
                aria-invalid={state === "invalid" || undefined}
                {...previewState(state)}
              />
            )}
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <Textarea
            aria-label={t("gallery.textarea.longDescription")}
            defaultValue={freeText.longDescription}
          />
        </Frame>
      </Part>
    </Section>
  )
}

export function InputGroupBlock() {
  const { t } = useTranslation()
  const { number } = useFormat()
  const text = inputs.inputGroup

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-sm">
          <InputGroup>
            <InputGroupInput
              aria-label={t("gallery.common.searchRequests")}
              placeholder={t("gallery.common.searchRequests")}
            />
            <InputGroupAddon>
              <SearchIcon />
            </InputGroupAddon>
            <InputGroupAddon align="inline-end">
              <Kbd>/</Kbd>
            </InputGroupAddon>
          </InputGroup>
        </Frame>
      </Part>

      <Part {...text.attached}>
        <Frame className="max-w-sm flex-col flex-nowrap items-stretch">
          <InputGroup>
            <InputGroupAddon>
              <InputGroupText className="font-mono">https://</InputGroupText>
            </InputGroupAddon>
            <InputGroupInput
              aria-label={t("gallery.inputGroup.clientAddress")}
              defaultValue={freeText.clientAddress}
            />
          </InputGroup>
          <InputGroup>
            <InputGroupInput
              aria-label={t("gallery.inputGroup.hoursEstimated")}
              defaultValue={number(12)}
              inputMode="numeric"
            />
            <InputGroupAddon align="inline-end">
              <InputGroupText>
                {t("gallery.inputGroup.hoursUnit")}
              </InputGroupText>
            </InputGroupAddon>
          </InputGroup>
          <InputGroup>
            <InputGroupInput
              aria-label={t("gallery.inputGroup.requestLink")}
              defaultValue={sampleRequest.id}
              className="font-mono"
              readOnly
            />
            <InputGroupAddon align="inline-end">
              <InputGroupButton
                size="icon-xs"
                aria-label={t("gallery.common.copyLink")}
              >
                <CopyIcon />
              </InputGroupButton>
            </InputGroupAddon>
          </InputGroup>
          <InputGroup>
            <InputGroupTextarea
              aria-label={t("gallery.inputGroup.comment")}
              placeholder={t("gallery.inputGroup.commentPlaceholder")}
            />
            <InputGroupAddon align="block-end" className="justify-end">
              <InputGroupButton variant="default">
                <SendIcon />
                {t("gallery.inputGroup.send")}
              </InputGroupButton>
            </InputGroupAddon>
          </InputGroup>
        </Frame>
      </Part>

      <Part {...text.states}>
        <Frame>
          <StateMatrix
            rows={["group"] as const}
            states={["default", "focus", "disabled", "invalid"]}
            render={(_, state) => (
              <InputGroup className="w-44">
                <InputGroupInput
                  aria-label={`${t("gallery.common.search")}, ${state}`}
                  placeholder={t("gallery.common.search")}
                  aria-invalid={state === "invalid" || undefined}
                  {...previewState(state)}
                />
                <InputGroupAddon>
                  <SearchIcon />
                </InputGroupAddon>
              </InputGroup>
            )}
          />
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <InputGroup>
            <InputGroupAddon>
              <InputGroupText className="font-mono">
                {freeText.longHost}
              </InputGroupText>
            </InputGroupAddon>
            <InputGroupInput
              aria-label={t("gallery.inputGroup.projectPath")}
              defaultValue={freeText.longPath}
            />
          </InputGroup>
        </Frame>
      </Part>
    </Section>
  )
}
