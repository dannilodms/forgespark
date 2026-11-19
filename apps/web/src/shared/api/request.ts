import type { Problem } from "./generated/model"

/**
 * A failure the API answered in its error format (ADR-0009): its `code` is
 * what the app translates, and its `trace_id` finds it in the logs.
 */
export class ApiError extends Error {
  readonly problem: Problem

  constructor(problem: Problem) {
    super(problem.detail ?? problem.title)
    this.name = "ApiError"
    this.problem = problem
  }
}

/** Whether a body is in the error format of the API. */
function isProblem(body: unknown): body is Problem {
  return (
    typeof body === "object" &&
    body !== null &&
    "code" in body &&
    typeof body.code === "string" &&
    "trace_id" in body
  )
}

/** What the app does when the API answers that nobody is logged in. */
let onUnauthorized: () => void = () => {}

/**
 * Tells `listener` of every 401 the API answers, before the request rejects:
 * the app sets it once, when it is put together, to open the login when a
 * session ends while the person works (`access` decides what that means).
 * This file has no router, and every request passes through it, so a 401
 * reaches the app whichever read or write met it. A later call replaces the
 * listener.
 */
export function whenUnauthorized(listener: () => void) {
  onUnauthorized = listener
}

/**
 * The fetch of the generated client. It resolves with the body of a success,
 * and rejects with an `ApiError` when the API answers in its error format.
 * Anything else, such as a network that fails, rejects as fetch does.
 */
export async function request<T>(url: string, init: RequestInit): Promise<T> {
  const response = await fetch(url, init)
  if (response.status === 401) {
    onUnauthorized()
  }
  // `application/json` and `application/problem+json`. A 204 has no body,
  // and neither has a 201 that points at nothing, such as the registration
  // of an administrator.
  const isJSON = response.headers.get("Content-Type")?.includes("json")
  const body: unknown = isJSON ? await response.json() : undefined
  if (!response.ok) {
    if (isProblem(body)) {
      throw new ApiError(body)
    }
    throw new Error(`The API answered ${response.status} out of its format`)
  }
  return body as T
}
