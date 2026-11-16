import { Button } from "@workspace/ui/components/button"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { Skeleton } from "@workspace/ui/components/skeleton"
import { InboxIcon, PlusIcon, SearchXIcon } from "lucide-react"

import { Frame, Part, Section } from "../parts"

export function SkeletonBlock() {
  return (
    <Section
      title="Skeleton"
      note="The shape of content that is still loading. It takes the place and the size of what will arrive, so nothing jumps."
    >
      <Part title="Typical">
        <Frame className="max-w-sm">
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <div className="flex min-w-0 flex-1 flex-col gap-2">
            <Skeleton className="h-4 w-2/5" />
            <Skeleton className="h-3 w-4/5" />
          </div>
        </Frame>
      </Part>

      <Part
        title="A list that is loading"
        note="One skeleton per row, with widths that differ as real text does."
      >
        <Frame className="max-w-md flex-col flex-nowrap items-stretch gap-0 p-0">
          {(["w-3/5", "w-4/5", "w-2/5", "w-3/4"] as const).map((width) => (
            <div
              key={width}
              className="flex h-7 items-center gap-3 border-b border-border-subtle px-3 last:border-b-0"
            >
              <Skeleton className="h-3 w-14 shrink-0" />
              <Skeleton className={`h-3 ${width}`} />
            </div>
          ))}
        </Frame>
      </Part>

      <Part
        title="Edge cases"
        note="A skeleton shrinks with its container and never pushes the layout."
      >
        <Frame className="max-w-40">
          <Skeleton className="h-4 w-80 max-w-full" />
          <Skeleton className="h-3 w-64 max-w-full" />
        </Frame>
      </Part>
    </Section>
  )
}

export function EmptyBlock() {
  return (
    <Section
      title="Empty"
      note="What a list, a table or a pane shows when it has nothing. It says why, and offers the next step when there is one."
    >
      <Part title="Typical">
        <Frame>
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <InboxIcon />
              </EmptyMedia>
              <EmptyTitle>Nothing is waiting for you</EmptyTitle>
              <EmptyDescription>
                A request assigned to you appears here.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button>
                <PlusIcon data-icon="inline-start" />
                New request
              </Button>
            </EmptyContent>
          </Empty>
        </Frame>
      </Part>

      <Part
        title="After a search, and with two actions"
        note="A search that finds nothing says what was searched and offers the way back. One action is primary."
      >
        <div className="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-3">
          <Frame>
            <Empty className="border">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <SearchXIcon />
                </EmptyMedia>
                <EmptyTitle>No request matches</EmptyTitle>
                <EmptyDescription>
                  Nothing was found for{" "}
                  <span className="font-mono">archived export</span>.
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button variant="outline">Clear the search</Button>
              </EmptyContent>
            </Empty>
          </Frame>
          <Frame>
            <Empty className="border">
              <EmptyHeader>
                <EmptyTitle>No project yet</EmptyTitle>
                <EmptyDescription>
                  A client needs a project before a request can be opened.
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent className="flex-row justify-center">
                <Button>New project</Button>
                <Button variant="ghost">Import</Button>
              </EmptyContent>
            </Empty>
          </Frame>
        </div>
      </Part>

      <Part
        title="Edge cases"
        note="A long headline and a long description wrap inside a narrow pane."
      >
        <Frame className="max-w-xs">
          <Empty>
            <EmptyHeader>
              <EmptyTitle>
                Nenhuma solicitação aguarda a sua aprovação
              </EmptyTitle>
              <EmptyDescription>
                Quando um cliente enviar uma solicitação para este projeto, ela
                aparece aqui com o nome de quem a enviou.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button variant="outline" className="max-w-full">
                <span className="truncate">
                  Ver todas as solicitações deste projeto
                </span>
              </Button>
            </EmptyContent>
          </Empty>
        </Frame>
      </Part>
    </Section>
  )
}
