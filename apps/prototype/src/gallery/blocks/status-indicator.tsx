import { StatusIndicator } from "@workspace/ui/components/status-indicator"

import { Frame, Part, Section } from "../parts"

const meanings = [
  {
    status: "working",
    label: "In progress",
    when: "Someone is on it. Nothing is asked of the reader.",
  },
  {
    status: "attention",
    label: "Waiting for the client",
    when: "It is stopped until someone acts.",
  },
  {
    status: "done",
    label: "Delivered",
    when: "It ended as expected.",
  },
  {
    status: "error",
    label: "Failed",
    when: "It ended wrong, or it cannot go on.",
  },
] as const

const requests = [
  {
    id: "REQ-0142",
    title: "Export leaves out archived projects",
    status: "working",
    label: "In progress",
  },
  {
    id: "REQ-0139",
    title: "New logo on the invoice",
    status: "attention",
    label: "Waiting for the client",
  },
  {
    id: "REQ-0131",
    title: "Login by e-mail code",
    status: "done",
    label: "Delivered",
  },
  {
    id: "REQ-0127",
    title: "Nightly import of payments",
    status: "error",
    label: "Failed",
  },
] as const

export function StatusIndicatorBlock() {
  return (
    <Section
      title="Status indicator"
      note="The status of a record: a sign and a word in one of the four status colours. The sign belongs to the meaning, and the word comes from the screen."
    >
      <Part title="Typical">
        <Frame className="max-w-md justify-between">
          <span className="min-w-0 truncate">
            Export leaves out archived projects
          </span>
          <StatusIndicator status="working">In progress</StatusIndicator>
        </Frame>
      </Part>

      <Part
        title="The four meanings"
        note="A colour has one meaning everywhere. Which status of a record shows which meaning is decided with each screen."
      >
        <div className="grid grid-cols-[repeat(auto-fill,minmax(13rem,1fr))] gap-3">
          {meanings.map(({ status, label, when }) => (
            <Frame key={status} className="flex-col items-start gap-2">
              <StatusIndicator status={status}>{label}</StatusIndicator>
              <span className="type-code text-muted-foreground">{status}</span>
              <span className="type-caption text-muted-foreground">{when}</span>
            </Frame>
          ))}
        </div>
      </Part>

      <Part
        title="Variants"
        note="Plain in a row or a table, where the line already separates. A badge beside a title, where the status stands alone."
      >
        <Frame className="flex-col items-start">
          <div className="flex flex-wrap items-center gap-4">
            {meanings.map(({ status, label }) => (
              <StatusIndicator key={status} status={status}>
                {label}
              </StatusIndicator>
            ))}
          </div>
          <div className="flex flex-wrap items-center gap-3">
            {meanings.map(({ status, label }) => (
              <StatusIndicator key={status} status={status} variant="badge">
                {label}
              </StatusIndicator>
            ))}
          </div>
        </Frame>
      </Part>

      <Part
        title="In a list"
        note="Each row carries its status at the end, so a column of them can be read down."
      >
        <Frame className="max-w-xl flex-col flex-nowrap items-stretch gap-0 p-0">
          {requests.map(({ id, title, status, label }) => (
            <div
              key={id}
              className="flex min-h-7 items-center gap-3 border-b border-border-subtle px-3 py-1 last:border-b-0"
            >
              <span className="shrink-0 type-code text-muted-foreground">
                {id}
              </span>
              <span className="min-w-0 flex-1 truncate">{title}</span>
              <StatusIndicator status={status} className="shrink">
                {label}
              </StatusIndicator>
            </div>
          ))}
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A long word is cut with an ellipsis. The sign is never cut."
      >
        <Frame className="max-w-xs flex-col items-start">
          <StatusIndicator status="attention" className="shrink">
            Aguardando a aprovação do responsável pelo projeto no cliente
          </StatusIndicator>
          <StatusIndicator status="error" variant="badge" className="shrink">
            No se pudo entregar la solicitud al equipo de facturación
          </StatusIndicator>
        </Frame>
      </Part>
    </Section>
  )
}
