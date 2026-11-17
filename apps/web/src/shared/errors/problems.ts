import type { z } from "zod"

import { ApiError } from "../api/request"

/**
 * What is wrong with one field: a code the app translates and the values its
 * text is filled with, as the API names them (`too_long`, `{max_length: 50}`).
 */
export type FieldProblem = { code: string; params: Record<string, unknown> }

/**
 * The problem of each field, by the name of the field in the form, which is
 * its place in the body: `consultancy.address.city`.
 */
export type FieldProblems = Record<string, FieldProblem>

/** The code of a failure the API answered, such as `setup_code_invalid`. */
export function answerCodeOf(error: unknown): string | undefined {
  return error instanceof ApiError ? error.problem.code : undefined
}

/**
 * The fields an answer of the API refused, when it refused fields: a 422
 * `validation_failed`, whose `errors` name each field by its place in the
 * body (`body.code`). Anything else names no field.
 */
export function refusedFields(error: unknown): FieldProblems | undefined {
  if (
    !(error instanceof ApiError) ||
    error.problem.code !== "validation_failed"
  ) {
    return undefined
  }
  const problems: FieldProblems = {}
  for (const { location, code, params } of error.problem.errors ?? []) {
    if (location.startsWith("body.")) {
      problems[location.slice("body.".length)] = { code, params }
    }
  }
  return Object.keys(problems).length > 0 ? problems : undefined
}

/** A broken rule of the contract, in the codes the API answers with. */
function problemOf(issue: z.core.$ZodIssue): FieldProblem {
  if (issue.code === "too_small" && issue.minimum === 1) {
    return { code: "required", params: {} }
  }
  if (issue.code === "too_big") {
    return { code: "too_long", params: { max_length: Number(issue.maximum) } }
  }
  return { code: "invalid", params: {} }
}

/** A validator of TanStack Form, of a form or of a group of its fields. */
type ContractValidator<T> = ({
  value,
}: {
  value: T
}) => { fields: FieldProblems } | undefined

/**
 * A validator of TanStack Form that checks what the contract declares, with
 * the schema orval generated, and names the problem of each field with the
 * codes the API itself answers with. A form checks nothing else (Forms and
 * validation in CODING_STANDARDS.md).
 *
 * A form whose values are not the body as it is sent, such as one that holds
 * an empty text for an optional field the body leaves out, gives `toBody`,
 * which turns its values into the body. The name of a field is then its
 * place in the body, which the form names its field with too.
 */
export function checkTheContract<T>(
  contract: z.ZodType<unknown, T>
): ContractValidator<T>
export function checkTheContract<Typed, T>(
  contract: z.ZodType<unknown, T>,
  toBody: (typed: Typed) => T
): ContractValidator<Typed>
export function checkTheContract(
  contract: z.ZodType,
  toBody: (typed: unknown) => unknown = (typed) => typed
): ContractValidator<unknown> {
  return ({ value }) => {
    const result = contract.safeParse(toBody(value))
    if (result.success) {
      return undefined
    }
    // One problem for each field: an empty field is `required`, whatever
    // else its value breaks.
    const problems = result.error.issues
      .map((issue) => ({ field: issue.path.join("."), ...problemOf(issue) }))
      .sort(
        (a, b) => Number(b.code === "required") - Number(a.code === "required")
      )
    const fields: FieldProblems = {}
    for (const { field, code, params } of problems) {
      fields[field] ??= { code, params }
    }
    return { fields }
  }
}
