import { useReducer } from "react"
import { useTranslation } from "react-i18next"

import { FramedPane } from "../../shared/components/framed-pane"
import { AdministratorStep } from "../components/administrator-step"
import { CodeStep } from "../components/code-step"
import { ConsultancyStep } from "../components/consultancy-step"
import { ReviewStep } from "../components/review-step"
import { type SetupForm, useSetupForm } from "../components/setup-form"
import { type SetupImages, useSetupImage } from "../components/setup-image"
import {
  type SetupAction,
  type SetupFlow,
  setupReducer,
  startOfTheSetup,
} from "./setup-flow"

/**
 * The setup: one route, whose steps are one reducer, whose fields are one
 * form beside it, and whose two images are held beside both, so that an
 * image outlives the step it was chosen on. Nothing of it outlives the
 * screen: a page loaded again starts at the step of the setup code, with
 * nothing typed and no image.
 */
export function SetupScreen() {
  const { t } = useTranslation("consultancy")
  const [flow, dispatch] = useReducer(setupReducer, startOfTheSetup)
  const form = useSetupForm()
  const code = flow.step === "code" ? undefined : flow.opened.code
  const codeStoppedWorking = () => dispatch({ type: "codeStoppedWorking" })
  // The file of each image is a field of the form, which concluding sends
  // and where a problem the API names about it shows. A problem goes when
  // the file changes: TanStack Form clears it on a change of the value.
  const logo = useSetupImage({
    purpose: "consultancy_logo",
    code,
    onFileChanged: (fileId) =>
      form.setFieldValue("consultancy.logo_file_id", fileId, {
        dontUpdateMeta: true,
      }),
    onCodeStoppedWorking: codeStoppedWorking,
  })
  const photo = useSetupImage({
    purpose: "user_photo",
    code,
    onFileChanged: (fileId) =>
      form.setFieldValue("administrator.photo_file_id", fileId, {
        dontUpdateMeta: true,
      }),
    onCodeStoppedWorking: codeStoppedWorking,
  })
  const images: SetupImages = { logo, photo }

  return (
    <FramedPane>
      <h1 className="type-display">{t("setup.heading")}</h1>
      <Step flow={flow} dispatch={dispatch} form={form} images={images} />
    </FramedPane>
  )
}

function Step({
  flow,
  dispatch,
  form,
  images,
}: {
  flow: SetupFlow
  dispatch: (action: SetupAction) => void
  form: SetupForm
  images: SetupImages
}) {
  switch (flow.step) {
    case "code":
      return (
        <CodeStep
          stoppedWorking={flow.resumeAt !== undefined}
          onOpened={(opened) => {
            dispatch({ type: "codeAccepted", opened })
            images.logo.resume(opened.code)
            images.photo.resume(opened.code)
          }}
        />
      )
    case "consultancy":
      return (
        <ConsultancyStep
          form={form}
          logo={images.logo}
          refused={flow.refused}
          onProblemsShown={() =>
            dispatch({ type: "problemsShown", step: "consultancy" })
          }
          onContinue={() => dispatch({ type: "wentTo", step: "administrator" })}
        />
      )
    case "administrator":
      return (
        <AdministratorStep
          form={form}
          photo={images.photo}
          opens={flow.opened.opens}
          refused={flow.refused}
          onProblemsShown={() =>
            dispatch({ type: "problemsShown", step: "administrator" })
          }
          onContinue={() => dispatch({ type: "wentTo", step: "review" })}
          // The way back in has no step of the consultancy.
          onBack={
            flow.opened.opens === "first_setup"
              ? () => dispatch({ type: "wentTo", step: "consultancy" })
              : undefined
          }
        />
      )
    case "review":
      return (
        <ReviewStep
          form={form}
          images={images}
          opened={flow.opened}
          onGoTo={(step) => dispatch({ type: "wentTo", step })}
          onFieldsRefused={(fields) =>
            dispatch({ type: "fieldsRefused", fields })
          }
          onCodeStoppedWorking={() => dispatch({ type: "codeStoppedWorking" })}
        />
      )
  }
}
