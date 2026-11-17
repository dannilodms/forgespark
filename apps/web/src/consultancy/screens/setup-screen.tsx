import { useReducer } from "react"
import { useTranslation } from "react-i18next"

import { FramedPane } from "../../shared/components/framed-pane"
import { CodeStep } from "../components/code-step"
import { NextStepPlaceholder } from "../components/next-step-placeholder"
import { type SetupFlow, setupReducer, startOfTheSetup } from "./setup-flow"

/**
 * The setup: one route, whose steps are one reducer. Nothing of it outlives
 * the screen: a page loaded again starts at the step of the setup code.
 */
export function SetupScreen() {
  const { t } = useTranslation("consultancy")
  const [flow, dispatch] = useReducer(setupReducer, startOfTheSetup)

  return (
    <FramedPane>
      <h1 className="type-display">{t("setup.heading")}</h1>
      <Step
        flow={flow}
        onOpened={(opened) => dispatch({ type: "codeAccepted", opened })}
      />
    </FramedPane>
  )
}

function Step({
  flow,
  onOpened,
}: {
  flow: SetupFlow
  onOpened: Parameters<typeof CodeStep>[0]["onOpened"]
}) {
  switch (flow.step) {
    case "code":
      return <CodeStep onOpened={onOpened} />
    case "consultancy":
    case "administrator":
      return <NextStepPlaceholder opened={flow.opened} />
  }
}
