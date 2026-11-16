import { Button } from "@workspace/ui/components/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@workspace/ui/components/dialog"
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@workspace/ui/components/sheet"
import { StatusIndicator } from "@workspace/ui/components/status-indicator"

import { Frame, Part, Section } from "../parts"

export function DialogBlock() {
  return (
    <Section
      title="Dialog"
      note="A question or a short form that has to be answered before the work goes on. It holds the focus until it closes, and Escape closes it."
    >
      <Part title="Typical">
        <Frame>
          <Dialog>
            <DialogTrigger render={<Button variant="outline" />}>
              Open the rename dialog
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Rename the project</DialogTitle>
                <DialogDescription>
                  The client sees the new name on every request.
                </DialogDescription>
              </DialogHeader>
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="dialog-project">Project name</FieldLabel>
                  <Input id="dialog-project" defaultValue="Billing portal" />
                </Field>
              </FieldGroup>
              <DialogFooter>
                <DialogClose render={<Button variant="ghost" />}>
                  Cancel
                </DialogClose>
                <DialogClose render={<Button />}>Save changes</DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </Frame>
      </Part>

      <Part
        title="Confirming something that cannot be undone"
        note="The title asks the question and names the thing. The button repeats the verb, and the way out is the quiet one."
      >
        <Frame>
          <Dialog>
            <DialogTrigger render={<Button variant="outline" />}>
              Open the delete confirmation
            </DialogTrigger>
            <DialogContent showCloseButton={false}>
              <DialogHeader>
                <DialogTitle>Delete the project Billing portal?</DialogTitle>
                <DialogDescription>
                  Its 7 open requests and the hours logged on them are deleted
                  with it. This cannot be undone.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose render={<Button variant="ghost" />}>
                  Cancel
                </DialogClose>
                <DialogClose render={<Button variant="destructive" />}>
                  Delete project
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long title and a long text wrap. A dialog taller than the screen scrolls inside itself, and its buttons stack on a narrow screen."
      >
        <Frame>
          <Dialog>
            <DialogTrigger render={<Button variant="outline" />}>
              Open a dialog with a long text
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>
                  Enviar a solicitação de alteração para a aprovação do
                  responsável pelo projeto
                </DialogTitle>
                <DialogDescription>
                  O responsável recebe um e-mail com o resumo da solicitação e o
                  prazo estimado. Enquanto ele não responder, a solicitação fica
                  parada e as horas lançadas nela não entram na próxima fatura.
                </DialogDescription>
              </DialogHeader>
              <FieldGroup>
                {[
                  "Assunto",
                  "Prazo estimado",
                  "Responsável",
                  "Observações",
                ].map((label, index) => (
                  <Field key={label}>
                    <FieldLabel htmlFor={`dialog-long-${index}`}>
                      {label}
                    </FieldLabel>
                    <Input id={`dialog-long-${index}`} />
                    <FieldDescription>
                      Aparece no e-mail enviado ao responsável pelo projeto.
                    </FieldDescription>
                  </Field>
                ))}
              </FieldGroup>
              <DialogFooter>
                <DialogClose render={<Button variant="ghost" />}>
                  Cancelar
                </DialogClose>
                <DialogClose render={<Button />}>
                  Enviar para a aprovação
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </Frame>
      </Part>
    </Section>
  )
}

const sides = ["left", "top", "bottom"] as const

export function SheetBlock() {
  return (
    <Section
      title="Sheet"
      note="A panel that slides over the edge of the screen and keeps the page behind it. The sidebar becomes one on a narrow screen."
    >
      <Part title="Typical">
        <Frame>
          <Sheet>
            <SheetTrigger render={<Button variant="outline" />}>
              Open the request sheet
            </SheetTrigger>
            <SheetContent>
              <SheetHeader>
                <SheetTitle>Export leaves out archived projects</SheetTitle>
                <SheetDescription className="font-mono">
                  REQ-0142
                </SheetDescription>
              </SheetHeader>
              <div className="flex flex-col gap-3 px-4">
                <StatusIndicator status="working">In progress</StatusIndicator>
                <p>
                  The client asked for the export to include archived projects
                  and the files attached to each request.
                </p>
              </div>
              <SheetFooter>
                <Button>Open the request</Button>
                <SheetClose render={<Button variant="ghost" />}>
                  Close
                </SheetClose>
              </SheetFooter>
            </SheetContent>
          </Sheet>
        </Frame>
      </Part>

      <Part
        title="Sides"
        note="A sheet comes from the right by default. Navigation comes from the left, and a short choice from the top or the bottom."
      >
        <Frame>
          {sides.map((side) => (
            <Sheet key={side}>
              <SheetTrigger render={<Button variant="outline" />}>
                Open a sheet from the {side}
              </SheetTrigger>
              <SheetContent side={side}>
                <SheetHeader>
                  <SheetTitle>Filters</SheetTitle>
                  <SheetDescription>
                    This sheet came from the {side} of the screen.
                  </SheetDescription>
                </SheetHeader>
                <SheetFooter>
                  <SheetClose render={<Button variant="secondary" />}>
                    Close
                  </SheetClose>
                </SheetFooter>
              </SheetContent>
            </Sheet>
          ))}
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long title wraps beside the close button, and a long text scrolls inside the sheet."
      >
        <Frame>
          <Sheet>
            <SheetTrigger render={<Button variant="outline" />}>
              Open a sheet with a long text
            </SheetTrigger>
            <SheetContent>
              <SheetHeader>
                <SheetTitle>
                  La exportación de las solicitudes no incluye los proyectos
                  archivados
                </SheetTitle>
                <SheetDescription>
                  Solicitud abierta por la persona responsable del proyecto en
                  la empresa del cliente.
                </SheetDescription>
              </SheetHeader>
              <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-4">
                {[1, 2, 3, 4, 5, 6].map((paragraph) => (
                  <p key={paragraph}>
                    El cliente pidió que la exportación incluya los proyectos
                    archivados y los archivos adjuntos de cada solicitud, con el
                    nombre de quien la abrió y la fecha de cada cambio.
                  </p>
                ))}
              </div>
              <SheetFooter>
                <SheetClose render={<Button variant="secondary" />}>
                  Cerrar
                </SheetClose>
              </SheetFooter>
            </SheetContent>
          </Sheet>
        </Frame>
      </Part>
    </Section>
  )
}
