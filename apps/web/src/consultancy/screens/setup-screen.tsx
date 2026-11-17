import { useReducer } from "react"
import { useTranslation } from "react-i18next"

import { FramedPane } from "../../shared/components/framed-pane"
import { AdministratorStep } from "../components/administrator-step"
import { CodeStep } from "../components/code-step"
import { ConsultancyStep } from "../components/consultancy-step"
import { NextStepPlaceholder } from "../components/next-step-placeholder"
import { ReviewStep } from "../components/review-step"
import { type SetupForm, useSetupForm } from "../components/setup-form"
import {
  type SetupAction,
  type SetupFlow,
  setupReducer,
  startOfTheSetup,
} from "./setup-flow"

/**
 * The setup: one route, whose steps are one reducer, and whose fields are one
 * form beside it. Nothing of it outlives the screen: a page loaded again
 * starts at the step of the setup code, with nothing typed.
 */
export function SetupScreen() {
  const { t } = useTranslation("consultancy")
  const [flow, dispatch] = useReducer(setupReducer, startOfTheSetup)
  const form = useSetupForm()

  return (
    <FramedPane>
      <h1 className="type-display">{t("setup.heading")}</h1>
      <Step flow={flow} dispatch={dispatch} form={form} />
    </FramedPane>
  )
}

function Step({
  flow,
  dispatch,
  form,
}: {
  flow: SetupFlow
  dispatch: (action: SetupAction) => void
  form: SetupForm
}) {
  switch (flow.step) {
    case "code":
      return (
        <CodeStep
          stoppedWorking={flow.resumeAt !== undefined}
          onOpened={(opened) => dispatch({ type: "codeAccepted", opened })}
        />
      )
    case "consultancy":
      return (
        <ConsultancyStep
          form={form}
          onContinue={() => dispatch({ type: "wentTo", step: "administrator" })}
        />
      )
    case "administrator":
      return (
        <AdministratorStep
          form={form}
          onContinue={() => dispatch({ type: "wentTo", step: "review" })}
          onBack={() => dispatch({ type: "wentTo", step: "consultancy" })}
        />
      )
    case "review":
      return (
        <ReviewStep
          form={form}
          opened={flow.opened}
          onGoTo={(step) => dispatch({ type: "wentTo", step })}
          onCodeStoppedWorking={() => dispatch({ type: "codeStoppedWorking" })}
        />
      )
    case "newAdministrator":
      return <NextStepPlaceholder />
  }
}
