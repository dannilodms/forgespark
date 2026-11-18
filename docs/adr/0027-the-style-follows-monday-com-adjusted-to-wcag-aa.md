---
status: accepted
---

# The style follows monday.com, adjusted to WCAG AA

ADR-0025 kept the look of Zed, a code editor. On 2026-11-18 the maintainer found that it read as a tool for programmers, while the product is the operational hub of a consultancy: clients, projects, tasks and the requests of clients, closer to a CRM than to an editor. Twelve styles were drawn on the same screen of requests, in both themes, from the values each product publishes, and the maintainer chose monday.com. A clickable prototype of the screens of the CRM in that style (home, clients and their knowledge base, projects with planned and logged hours, requests, settings) was then approved. It was made outside the repository and is a reference, not code: each screen is still built in `apps/web` with its feature (ADR-0025).

- **What comes from monday.com.** The values of its design system, Vibe: the blue-grey frame that holds the header and the sidebar, the white work surface inset in it, the blue `#0073ea` of the primary action, Figtree for text and Poppins for headings, statuses as filled labels in vivid colours, rows in coloured groups with sums at their foot, 4px controls, 8px cards and floating surfaces, 16px dialogs, and shadows only on what floats.
- **Adjusted to WCAG AA.** Every pair is measured. A value of Vibe that fails is changed by the smallest step that passes, and `DESIGN.md` gives both. A label takes the dark ink or white, whichever reaches 4.5:1 on its fill.
- **What does not change.** shadcn/ui on Base UI and Tailwind CSS, light and dark themes, desktop first, the 16px text the maintainer chose before (PR #7), the 3px focus outline and the brand mark. The components stay; the style is their variables and variants.
- **Colour.** The old rule that colour only means status gives way to three uses: blue for action, the label palette for the values of statuses and categories, and group colours. Four feedback colours remain for alerts, toasts and problems in a field.
- **Mono.** Vibe defines none. JetBrains Mono is used only for what is copied or compared character by character: codes, identifiers, secrets, addresses.

## Considered Options

The twelve styles, with their evidence, are in section 5 of `docs/design/references.md`.

- **Keeping Zed.** Ready, documented and measured, but it reads as an editor, and its close greys separate areas poorly.
- **Linear.** Calm and dense, and its subject, the work of a software team, is close. Its grey text and the purple of its light theme fail AA, and it is widely copied.
- **Attio and Twenty.** CRMs of records, neutral and precise; Twenty is open source under the same licence. Little character of their own.
- **HubSpot and Pipedrive.** CRMs for sales, familiar to small companies. HubSpot has no usable dark theme; Pipedrive's green action blurs with the green of done.
- **Salesforce, Dynamics 365 and IBM Carbon.** Corporate systems with the most documentation, heavier than a small consultancy needs, and each would change the shape of most components (pills and 20px cards, shadows everywhere, square corners).
- **Vercel Geist and Stripe.** Polished and mostly monochrome; little colour to tell statuses apart at a glance.

## Consequences

- `DESIGN.md` is rewritten: rules, style variables and their values, type roles, the label palette, group colours and the recipes of the CRM (app shell, board table, record page, side panel, form page). The framed pane of the setup and the login screen takes the new look.
- IBM Plex Sans, IBM Plex Serif and Lilex leave the catalog; Figtree, Poppins and JetBrains Mono enter it.
- The grid texture, the diamond nodes, the title and status bars and the inset edge of buttons, all from Zed, are gone, with their style variables.
- The label of a record's status, the board table and the app shell have no code until the first screen that needs them.
- The files of the brand keep the blue and the typeface of the previous style until they are redrawn, in a change of their own.
