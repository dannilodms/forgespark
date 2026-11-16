import { Button } from "@workspace/ui/components/button"
import { Toaster, toast } from "@workspace/ui/components/toast"

import { Frame, Part, Section } from "../parts"

const kinds = [
  {
    label: "Show a plain toast",
    when: "Something happened and nothing has to be done.",
    show: () =>
      toast.add({
        title: "Draft saved",
        description: "It stays in this browser until the request is sent.",
      }),
  },
  {
    label: "Show a success toast",
    when: "What the person asked for is done.",
    show: () =>
      toast.add({
        type: "success",
        title: "Request created",
        description: "REQ-0143 is in the queue of the Billing portal.",
      }),
  },
  {
    label: "Show a warning toast",
    when: "It worked, and something needs attention.",
    show: () =>
      toast.add({
        type: "warning",
        title: "Saved without a client",
        description: "Choose a client before the request can be billed.",
      }),
  },
  {
    label: "Show an error toast",
    when: "It did not work. The toast says what to do next.",
    show: () =>
      toast.add({
        type: "error",
        priority: "high",
        title: "The export failed",
        description: "Nothing was changed. Try again in a moment.",
      }),
  },
  {
    label: "Show an information toast",
    when: "A fact the person may want to know.",
    show: () =>
      toast.add({
        type: "info",
        title: "A new version is ready",
        description: "It loads the next time this page is opened.",
      }),
  },
] as const

function exportProject() {
  void toast.promise(
    new Promise<void>((resolve) => {
      window.setTimeout(resolve, 2000)
    }),
    {
      loading: {
        title: "Exporting the project",
        description: "This takes a few seconds.",
      },
      success: {
        title: "Export ready",
        description: "billing-portal.zip was downloaded.",
      },
      error: {
        title: "The export failed",
        description: "Nothing was changed. Try again in a moment.",
      },
    }
  )
}

function archiveRequest() {
  toast.add({
    title: "Request archived",
    description: "REQ-0142 left the queue.",
    actionProps: {
      children: "Undo",
      onClick: () => toast.add({ type: "success", title: "Request restored" }),
    },
  })
}

function showLongToast() {
  toast.add({
    type: "warning",
    title:
      "A solicitação foi salva, mas o cliente ainda não foi avisado por e-mail",
    description:
      "O responsável pelo projeto não tem um e-mail cadastrado. Cadastre o e-mail e envie o aviso outra vez na página da solicitação.",
    actionProps: { children: "Abrir a solicitação" },
  })
}

export function ToastBlock() {
  return (
    <Section
      title="Toast"
      note="A short message about something that just happened. It leaves by itself, so it never holds the only copy of an error the person has to fix."
    >
      {/* The viewport sits above the prototype bar, which takes the bottom of the screen. */}
      <Toaster className="bottom-16" />

      <Part title="Typical">
        <Frame>
          <Button variant="secondary" onClick={kinds[1].show}>
            {kinds[1].label}
          </Button>
        </Frame>
      </Part>

      <Part
        title="Kinds"
        note="Each kind has its own sign. A status colour never appears without it."
      >
        <div className="grid grid-cols-[repeat(auto-fill,minmax(15rem,1fr))] gap-3">
          {kinds.map(({ label, when, show }) => (
            <Frame key={label} className="flex-col items-start gap-2">
              <Button variant="outline" onClick={show}>
                {label}
              </Button>
              <span className="type-caption text-muted-foreground">{when}</span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part
        title="Loading and an action"
        note="A toast follows work in progress and changes when it ends. A toast can carry one action, such as undoing what was just done."
      >
        <Frame>
          <Button variant="outline" onClick={exportProject}>
            Show a loading toast
          </Button>
          <Button variant="outline" onClick={archiveRequest}>
            Show a toast with an action
          </Button>
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long title and a long description wrap inside the toast, and several toasts stack."
      >
        <Frame>
          <Button variant="outline" onClick={showLongToast}>
            Show a toast with a long text
          </Button>
        </Frame>
      </Part>
    </Section>
  )
}
