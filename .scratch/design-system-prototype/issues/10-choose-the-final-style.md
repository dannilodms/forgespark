# 10 - Choose the final style

Status: ready-for-human
Blocked by: 06, 07, 08, 09

Read `../spec.md` and `DESIGN.md` first.

## What to do

This is the maintainer's decision; an agent cannot make it.

Open the prototype and compare the three style variations (`zed`, `figcraft`, `herdr`) on the Compare route, in the rest of the gallery and in the sample work screen, in light and dark, and in at least two languages. For each trait below, pick the style variation to take it from, or describe something different:

- Surfaces and neutrals, light theme
- Surfaces and neutrals, dark theme
- Accent colour and primary button
- Corner radius
- Display typeface
- Interface typeface
- Mono typeface
- Small labels: sentence case or uppercase mono
- Panes: flush or separate cards
- Depth: edges and shadows
- Page decoration: grid, ticks, none
- Density: control and row heights

The references leave three of these genuinely open, with no way to derive them: the pane model, a coloured versus neutral primary button, and whether to use a characterful display typeface at all.

The Foundations route marks every colour pair that fails 4.5:1. A style variation may fail; the final style may not. Where a chosen trait fails, say how to correct it, or leave the correction to the follow-up issue.

## Outcome

Record the choices under an `## Answer` heading in this file. A follow-up issue then:

- builds the final style from them, meeting WCAG AA as `DESIGN.md` requires;
- removes the three style variations, the style variation selector and the Compare route;
- rewrites "Style variations" in `DESIGN.md` as the values of the final style, and removes what the document says only because style variations exist.

The gallery stays.
