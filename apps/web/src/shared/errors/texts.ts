import { useTranslation } from "react-i18next"

import { ApiError } from "../api/request"
import sharedTexts from "../locales/en-US.json"
import type { FieldProblem } from "./problems"

/** What a failure is, in the language of the screen. */
export type ErrorText = {
  title: string
  todo: string
  /** The trace identifier of the answer, to find it in the logs. */
  traceId?: string
}

type AnswerCode = keyof typeof sharedTexts.errors.answers
type FieldCode = keyof typeof sharedTexts.errors.fields

function isAnswerCode(code: string): code is AnswerCode {
  return Object.hasOwn(sharedTexts.errors.answers, code)
}

function isFieldCode(code: string): code is FieldCode {
  return Object.hasOwn(sharedTexts.errors.fields, code)
}

/**
 * Translates failures by their code and params (Errors in
 * CODING_STANDARDS.md). A code with no text of its own reads as unexpected,
 * and still shows its trace identifier.
 */
export function useErrorTexts() {
  const { t } = useTranslation()

  /** A failure of a request, or of anything else. */
  function ofError(error: unknown): ErrorText {
    if (error instanceof ApiError) {
      const { code, params, trace_id: traceId } = error.problem
      if (isAnswerCode(code)) {
        return {
          title: t(`errors.answers.${code}.title`, params),
          todo: t(`errors.answers.${code}.todo`, params),
          traceId,
        }
      }
      return {
        title: t("errors.unexpected.title"),
        todo: t("errors.unexpected.todo"),
        traceId,
      }
    }
    // fetch rejects with a TypeError when the network fails.
    if (error instanceof TypeError) {
      return {
        title: t("errors.unreachable.title"),
        todo: t("errors.unreachable.todo"),
      }
    }
    return {
      title: t("errors.unexpected.title"),
      todo: t("errors.unexpected.todo"),
    }
  }

  /** What is wrong with a field. A code with no text of its own reads as `invalid`. */
  function ofField({ code, params }: FieldProblem): string {
    return isFieldCode(code)
      ? t(`errors.fields.${code}`, params)
      : t("errors.fields.invalid")
  }

  return { ofError, ofField }
}
