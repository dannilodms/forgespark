# 01 - Frontend workspace and first component

Status: ready-for-agent

Read `../spec.md` and `DESIGN.md` first.

## What to build

The thinnest path from an empty repository to a themed component on screen.

- A pnpm workspace at the repository root, managed with Vite+, containing `packages/ui` and `apps/prototype`.
- `packages/ui`: the shadcn/ui monorepo layout with Tailwind CSS v4, a stylesheet defining the style variables for light and dark, and one component, Button. The style variables carry the names and roles listed under "Style variables" in `DESIGN.md`, including the ones it adds to shadcn/ui's.
- `apps/prototype`: a React single-page app with TanStack Router. Its home page renders the Button from `packages/ui` and has a light/dark toggle.
- A root `Taskfile.yml` with tasks to start the prototype and to run the checks.
- Update the "Local development" section of `README.md` and `README.pt-BR.md` with the commands.

The initial style values are placeholders; the style variations come in issues 05 to 07.

The folders `packages/ui` and `apps/prototype` already exist with a `.gitkeep` placeholder each. Remove the placeholder from a folder when you put real files in it.

## Acceptance criteria

- [ ] One command from the repository root installs dependencies; one command starts the prototype.
- [ ] The home page shows a Button imported from `packages/ui`, not defined in the app.
- [ ] The toggle switches between light and dark, and the choice survives a reload. With no stored choice, the system preference applies.
- [ ] Changing a style variable in `packages/ui` changes the Button in the running prototype without a restart.
- [ ] Every style variable named in `DESIGN.md` exists in the stylesheet, in light and dark.
- [ ] Format, lint and type checks pass through the task runner.
- [ ] Both READMEs document the commands.
