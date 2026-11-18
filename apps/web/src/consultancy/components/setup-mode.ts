import type { ParseKeys } from "i18next"

import type { OpenedSetupCode } from "../data/setup"

/** A key of the translations of the area, which `t` accepts. */
type Text = ParseKeys<"consultancy">

/**
 * The texts of the setup that differ between its two modes, by their keys in
 * the translations of the area.
 */
type TextsOfTheMode = Record<
  | "administratorTitle"
  | "administratorIntro"
  | "reviewIntro"
  | "changeAdministrator"
  | "logsInWith"
  | "imagesSending"
  | "imagesNotSent"
  | "send",
  Text
>

/**
 * What a setup code opened decides the mode of the setup: the first setup,
 * which registers the consultancy and its first administrator, or the way
 * back in, which registers a new administrator and has no step of the
 * consultancy. This is the one place that tells the two apart.
 */
export type SetupMode = {
  /** Whether the setup has the step of the consultancy, and sends it. */
  hasConsultancy: boolean
  texts: TextsOfTheMode
}

const modes: Record<OpenedSetupCode["opens"], SetupMode> = {
  first_setup: {
    hasConsultancy: true,
    texts: {
      administratorTitle: "setup.administrator.title",
      administratorIntro: "setup.administrator.intro",
      reviewIntro: "setup.review.intro",
      changeAdministrator: "setup.review.changeAdministrator",
      logsInWith: "setup.review.logsInWith",
      imagesSending: "setup.review.imagesSending",
      imagesNotSent: "setup.review.imageNotSent",
      send: "setup.review.conclude",
    },
  },
  new_administrator: {
    hasConsultancy: false,
    texts: {
      administratorTitle: "setup.newAdministrator.title",
      administratorIntro: "setup.newAdministrator.intro",
      reviewIntro: "setup.review.newAdministrator.intro",
      changeAdministrator: "setup.review.newAdministrator.change",
      logsInWith: "setup.review.newAdministrator.logsInWith",
      imagesSending: "setup.review.newAdministrator.photoSending",
      imagesNotSent: "setup.review.newAdministrator.photoNotSent",
      send: "setup.review.newAdministrator.register",
    },
  },
}

/** The mode of the setup a setup code opened. */
export function modeOf(opened: Pick<OpenedSetupCode, "opens">): SetupMode {
  return modes[opened.opens]
}
