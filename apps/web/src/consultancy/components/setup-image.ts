import { useEffect, useReducer, useRef } from "react"

import { answerCodeOf } from "../../shared/errors/problems"
import { type ImagePurpose, sendSetupImage } from "../data/setup"

/**
 * An image chosen in the setup: the file, and the address of the browser's
 * own copy of it, which shows it on the screen before and after it is sent.
 */
export type ChosenImage = { file: File; preview: string }

/**
 * Where the upload of one image of the setup is: one flow of named steps
 * (State in CODING_STANDARDS.md).
 *
 * - `none`: no image is chosen, or the one chosen was removed.
 * - `sending`: the image is on its way to the storage.
 * - `sent`: the storage keeps it, and the form holds its file.
 * - `failed`: the API or the storage refused it, or did not answer.
 * - `waitingForCode`: the setup code stopped working before the image was
 *   sent; a new code sends it.
 */
export type SetupImage =
  | { step: "none" }
  | { step: "sending"; image: ChosenImage }
  | { step: "sent"; image: ChosenImage }
  | { step: "failed"; image: ChosenImage; error: unknown }
  | { step: "waitingForCode"; image: ChosenImage }

type SetupImageAction =
  /** An image starts its way to the storage: chosen, or sent again. */
  | { type: "sending"; image: ChosenImage }
  /** The image being sent is kept by the storage. */
  | { type: "sent" }
  /** The image being sent did not reach the storage. */
  | { type: "failed"; error: unknown }
  /** An image waits for a setup code that works. */
  | { type: "waiting"; image: ChosenImage }
  /** The person removed the image. */
  | { type: "removed" }

function setupImageReducer(
  upload: SetupImage,
  action: SetupImageAction
): SetupImage {
  switch (action.type) {
    case "sending":
      return { step: "sending", image: action.image }
    case "sent":
      return upload.step === "sending"
        ? { step: "sent", image: upload.image }
        : upload
    case "failed":
      return upload.step === "sending"
        ? { step: "failed", image: upload.image, error: action.error }
        : upload
    case "waiting":
      return { step: "waitingForCode", image: action.image }
    case "removed":
      return { step: "none" }
    default:
      return upload
  }
}

type SetupImageOptions = {
  purpose: ImagePurpose
  /** The setup code that opened the setup, while one has. */
  code: string | undefined
  /**
   * The file of the image on the screen, once it is sent, or an empty text
   * while there is none: what concluding sends.
   */
  onFileChanged: (fileId: string) => void
  /** A request answered that the setup code no longer works. */
  onCodeStoppedWorking: () => void
}

/**
 * One image of the setup: the logo of the consultancy or the photo of the
 * administrator. It is held by the screen, beside the form, so that it
 * outlives the step it is chosen on: it is still sent while the person types
 * the next step, and the review shows it.
 *
 * An image is sent as soon as it is chosen, and shown meanwhile from the
 * browser's own copy. Only the image on the screen counts: a result that
 * arrives for one replaced or removed is dropped. When the setup code stops
 * working while an image is sent, the image waits, and `resume` sends it
 * with a new code; an image already sent stays valid with the new code.
 */
export function useSetupImage({
  purpose,
  code,
  onFileChanged,
  onCodeStoppedWorking,
}: SetupImageOptions) {
  const [upload, dispatch] = useReducer(setupImageReducer, { step: "none" })
  // The image on the screen, which a result that arrives must be about.
  const onTheScreen = useRef<ChosenImage | undefined>(undefined)

  // The browser's copy of the image is let go of when the screen closes.
  useEffect(
    () => () => {
      if (onTheScreen.current) {
        URL.revokeObjectURL(onTheScreen.current.preview)
      }
    },
    []
  )

  async function send(image: ChosenImage, withCode: string) {
    dispatch({ type: "sending", image })
    try {
      const fileId = await sendSetupImage({
        purpose,
        file: image.file,
        code: withCode,
      })
      if (onTheScreen.current !== image) {
        return
      }
      onFileChanged(fileId)
      dispatch({ type: "sent" })
    } catch (error) {
      if (onTheScreen.current !== image) {
        return
      }
      if (answerCodeOf(error) === "setup_code_invalid") {
        dispatch({ type: "waiting", image })
        onCodeStoppedWorking()
        return
      }
      dispatch({ type: "failed", error })
    }
  }

  /** Puts an image on the screen in place of any before it, or none. */
  function show(image: ChosenImage | undefined) {
    if (onTheScreen.current) {
      URL.revokeObjectURL(onTheScreen.current.preview)
    }
    onTheScreen.current = image
    onFileChanged("")
  }

  /** Takes a file the person chose, in place of any image before it. */
  function choose(file: File) {
    const image = { file, preview: URL.createObjectURL(file) }
    show(image)
    if (code === undefined) {
      dispatch({ type: "waiting", image })
      return
    }
    void send(image, code)
  }

  /** Removes the image, sent or not: concluding sends none. */
  function remove() {
    show(undefined)
    dispatch({ type: "removed" })
  }

  /** Sends again an image that failed, with an upload of its own. */
  function tryAgain() {
    if (upload.step !== "failed" || code === undefined) {
      return
    }
    void send(upload.image, code)
  }

  /**
   * A new setup code opened the setup: an image that waited for one is sent
   * with it. An image already sent stays as it is.
   */
  function resume(newCode: string) {
    if (upload.step !== "waitingForCode") {
      return
    }
    void send(upload.image, newCode)
  }

  return { upload, choose, remove, tryAgain, resume }
}

/** One image of the setup, as the screen holds it. */
export type SetupImageField = ReturnType<typeof useSetupImage>

/** The two images of the setup. */
export type SetupImages = { logo: SetupImageField; photo: SetupImageField }

/**
 * Why the images on the screen are not all in the storage yet, which keeps
 * them from being sent with the setup: one is still on its way there
 * (`sending`, waiting for a setup code included), or one did not reach it
 * (`notSent`). Nothing when every image chosen is stored, or none was chosen.
 */
export function whyTheImagesAreNotStored(
  images: SetupImages
): "sending" | "notSent" | undefined {
  const steps = [images.logo.upload.step, images.photo.upload.step]
  if (steps.includes("failed")) {
    return "notSent"
  }
  if (steps.includes("sending") || steps.includes("waitingForCode")) {
    return "sending"
  }
  return undefined
}
