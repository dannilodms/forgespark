# 01 - Frontend workspace and first component

Status: ready-for-agent

Read `../spec.md` and `DESIGN.md` first.

## What to build

The thinnest path from an empty repository to a themed component on screen.

- A pnpm workspace at the repository root, managed with Vite+, containing `packages/ui` and `apps/prototype`.
- `packages/ui`: the shadcn/ui monorepo layout with Tailwind CSS v4, a stylesheet defining the style variables for light and dark, and one component, Button. The style variables carry the names and roles listed under "Style variables" in `DESIGN.md`, including the ones it adds to shadcn/ui's.
- `apps/prototype`: a React single-page app with TanStack Router. Its home page renders the Button from `packages/ui`.
- The prototype bar: a small floating bar shown over every route of the prototype, outside the page it floats over. It holds the light/dark toggle. Later issues add the language and the scenarios to it.
- A root `Taskfile.yml` with tasks to start the prototype and to run the checks.
- Update the "Local development" section of `README.md` and `README.pt-BR.md` with the commands.

The initial style values are placeholders; issue 03 sets the values of `DESIGN.md`.

The folders `packages/ui` and `apps/prototype` already exist with a `.gitkeep` placeholder each. Remove the placeholder from a folder when you put real files in it.

## Acceptance criteria

- [x] One command from the repository root installs dependencies; one command starts the prototype.
- [x] The home page shows a Button imported from `packages/ui`, not defined in the app.
- [x] The toggle in the prototype bar switches between light and dark, and the choice survives a reload. With no stored choice, the system preference applies.
- [x] The prototype bar is on every route, can be reached by keyboard, and covers no control of the page at desktop and 375px widths.
- [x] Changing a style variable in `packages/ui` changes the Button in the running prototype without a restart.
- [x] Every style variable named in `DESIGN.md` exists in the stylesheet, in light and dark.
- [x] Format, lint and type checks pass through the task runner.
- [x] Both READMEs document the commands.
