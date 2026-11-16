import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@workspace/ui/components/alert"
import { Button } from "@workspace/ui/components/button"
import { CircleXIcon, InfoIcon } from "lucide-react"

import { Frame, Part, Section } from "../parts"

export function AlertBlock() {
  return (
    <Section
      title="Alert"
      note="A message that stays on the screen, beside what it is about. For a message about something that just happened, use a toast."
    >
      <Part title="Typical">
        <Frame className="max-w-xl">
          <Alert>
            <InfoIcon />
            <AlertTitle>Billing closes on Friday</AlertTitle>
            <AlertDescription>
              A request closed after that day goes on the invoice of next month.
            </AlertDescription>
          </Alert>
        </Frame>
      </Part>

      <Part
        title="Variants"
        note="The destructive variant reports an error. It always carries an icon, so it does not depend on colour."
      >
        <Frame className="max-w-xl flex-col flex-nowrap items-stretch">
          <Alert>
            <InfoIcon />
            <AlertTitle>This project is archived</AlertTitle>
            <AlertDescription>
              Its requests can be read, and no new one can be opened.
            </AlertDescription>
          </Alert>
          <Alert variant="destructive">
            <CircleXIcon />
            <AlertTitle>The export failed</AlertTitle>
            <AlertDescription>
              The file was not created. Nothing was changed in the project.
            </AlertDescription>
          </Alert>
        </Frame>
      </Part>

      <Part
        title="What it can carry"
        note="A title alone, a description alone, and an action at the end."
      >
        <Frame className="max-w-xl flex-col flex-nowrap items-stretch">
          <Alert>
            <InfoIcon />
            <AlertTitle>Two people are editing this request.</AlertTitle>
          </Alert>
          <Alert>
            <AlertDescription>
              Hours logged today can be changed until midnight.
            </AlertDescription>
          </Alert>
          <Alert variant="destructive">
            <CircleXIcon />
            <AlertTitle>The export failed</AlertTitle>
            <AlertDescription>The file was not created.</AlertDescription>
            <AlertAction>
              <Button variant="outline" size="xs">
                Try again
              </Button>
            </AlertAction>
          </Alert>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long title and a long description wrap inside the alert."
      >
        <Frame className="max-w-xs">
          <Alert variant="destructive">
            <CircleXIcon />
            <AlertTitle>
              Não foi possível enviar a solicitação para a aprovação do cliente
            </AlertTitle>
            <AlertDescription>
              O responsável pelo projeto ainda não tem um e-mail cadastrado.
              Cadastre o e-mail e envie a solicitação outra vez.
            </AlertDescription>
          </Alert>
        </Frame>
      </Part>
    </Section>
  )
}
