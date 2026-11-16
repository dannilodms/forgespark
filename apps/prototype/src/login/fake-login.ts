/*
 * Fake data. There is no backend behind the prototype: each function waits a
 * moment and answers as the real operation would. Every attempt succeeds
 * here; the scenarios of the login screen come with their own issue.
 */
import type { LoginProviderId } from "./login-flow"

function answerAfter(milliseconds: number) {
  return new Promise<void>((resolve) => setTimeout(resolve, milliseconds))
}

/** Asks for a login code. The answer is the same whoever the e-mail belongs to. */
export async function askForLoginCode(_email: string) {
  await answerAfter(700)
}

/** Checks a login code. */
export async function checkLoginCode(_code: string) {
  await answerAfter(700)
  return { loggedIn: true }
}

/** Goes to a login provider and comes back. The prototype shows no page of the provider. */
export async function logInThrough(_provider: LoginProviderId) {
  await answerAfter(900)
  return { loggedIn: true }
}
