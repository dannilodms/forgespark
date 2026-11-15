# Design references: Zed, FigCraft, Herdr

Research date: 2026-11-15. Sources fetched over plain HTTP (HTML + linked stylesheets) and through a remote headless browser (Firecrawl, 1920 px wide viewport) for screenshots and computed styles.

## How to read the evidence tags

Every factual item carries one of these tags:

- **[CSS]** read directly from the site's HTML/CSS source: custom properties, `@font-face`, class strings, rule bodies.
- **[CSS-computed]** `getComputedStyle()` values read from the live page rendered in the remote browser. Same reliability class as [CSS]; numbers are the rendered result at a 1920 px viewport.
- **[SHOT]** observed in a screenshot taken during the research. The screenshots are captures of third-party sites and were not kept in this repository. Hex values tagged [SHOT-px] are pixel samples from those images (JPEG/WebP, so +/- a few units).
- **[INF]** inferred, or from prior knowledge, not confirmed on the pages fetched.

oklch values in brackets are my own conversions of the hex values (rounded), provided for Tailwind/shadcn use; the sites themselves mostly declare hex, `lab()` or `color-mix()`.

---

## 1. Zed (https://zed.dev)

### What it is, and what the page shows

- A native, GPU-rendered code editor written in Rust, with agentic AI and collaboration features. [CSS] (meta description)
- The home page shows the application in two ways:
  - A **live HTML/CSS recreation of the Zed window** in the hero (`.theme-preview`), themed by ~250 CSS variables that follow Zed's theme schema (`--editor-background`, `--panel-background`, `--tab-active_background`, `--syntax-*`, `--terminal-ansi-*`, `--players-*`). Both a light and a dark set are inlined in the page. [CSS]
  - **Real screenshots/video posters of the native app** (dark theme) in the feature tiles. [SHOT]
- Secondary page examined: `/extensions` (search input, filter chips, card grid). [SHOT][CSS]

### 1A. Zed website

**Typography**

- Four self-hosted families, declared via `@font-face` [CSS]:
  - `writer` = **iA Writer Quattro S** (400/700 + italics): the **body font** for the whole site (`body` carries its class). It is a quasi-proportional "duospace" face, so body copy reads like a refined typewriter.
  - `plexSerif` = **IBM Plex Serif** variable (300-700, roman + italic): all headings.
  - `zedMono` = **Lilex** Regular: code in the app mock.
  - `plexSans` = **IBM Plex Sans** Regular: preloaded and assigned to the app mock (`font-plexSans`), but the computed font of the mock's UI labels is still `writer` (see 1B).
  - `kbd` hints use the system mono stack (`ui-monospace`).
- Heading scale (custom classes) [CSS]:
  - `.h0` `clamp(2.6rem, 2rem + 2.5vw, 3rem)`, weight `var(--h0-weight)` = **340 light / 320 dark**, letter-spacing -0.02em, line-height 1.2. Rendered: 48px / 57.6px, **italic**, accent blue. [CSS-computed]
  - `.h1` up to 2rem, weight 390/320; `.h2` up to 1.6rem (rendered 25.6px / 32px, weight 390 light, 360 dark); `.h3` up to 1.35rem; `.h4` ~1.15rem; `.h5` ~1.05rem (rendered 16.8px, line-height 1.1); `.h6` ~0.95rem.
  - Headings are **light-weight serif** (variable weights 320-410), never bold. Weights are tuned per theme (slightly lighter on dark).
- Body: 16px / 24px, weight 400. Hero lead 16px with letter-spacing -0.4px. Feature copy 13px / 19.5px. [CSS-computed]
- Nav links and buttons: 14px / 20px, weight 400, `tracking-tight` (-0.025em = -0.35px). [CSS-computed]
- Small text is everywhere: class tally on the home page: `text-sm` x155, `text-[11px]` x119, `text-[0.9375rem]` x68, `text-[.8125rem]` x64+35, `text-xs` x62, `text-[10px]` x50. Weights used: `font-[450]` x27, bold x8, medium x5. [CSS]
- Habits: serif for display, quasi-mono for everything else, real mono for `kbd` and code. No uppercase labels, no wide tracking.

**Colour**

Theme mechanism: `html.light` / `html.dark` class (not a media query in CSS). [CSS]

| Role | Light | Dark | Tag |
|---|---|---|---|
| Page background | `bg-cream-100/20` = cream #d1cfc8 at 20% over white, about **#f6f5f4** [oklch(0.971 0.002 68)]; sampled #f3f3f1 | `hsl(218,13%,7.5%)` = **#111216** [oklch(0.183 0.008 274)] | [CSS][SHOT-px] |
| Nav background | `--color-cream-50` **#f5f4f3** (`--nav-bg-color: #f7f7f2`) | **#121316** (`--nav-bg-color: #121417`) | [CSS] |
| Popover background | #ffffff | #111216 | [CSS] |
| Hairline border | **#dadde2** [oklch(0.897 0.008 261)] | offgray-600 #626976 at 14% alpha | [CSS] |
| Body text | offgray-800 **#474c55** [oklch(0.415 0.016 262)] | offgray-300 **#abafb8** [oklch(0.754 0.014 267)] | [CSS][CSS-computed] |
| High-contrast text | #000000 | #ffffff | [CSS] |
| Accent | `--color-accent-blue` **#1348dc** [oklch(0.479 0.230 264)] | text accent switches to blue-300 **#90c5ff**; filled buttons stay #1348dc | [CSS][CSS-computed] |
| Focus ring | `outline: 3px solid #094ece80`, offset 2px | same | [CSS] |

- Two custom neutral ramps built with `color-mix(in oklch, base, white/black N%)`: **cream** (base #7f7866, warm) and **offgray** (base #727a89, cool blue-gray; steps 50 #e3e4e7, 100 #ced1d6, 200 #c0c3ca, 300 #abafb8, 400 #878e9b, 600 #626976, 700 #535964, 800 #474c55, 900 #35383f, 950 #212328, 1000 #0d0d0f). Light mode mixes warm paper with cool gray text; dark mode is cool only. [CSS]
- Colour quantity: very low. One blue for the headline, primary button, links and hover borders; everything else is neutral. Decorative motif variables exist (`--motif-subtle-1: #ff5500a6`, `--motif-accent-1: #1e69f6`, `--dithering-front: #0751cf`). [CSS]

**Shape and depth**

- Radius: `rounded-sm` x133, `rounded` x116 (both 4px in this Tailwind 4 build), `rounded-xs` x79 (2px), `rounded-full` x22, `rounded-md`/`lg` x2 each. Buttons 4px, `kbd` 2px, app window mock 6px. [CSS][CSS-computed]
- Borders: 1px hairlines everywhere; sections are separated by full-bleed 1px lines rather than cards. Some image frames use `border border-dashed` with a 6px offset. [CSS][SHOT]
- Signature shadow 1, **inset "keycap" bottom edge**: `box-shadow: hsl(218 13% 50% / 0.1) 0 -2px 0 0 inset` on secondary buttons and cards; primary button `hsl(219 93% 30%) 0 -2px 0 0 inset, hsl(219 93% 95%) 0 1px 3px 0` (computed `rgb(5,55,148)` inset). Removed on hover. [CSS][CSS-computed]
- Signature shadow 2, **hard offset blocks** on hover: `--sh-default: 6px 6px 0 #074dcf0f, -6px -6px 0 #074dcf0f` (dark: 5px, `#3d7df50d`); cards use `--sh-alt` plus `hover:border-blue-300`. No blur. [CSS]
- Conventional soft shadow only on the floating app window: ring 1px + `shadow-xl` tinted offgray-500/20. [CSS-computed]
- Texture and blueprint motifs [CSS][SHOT]:
  - SVG **grid pattern** (10px cells, `data-texture="grid"`, accent blue at 15% / blue-300 at 5%).
  - **Diamond "nodes"**: 6px squares rotated 45 degrees, 1px border, placed where section lines intersect.
  - Vertical rules drawn as broken segments; a `<canvas>` shader behind the hero at `--hero-shader-opacity: .1` (light) / `.15` (dark) giving a faint dithered blue haze and a large line-art logo.
  - Very subtle gradients only (`.square-gradient-bg`: blue-200 5% to blue-50 10%).

**Density and spacing**

- Tailwind base unit 0.25rem. Content column `max-width: 1080px` (min 1120px at the widest breakpoint); header 57px tall. [CSS]
- Most-used gaps: `gap-1` x293, `gap-2` x159, `gap-1.5` x97, `gap-0.5` x76. Paddings: `p-1.5`, `p-0.5`, `px-4`, `px-12`, `py-6`, `p-4`, `p-2.5`, `p-3.5`. [CSS]
- Control heights: nav buttons 32px (`h-8`, padding 0 10px), hero CTAs 36px (`h-9`), search input 36px, filter chips 24px (`h-6`), `kbd` 20px. [CSS][CSS-computed]
- Marketing sections are moderately airy, but text and controls are small; overall it reads compact for a marketing site. [SHOT]

**Components visible**

- **Primary button**: #1348dc fill, white 14px text, 4px radius, inset darker bottom edge, leading 16px icon, trailing `kbd` shortcut chip (white 20% border, white 10% fill). Active state `translate-y-px scale-[.99]`. [CSS][SHOT]
- **Secondary button**: white 50% fill, 1px offgray-200/50 border, inset gray bottom edge, same size, trailing `kbd`. **Ghost button**: transparent border, hover `bg-offgray-50/80`. [CSS]
- **kbd chip**: 20px high, 2px radius, mono 11px, 1px gray-500/20 border, gray-50/50 fill. Shortcut hints appear inside almost every button and next to the search icon ("Ctrl + Shift + P"). [CSS][SHOT]
- **Search input** (`/extensions`): 36px, 4px radius, 1px gray-300 border, white fill, left icon (pl-9), italic placeholder, `shadow-xs`; focus shows a blue ring. [CSS][SHOT]
- **Filter chips / tabs**: 24px, mono 12px, transparent until selected; selected = blue-50/60 fill, blue-500/20 border, blue-700 text. [CSS][SHOT]
- **Cards** (extension tiles): 4px radius, 1px offgray-100 border, white/60 fill, padding 10px, inset bottom edge; three lines (title 15px serif, description 13px, author with 12px icon), download count right-aligned in mono. Hover: blue-300 border + offset block shadow. [CSS][SHOT]
- **Announcement bar**: full-width strip under the nav, tinted blue-gray, "News:" in blue. [SHOT]
- **Stats row**: large mono numerals (11,703 / 95,041) over small labels, separated by hairlines. [SHOT]
- Dialog/popover primitives appear to be Base UI (`.base-ui-disable-scrollbar` in the page). [CSS]

**Motion**

- Default transition 150ms `cubic-bezier(.4,0,.2,1)`; many controls use `duration-70` (70ms). [CSS]
- Hero entrance `hero-rise`: 1.2s `cubic-bezier(.16,1,.3,1)`, translateY(14px) scale(.99) to rest. Dialogs: `dialogContentShow` scale .96 to 1; popovers `scaleIn` from scale(.9) with rotateX(-10deg). Slow decorative rotations (32s, 50s), blinking caret (1s step-end), marquee. [CSS]
- Not observed in motion; values are from CSS only.

**Iconography**

- Site: **Lucide** (class names `lucide lucide-*`, 24 viewBox, stroke 2) rendered at 10-12px, plus custom 16x16 SVGs. [CSS]

### 1B. Zed application UI

Two sources: the HTML mock on the home page (values are exact) and real app posters (visual only).

**Theme tokens inlined in the page** [CSS] (the token names match Zed's theme schema; the values correspond to Zed's default "One Light" / "One Dark" themes [INF]):

| Token | Light | Dark |
|---|---|---|
| `--background` (title bar, status bar, window) | #dcdcdd | #3b414d |
| `--surface-background` / `--panel-background` / `--tab_bar-background` | #ebebec | #2f343e |
| `--editor-background` / `--tab-active_background` / `--toolbar-background` | #fafafa | #282c33 |
| `--border` | #c9c9ca | #464b57 |
| `--border-variant` | #dfdfe0 | #363c46 |
| `--border-focused` / `--pane-focused_border` | #7d82e8 | #47679e |
| `--border-selected` | #cbcdf6 | #293b5b |
| `--text` | #242529 | #dce0e5 |
| `--text-muted` | #58585a | #a9afbc |
| `--text-placeholder` / disabled | #7e8086 | #878a98 |
| `--text-accent` / `--icon-accent` | #5c78e2 | #74ade8 |
| `--element-background` | #ebebec | #2e343e |
| `--element-hover` / `--ghost_element-hover` | #dfdfe0 | #363c46 |
| `--element-active` / selected | #cacaca | #454a56 |
| `--editor-active_line-background` | #ebebecbf | #2f343ebf |
| `--editor-line_number` / active | #b4b4bb / #44454b | #4e5a5f / #d0d4da |
| `--search-match_background` | #5c79e266 | #74ade866 |
| success / warning / error / info | #669f59 / #a48819 / #d36151 / #5c78e2 | #a1c181 / #dec184 / #d07277 / #74ade8 |
| status backgrounds | solid pale tints (#dfeadb, #faf2e6, #fbdfd9, #e2e2fa) | the status colour at 10% alpha (`...1a`) |
| version control added / deleted / modified | #27a657 / #e06c76 / #d3b020 | same |
| syntax: keyword / string / function / type / number / property / comment | #a449ab / #649f57 / #5b79e3 / #3882b7 / #ad6e25 / #d3604f / #a2a3a7 | #b477cf / #a1c181 / #73ade9 / #6eb4bf / #bf956a / #d07277 / #5d636f |
| collaborator ("players") colours | #5c78e2, #984ea5, #ad6e26, #a349ab, #3a82b7, #d36151, #a48819, #669f59 | #74ade8, #be5046, #bf956a, #b477cf, #6eb4bf, #d07277, #dec184, #a1c181 |

Pixel samples of the mock agree (dark: sidebar #30343f, editor #292c33, status bar #3b414d; light: sidebar #ebebeb, editor #f9fbfa, status bar #dcdcdc). [SHOT-px]

**Structure and metrics of the mock** [CSS-computed]

- Three-level surface model: window chrome (darkest in dark mode, mid-gray in light) > panels > editor (lightest in light mode, darkest in dark). Surfaces are separated by **1px borders only**; no shadows inside the window.
- Title bar 27px, status bar 27px, each with a 1px border on the inner edge. Left panel 256px wide with a 1px right border.
- Text: UI labels 12px / 16px, secondary 11px / ~15px, tertiary 10px / 15px; code 12px / 20px in Lilex. All weight 400 (one 500). In this web mock the UI labels render in iA Writer Quattro, not in the app's own UI font.
- Radii inside the window: 4px (78 elements) and full circles (7); nothing larger.
- Icon buttons 20px (`size-5`, 1px padding) and 16px, transparent until hover (`--ghost_element-hover`); icons 12px (some 10px).
- Icons in the mock: Lucide at 10-12px plus custom 16x16 SVGs. [CSS]

**Components seen** [SHOT]

- **Sidebar list**: project group headers (11px muted, full-width divider), rows of two lines (12px title + 10px muted meta with branch icon, +/- diff counts in green/red, relative time), selected row = flat gray fill, leading status glyph (blue dot, spinner, check box).
- **Tab bar**: flat tabs, each with a 1px right divider; active tab takes the editor background and loses its bottom border; inactive tabs muted text. Back/forward arrows on the left, add/split/expand icons on the right.
- **Breadcrumb / toolbar row** under the tabs: mono path on the left, four 12px icons on the right.
- **Editor**: gutter numbers muted with the active line number bright, active line a faint band, code-lens lines ("2 references") in small muted type, inline diagnostics as tinted rounded chips at line end (error = pale red fill, hint = pale violet), squiggly underlines.
- **Agent panel**: user message in a bordered 4px-radius box on the editor colour; tool calls as single muted rows with a small icon; message composer at the bottom with ghost icon buttons and small dropdowns ("Write", model name).
- **Terminal panel** with its own tab strip; **status bar** with 12px icon buttons, diagnostics count (red circle "1", yellow triangle "5"), language name.
- macOS traffic lights top left; a search field ("Search...") in the title bar of the left column.
- Real app posters (dark) confirm the same layout: proportional sans UI text, mono code, flat tabs, 1px dividers, ~10px window corner radius, no visible shadows inside. [SHOT]

**Not determined for the real app**: exact font families and sizes (the app's defaults are, to my knowledge, "Zed Plex Sans"/IBM Plex Sans for UI and "Zed Mono"/Lilex for buffers [INF]), exact spacing, its native icon set, and a real light-theme screenshot (home page posters are all dark).

### Overall feel

- Adjectives: **precise, editorial, quiet, dense, engineered**.
- Distinguishing traits versus generic SaaS:
  1. Light-weight **italic serif headline + typewriter-like body font** instead of a bold geometric sans.
  2. **Drafting-table decoration**: hairline section rules with diamond nodes, 10px grid texture, dashed frames, hard offset shadows, and inset "keycap" button edges; no soft cards or big gradients.
  3. **Keyboard-first cues**: a `kbd` shortcut inside nearly every button, and an app UI at 10-12px with 27px bars.

---

## 2. FigCraft (https://getartcraft.com/apps/figcraft)

### What it is, and what the page shows

- An open-source, native (Rust) interface design and prototyping app, a Figma-style tool with frames, auto layout, components, variables and live collaboration; one of ~17 "Craft" apps by ArtCraft. Early alpha. [CSS] (meta, page text)
- The page is ArtCraft's marketing template, and it shows the **actual application**: a hero screenshot (dark), six further app screenshots (dark and light) and a collaboration video. [SHOT]
- So there are two distinct design languages here: the **ArtCraft website** (blueprint/HUD style) and the **FigCraft app** (a close Figma look-alike).

### 2A. ArtCraft website

**Typography** [CSS][CSS-computed]

- Fonts (next/font, self-hosted): **Archivo** (variable, weight 100-900, width 62-125%) for display; **Inter** (variable) for body; **Geist Mono** (variable) for labels, buttons and code; **Instrument Serif** (400 + italic) loaded but only used for article blockquotes.
- `.font-display`: Archivo at **weight 620, `font-stretch: 118%`** (an expanded, heavy grotesque). H1 72px / 73.44px (line-height 1.02), letter-spacing **-0.035em** (-2.52px). H2 60px / 61.2px, -0.035em (one at 0.98 / -0.04em). Section H3 30-36px / 1.05, -0.03em. Card H3 24px / 32px, -0.02em.
- Body: Inter 16px / 24px; lead paragraph 18px / 29.25px (`leading-relaxed` 1.625); card copy 16px / 26px.
- `.hud-label`: Geist Mono **11px (0.6875rem), weight 500, uppercase, letter-spacing 0.15em**; used for nav links, section numbers ("01 / FIGCRAFT"), captions, metadata. This is the dominant secondary voice (the class appears 247 times in the page HTML; 60 of the 98 `<p>` elements are hud-labels). Instrument Serif is not used anywhere on this page (0 `font-serif` classes).
- Buttons: Geist Mono 11px, **weight 700, uppercase, letter-spacing 0.12em**.
- Code block: Geist Mono 13px / 21px.
- Habits: sentence-case display with one accent-coloured word per headline; everything small is uppercase mono with wide tracking; numerals as two-digit indices (01, 02...).

**Colour** [CSS]

Theme mechanism: `:root[data-theme=dark]` plus `@media (prefers-color-scheme: dark) :root:not([data-theme=light])`; a moon toggle in the nav; theme change animates the custom properties over 0.4s (`html.theme-anim`).

| Token | Light | Dark |
|---|---|---|
| `--bg` | **#f2f1ee** [oklch(0.958 0.004 91)] | **#121316** [oklch(0.187 0.006 271)] |
| `--bg-raised` | #faf9f7 | #101014 |
| `--bg-sunken` | #e9e8e4 | #060607 |
| `--ink` | #101014 | #f2f1ee |
| `--ink-strong` | #000000 | #ffffff |
| `--muted` | #101014 at 62% (`9e`) | #f2f1ee at 60% |
| `--faint` | #101014 at 42% (`6b`) | #f2f1ee at 38% |
| `--line` | #101014 at 16% (`29`) | #f2f1ee at 15% |
| `--line-strong` | #101014 at 40% | #f2f1ee at 40% |
| `--accent` | #2d81ff | #2d81ff |
| `--accent-ink` | #1659c4 | #74aaff |
| `--invert-bg` / `--invert-fg` | #101014 / #f2f1ee | #ffffff / #0b0b0c |
| `--danger` | #b3261e | #ff6259 |
| `--scanline` | #101014 at 3% | #ffffff at 2.4% |

- Per-product accent: each app page sets `--app` / `--app-ink`. FigCraft: `--app: #02da96`, `--app-ink: #017a54` (light) / `#3fe6ae` (dark). Other apps use blue, coral, violet, amber, etc. Pricing tiers add `--plan` hues (#00a873, #9d4cff, #d97700, #3568c9).
- Neutrals are **one ink colour at different alphas** over a warm off-white (light) or cool near-black (dark); there is no gray ramp.
- Colour quantity: monochrome plus one accent word per heading, a small coloured index chip ("09" on the app green), and check icons. Primary buttons are **inverted ink** (black on light, white on dark), not accent-coloured. The app screenshots supply most of the colour. [SHOT]
- `::selection` = accent fill with white text; focus = `outline: 2px solid var(--accent-ink)`, offset 2px.

**Shape and depth** [CSS][SHOT]

- **Border radius is 0 everywhere** (`rounded-none`; the only other value is the pill constant for a few round items). App icons keep their own rounded squares.
- Borders: 1px in `--line`; stronger outlines in `--line-strong`. Grids are built as `gap-px` over a line-coloured background, so cells share hairlines. No box shadows at all (only `drop-shadow` on app icons, and an inset 2px ring on the highlighted pricing card).
- Blueprint motifs: the content column is bounded by full-height vertical lines; **11px crosshair "ticks"** (`.tick`, 1px `--line-strong`) sit at section corners; each section opens with a 43px sticky header strip (index / name left, descriptor right, translucent `bg/60` + `backdrop-blur-md`); on wide viewports a right-hand **ruler** with tick marks, a scroll percentage and rotating section names in Archivo capitals (`html[data-ruler=on]`).
- Gradients: only on pricing cards (tier colour at 14% fading to transparent). Sticky nav is `bg/95` with `backdrop-blur-sm`.

**Density and spacing** [CSS][CSS-computed]

- Base unit 0.25rem. Top nav 48px (+1px border); nav items are 48px-tall cells separated by vertical hairlines. Content column 1278px at a 1920 viewport.
- Card padding 32px (`p-8`), section side padding 40px (`px-10`), list rows `py-3`. Most-used: `px-6` x118, `gap-4` x98, `py-3` x78, `px-10` x38, `p-6` x36, `p-8` x24.
- Control heights: hero buttons **48px** (`h-12`, padding 0 24px), compact buttons 36px (`h-9`), nav "Launch app" 48px full-height block.
- Feel: large type and generous cells, but labels and buttons are tiny (11px), so it is airy at the macro level and fine-grained at the micro level.

**Components visible** [CSS][SHOT]

- **Primary button**: solid ink block (`--invert-bg`), 48px high, square, uppercase mono 11px bold, leading 16px Lucide icon; here a **split button** with a 1px divider and a chevron segment. Hover: opacity 80%.
- **Outline button**: 1px `--line-strong` border on page background; hover **inverts** to ink fill (`.invert-block`, 120ms linear).
- **Segmented button group** ("Share | Copy link | X | Reddit | ..."): one 1px outline, cells split by hairlines, 36px high.
- **Nav**: logo cell, uppercase mono links with small chevrons, one "current section" chip in inverted ink ("CRAFT APPS"), right-side icon cells (Discord, theme toggle) and a solid "LAUNCH APP" block.
- **Feature grid cards**: borderless cells in a hairline grid, index number top-right (mono, faint), Archivo 24px title, Inter body in muted ink.
- **Screenshot tiles**: image flush to the cell, caption strip below in uppercase mono with an index on the right; hover scales the image 1.03 over 500ms.
- **Download list rows**: title (Inter 16px), muted description, mono filename in faint ink, download icon at right, hairline dividers, hover `bg-sunken`.
- **Code block**: header strip ("TERMINAL" label + "COPY" button cell) over a `bg-sunken` body, 16px padding, square.
- **Badges**: tiny solid blocks (index chip on app colour with black text; plan chips).
- **App family cards**: screenshot, overlapping 56px app icon with drop shadow, two-tone name (ink + app colour), description, "EXPLORE X" link with arrow that nudges on hover.

**Motion** [CSS]

- Default 150ms; `duration-300` x64 and `duration-500` x32 with `ease-out` for image zoom and icon lift; hover inversion 120ms linear; ruler digits roll with `transform .28s cubic-bezier(.2,.9,.2,1)`; marquee 36s linear; scroll-reveal (`data-reveal`) fades sections in; smooth scrolling via Lenis (`lenis` class on `<html>`). Not observed in motion.

**Iconography** [CSS]

- **Lucide**, 24 viewBox, stroke 2, rendered at 12-16px (`h-3`/`h-3.5`/`h-4`). Brand glyphs (GitHub, Discord) as filled SVGs.

### 2B. FigCraft application UI (from screenshots only)

All of this is [SHOT] / [SHOT-px]; no CSS was available for the app.

- Layout: top tab strip with window controls; 54px icon rail (File / Assets / Variables with tiny labels); left panel (~240px: file title, Pages, Layers tree); central canvas; right inspector (~290px: Design / Prototype tabs, stacked property sections); floating bottom-centre toolbar; round "?" help button bottom right.
- Dark theme samples: canvas and title bar **#1e1e1e**, panels **#2c2c2c**, inputs and page row **#383838**, segmented active #3b3b3b, selected layer row violet-tinted #3e3052, primary/active tool **~#0d99ff**, "Ready for dev" badge green #1ba45d, component purple ~#9b72d4.
- Light theme samples: title bar #e8e8e8, panels #ffffff, canvas #f6f6f6, inputs #f6f6f6, selected layer row #e6f4ff, hover/child rows #f2faff, primary ~#0d9aff.
- Shape: inputs and rows have ~4-5px radius; the floating toolbar ~10px with a soft shadow and a 1px border; "Share" button ~5px radius, solid blue; avatar is a green circle.
- Density: very dense. Body labels look ~11-12px, section titles ~12px medium, rows ~32-38px in the layer tree, inspector inputs ~24px high, two-column property grids. Sections are separated by 1px lines, not cards.
- Components: tree with chevrons and type icons (frame #, component diamond in purple); segmented control with filled active segment; number inputs with a leading letter label (W, H, X, Y); selects with chevrons; checkbox (blue fill, white check); 3x3 alignment grid; colour-swatch row with opacity and eye/minus icons; tool buttons with dropdown carets, active tool as blue filled square.
- Icons: thin-stroke (about 1.25-1.5px) custom line icons; set not identifiable.
- App UI font: a humanist sans with a slightly rounded feel; family not determinable from the image.

### Overall feel

- Website adjectives: **technical, architectural, monochrome, confident, HUD-like**.
- Distinguishing traits versus generic SaaS:
  1. **Zero radius + hairline blueprint grid** with crosshair ticks, numbered section strips and a scroll ruler.
  2. **Expanded heavy grotesque display (Archivo 620 @ 118% width)** paired with **11px uppercase wide-tracked mono** for every label and button.
  3. **Ink-inverted buttons and alpha-of-one-ink neutrals**; accent colour is per product and used for a single word.

---

## 3. Herdr (https://herdr.dev/)

### What it is, and what the page shows

- A native desktop app (macOS, Windows, Linux, plus mobile companions) for running and supervising terminal coding agents; its layout, sidebar, tabs and themes are plain files that an agent can rewrite, hence "No two Herdrs look alike". Open source. [CSS] (meta, page text)
- The page shows the application heavily:
  - A tilted, scrolling wall of **community screenshots** of customised Herdr windows (hero). [SHOT]
  - A **live HTML/CSS recreation of the default app window** (`.hw`, authored at 1477x831 and scaled) with six switchable looks: Herdr (default), Terminals, Code, Board, Pasture, 1998. All tokens are in `/live/herdr-look.css`. [CSS][SHOT]
  - Mobile app screenshots (inbox, terminal, foldable). [SHOT]
- Secondary page examined: `/gallery/` (tabs, search, select, cards, drawer). [CSS][SHOT]
- **The website is dark only**: no `prefers-color-scheme` rule in any stylesheet, a single `theme-color` of #0e0f11. The only light surfaces are the "Board" and "Pasture" sample looks of the app mock. [CSS]

### 3A. Herdr website

**Typography** [CSS][CSS-computed]

- Two families, self-hosted TTF variables: **Geist** (100-900) for everything; **JetBrains Mono** (100-800, aliased `JBM`) for code, kbd hints, list numbers and markdown-style markers. No serif.
- Display: hero H1 `clamp(52px, 7vw, 96px)`, **weight 560**, line-height 0.98, letter-spacing **-0.045em** (96px / -4.32px rendered). Section H2 `clamp(36-44px, ~5vw, 64-72px)`, weight 560, line-height 1.02, -0.035em. Doc H2 34px / 1.1, -0.025em; H3 20px, -0.015em. Stat numerals 34px, 560, -0.03em, tabular-nums.
- Two-tone headlines: second line in `--mute` gray.
- Body: lead 19px / 1.55; section copy 18px / 1.6; doc copy 17px / 1.68; notes 14.5px / 1.55; nav 14px; captions 13-13.5px. Bold = weight 500 in `--fg`, never heavier.
- Mono: code blocks 13.5px / 1.75; inline code 14px; kbd 11px; section markers "##" in mono 24px, colour #3f4248 (markdown affectation); TOC numbers mono 12px.
- No uppercase, no wide tracking. Negative tracking grows with size.

**Colour** [CSS]

| Token | Value |
|---|---|
| `--bg` | **#0e0f11** [oklch(0.168 0.004 264)] |
| `--card` | #131416; code/box surfaces #121315, #18191c; raised controls **#1b1d21** |
| `--line` | **#1f2125**; stronger control border **#2b2e33**; link underline #33363c / #3a3d43 |
| `--fg` | **#f3f3f4** |
| `--mute` | #8a8e96 |
| `--dim` | #5d6168 |
| `--amber` | #e8912a (active TOC number); inline command text #e8b46a |
| Terminal semantic colours | blue #76a8ff (working), amber #f2a33c (needs you), green #4cc48c (done), coral #d97757, dim #62666e |

- Colour quantity: essentially **monochrome**. The primary CTA is a **white pill with near-black text**. Colour appears only as status (blue/amber/green dots and words) and inside screenshots.
- Text hierarchy is three grays (fg / mute / dim) on one background.

**Shape and depth** [CSS]

- Radii are **medium to large and varied**: pill buttons 999px; cards 14px; boxes, code frames and lightbox images 12px; `pre` 10px; segmented buttons and inputs 8px; small buttons 7px; inline code 5-6px; avatars 8px; phone shots 22-30px.
- Borders: 1px `--line` for section rules and boxes; raised controls get a 1px #2b2e33 border on #1b1d21.
- Depth pattern for anything image-like: **1px white-alpha ring + large soft drop shadow**, e.g. `0 0 0 1px #ffffff1a, 0 20px 50px -10px #000c` (tiles), `0 0 0 1px #ffffff17, 0 50px 120px -40px #000000d9` (scenes). Cards: `0 0 0 1px var(--line)`, on hover `0 0 0 1px #3a3d43, 0 30px 60px -30px #000c` with translateY(-2px).
- Gradients: fade masks over the screenshot wall and at the bottom of the app mock; one faint radial glow behind the app (`rgba(120,140,190,.14)`); a huge ram-logo watermark at ~4% white. No grid, no noise, no dashed lines on the site itself. Glass: lightbox `backdrop-filter: blur(10px)`, floating prompt `#16171aeb` + blur(14px).

**Density and spacing** [CSS]

- Wrap `max-width: 1280px`, side padding 32px. Nav 72px. Section padding 90-120px vertical; doc layout 220px TOC + 760px text column with 96px gap. Airy.
- Control heights: nav pill 34px (padding 0 14px), hero pill 44px (0 20px), large download pill 46px (0 24px), segmented/tabs 30-32px, small copy button 28-30px, search and select 34px, gallery type tabs 42px.
- Spacing values are ad hoc pixels (6, 8, 10, 12, 14, 18, 20, 26, 28, 32...), no strict 4px grid.

**Components visible** [CSS][SHOT]

- **Primary pill**: `--fg` fill, #0e0f11 text, weight 500, 14-16px; hover pure white + `0 0 24px #fff3` glow + translateY(-1px).
- **Ghost link**: muted text + arrow that slides 3px on hover.
- **Segmented selector** (look picker): 30px, 8px radius, 1px `--line` border, transparent; active = #1b1d21 fill + #2b2e33 border; trailing mono `kbd` number.
- **Tabs** (gallery): 32px, 8px radius, transparent border; active = #1b1d21 + #2b2e33; count in mono dim text.
- **Search field**: 34px, 8px radius, `--card` fill, 1px `--line`, leading "/" hint. **Select**: same metrics.
- **Card**: 14px radius, `--card` fill, 1px ring, 16:10 screenshot, meta row (32px rounded avatar, 16.5px/560 title, 13px dim author, star count right) with a top hairline.
- **Code box**: 12px radius, #121315, tab strip on top (28px tabs, 7px radius), mono command with dim "$", "Copy" button (30px, 7px radius, bordered).
- **Accordion** (FAQ): hairline rows, 17px/500 summary, mono "+" that rotates 45 degrees.
- **Drawer/dialog**: centred, 16px radius, #111214, 1px #2b2e33, scale .98 to 1.
- **TOC**: sticky, mono file name header ("README.md"), numbered links, active number in amber.
- **Status words/dots** inside terminal samples.

**Motion** [CSS]

- Hero wall: columns drift at ~90s linear, alternate directions, pause on hover; wall is in 3D perspective (`rotateX(18deg) rotateY(-12deg) rotate(6deg)`).
- Hover micro-motion 150-250ms (`ease`), cards 200ms, drawer `.25s cubic-bezier(.2,.8,.2,1)`, app-mock pane changes `.5-.6s cubic-bezier(.2,.8,.2,1)`, look cross-fade 350ms. Terminal lines slide in 400ms; shimmer text 1.6s; spinner 900ms; caret blink 1s steps. `prefers-reduced-motion` respected. Not observed in motion.

**Iconography** [CSS]

- Inline SVG from `/live/icons.js`: mostly **Lucide** (34 icons carry `lucide` classes; 24 viewBox, stroke 2) plus a few custom ones drawn at stroke 1.5 (branch, search) and 16-viewBox glyphs at 1.3. Rendered at 11-15px. Agent brand marks as small images.

### 3B. Herdr application UI (default look, from the site's own HTML/CSS recreation)

Values are exact for the recreation ([CSS]); that the native app matches it pixel for pixel is [INF].

**Tokens (`.hw`)**

| Token | Value |
|---|---|
| `--canvas` (window background, gaps between panes) | #131518 |
| `--chrome` (sidebar, tab bar, status bar) | **#101114** |
| `--bg` (pane surface) | **#17191d** |
| `--raised` (active tab, chips) | #2c2e31 |
| `--hover` / `--selected` | white 5.5% / white 9% |
| `--border` | white 9.4% |
| `--fg` / `--tfg` (terminal text) / `--muted` / `--dim` | #eeeeef / #d2d4d9 / #98999b / #696a6d |
| `--accent` (focused pane border) | **#96b2d7** (soft desaturated blue) |
| status | working #76a8ff, blocked #f2a33c, done #4cc48c, idle #696a6d |
| terminal palette | red #dc8c95, green #a6bd91, yellow #d8be87, blue #96b2d7, magenta #b5a2ce, cyan #8cbdc3, gray #818793, white #f0f0f2 (muted pastels) |
| diff | deletion rgba(220,100,110,.28), addition rgba(120,190,120,.22) |

**Metrics**

- Base text 13px / 1.25 Geist. Terminal text JetBrains Mono 13.75px / 21.5px, padding 12px 14px, ligatures off.
- Window radius 12px (0 on Linux); shadow `0 0 0 1px rgba(255,255,255,.08), 0 40px 120px -30px rgba(0,0,0,.6)`.
- Sidebar 268px, `--chrome`, 1px right border. Brand row 46px (13.5px, weight 600; macOS traffic lights). Section headers 31-41px tall, 11.5px dim. **Rows 30px high, 8px radius, 0 8px padding, 9px gap**; hover/selected are white-alpha fills. Group labels 11px dim, 26-30px tall. Footer rows 28px.
- Avatars: 18px squares, 5px radius, coloured fill with a white 10.5px/700 initial.
- Tab bar 42px on `--chrome` with bottom border; tabs 29px high, 7px radius, 12.8px text; active = `--raised` fill + 1px border; inactive transparent; trailing status ring or dot.
- Work area padding 8px, **panes as separate rounded cards**: 10px radius, 1px border, 8px gap; the focused pane's border turns `--accent`. Pane header 30px, 12px text, bottom border (name weight 500 + dim cwd).
- Status bar 28px, 11.8px muted text, top border.
- Icon buttons 26px square, 6px radius, hover fill. Text buttons 26px, 12px text.
- **Badge** ("Needs you"): 18px high, 5px radius, amber at 15% fill, amber text, 10.5px/600.
- Chips (agent tree): 30px, 8px radius, `--raised` + border, mono 11.5px id.
- Floating command prompt: 14px radius, translucent #16171a + blur, 1px white-12% ring, mono 15px.

**Light variant seen** ("Board" sample look, an example layout rather than an official light theme) [CSS][SHOT]

- Background #f6f7f9, surfaces #ffffff, border #e5e7ec, text #252b37 / #303746, muted #89909d, accent violet #7465db (selected chip #eeebfc fill, #ddd7f9 border), column cards 7px radius with a 3px coloured left bar, filter chips 27px / 6px radius. Terminal palette switches to saturated dark hues on white.

### Overall feel

- Adjectives: **calm, monochrome, soft-edged, terminal-native, roomy (site) / compact (app)**.
- Distinguishing traits versus generic SaaS:
  1. **No brand colour**: a white pill CTA, three grays, and colour reserved for agent status (blue working, amber needs-you, green done).
  2. **Panes as rounded cards floating on a darker canvas**, with a pale-blue focus border and white-alpha hover/selection, giving a tiling-window-manager feel.
  3. **One sans at weight 560 with very tight tracking** plus JetBrains Mono, and markdown/terminal affectations ("##" headings, "README.md" TOC, `$` prompts).

---

## Common ground and tensions

### What the three share

- **Tool-for-builders aesthetics**: each shows a real multi-pane work surface (sidebar or rail, tabs, panels, status bar) rather than abstract illustration.
- **Near-black, slightly cool dark mode** with almost no chroma: Zed site #111216, ArtCraft #121316, Herdr #0e0f11; app surfaces step up in small increments (Herdr #101114 / #131518 / #17191d; Zed #282c33 / #2f343e / #3b414d). All sit around oklch L 0.17-0.37, C < 0.025, hue ~260-275.
- **Off-white ink, never pure white** for body text (#f3f3f4, #f2f1ee, #dce0e5), with two muted steps for hierarchy (alpha-based at ArtCraft and Herdr app, ramp-based at Zed and Herdr site).
- **1px hairlines carry the structure.** Shadows are either absent (ArtCraft, inside Zed's window) or reserved for floating objects (Herdr windows, Zed's hero mock).
- **Monospace as a second voice**: kbd hints, counts, indices, paths, metadata.
- **Colour restraint**: one accent at most, plus semantic status colours. Screenshots provide the colour.
- **Small type and compact controls in app chrome**: 11-13px UI text, 24-36px controls, 27-30px bars and rows.
- **Lucide icons** at 12-16px on all three sites (stroke 2 by default; Herdr redraws some at 1.5).
- **Fast, plain motion**: 70-200ms colour/opacity transitions, small translate/scale (1-2px, 0.98-1.03), a springy `cubic-bezier(.2,.8,.2,1)`-style ease for panels. No bouncy or decorative animation on controls.
- **Keyboard cues** visible in the UI (Zed's kbd-in-button, Herdr's numbered segments and "/" search hint).

### Where they conflict

| Axis | Zed | FigCraft / ArtCraft | Herdr |
|---|---|---|---|
| Corner radius | 2-4px (6px window) | **0** on the site; ~4-5px in the app | **7-14px**, pills at 999px |
| Display type | Light **italic serif** (IBM Plex Serif, w 320-390) | **Expanded heavy grotesque** (Archivo 620, 118% width) | Neutral sans (Geist 560), very tight tracking |
| Body type | Quasi-mono (iA Writer Quattro) | Inter | Geist |
| Label style | Sentence case, normal tracking | **UPPERCASE mono, +0.12-0.15em** | Sentence case, normal tracking |
| Light theme | Warm cream paper #f5f4f3 with cool gray text | Warm gray paper #f2f1ee with black ink | None on the site (sample "Board" look is cool white) |
| Accent | Saturated blue #1348dc (dark: #90c5ff text) | Black/white inverted blocks; per-product hue (green #02da96) for one word | None; white pill; pale blue #96b2d7 focus; status hues |
| Depth cue | Inset keycap edge, hard offset blocks | None | 1px light ring + big soft shadow |
| Decoration | Grid texture, diamond nodes, dashed frames, dither | Crosshair ticks, section index strips, scroll ruler | Clean; fades and one glow |
| Density | Dense app (12px, 27px bars), compact site | Airy site with tiny labels; very dense app | Airy site; medium-dense app (13px, 30px rows, 8px gutters) |
| Pane model | Flush panes divided by 1px lines | Flush panels divided by 1px lines | **Separate rounded cards** with gaps |
| Primary button | Blue, 36px, 4px radius, kbd chip | Ink block, 48px, square, uppercase mono | White pill, 34-46px |

The sharpest conflicts are **radius** (0 vs 4 vs 14 + pill), **label casing** (uppercase tracked mono vs sentence case) and **display typeface** (serif italic vs expanded grotesque vs plain sans). These cannot all be honoured at once; one must be chosen per axis.

### Traits that combine into one coherent system

These are compatible with each other and with shadcn/ui + Tailwind:

- **Surfaces**: cool near-black dark theme (about #0e0f11-#121316 base, two or three raised steps of +3-5% lightness) and a warm paper light theme (about #f2f1ee-#f6f5f4 base with a slightly lighter raised step). All three precedents support this pairing (Zed and ArtCraft directly).
- **Neutrals as ink at fixed alphas** (ArtCraft) or white-alpha hover/selected (Herdr app): 5-6% hover, 9% selected, ~10-16% border, ~40% strong border, ~60% muted text, ~40% faint text. This maps cleanly onto shadcn's `border`, `muted`, `accent` tokens and works in both themes.
- **Structure by 1px borders, not shadows**; reserve one shadow recipe (1px ring + large soft blur) for popovers, dialogs and command palettes.
- **Small radius, about 4-6px** for controls and 8-10px for floating surfaces: the midpoint between Zed and Herdr, and it matches what FigCraft's app actually uses. (Going to 0 would follow the ArtCraft site only; going to pills would follow Herdr only.)
- **Type roles**: a neutral sans for UI at 13-14px (Geist/Inter both evidenced), a mono for metadata, counts, kbd and paths (Geist Mono / JetBrains Mono / Lilex all evidenced), and a display treatment used sparingly. Tight negative tracking on large headings is common to ArtCraft and Herdr (-0.02 to -0.045em).
- **Mono micro-labels** for section indices, table headers and status text. Uppercase + tracking (ArtCraft) is optional and can be limited to very small labels without clashing with the other two.
- **Single accent + status palette**: a blue in the Zed/ArtCraft range (#1348dc-#2d81ff; lighter #74aaff-#90c5ff on dark) or a neutral inverted primary (ArtCraft/Herdr); status colours blue = working, amber = needs attention, green = done, red = error (Herdr and Zed agree on these meanings). Status backgrounds as the same hue at 10-15% alpha (Zed dark tokens, Herdr badge).
- **Compact control scale**: 24px (chips), 28-30px (rows, small buttons), 32px (default), 36px (prominent); 27-30px toolbars and status bars; icons 12-16px.
- **kbd chips inside buttons and menus** (Zed) and numbered/tabbed selectors (Herdr).
- **Blueprint accents as optional garnish**: corner ticks, diamond nodes at line intersections, a faint 10px grid, numbered section strips. Zed and ArtCraft both use this vocabulary, and it does not conflict with Herdr's cleaner look if applied only to page-level framing, not to components.
- **Motion**: 100-150ms colour transitions, 200-250ms `cubic-bezier(.2,.8,.2,1)` for overlays, scale .96-.98 to 1 on open.

Decisions the references leave open (a human choice, not derivable): flush panes vs carded panes; coloured vs inverted-neutral primary button; whether to adopt a characterful display face (serif italic or expanded grotesque) or stay with the UI sans.

---

## Not verified / limitations

- **Zed native app**: exact UI and buffer font families, sizes, spacing and icon set were not readable; the font names given for the real app are prior knowledge. The 12px/27px metrics come from the website's HTML recreation, whose UI labels render in the site's body font. No real light-theme app screenshot was found on the pages fetched; light app colours come from the inlined theme variables only. The "One Light / One Dark" naming is from prior knowledge.
- **Zed full-page capture** (`zed-home-full.png`) shows the hero app mock overflowing over later sections; this is a capture artefact, so lower-page layout was judged from class strings rather than that image.
- **Zed dark mode** was produced by switching the `html` class from `light` to `dark` in the remote browser; how the site chooses its default (system preference vs stored choice) was not traced.
- **FigCraft app**: all colours are pixel samples from lossy WebP images; fonts, exact sizes, radii and the icon set are visual estimates. The collaboration video was not watched.
- **FigCraft dark full page**: scroll-reveal left most sections blank in the dark full-page capture; dark mode was verified visually only for the hero and the highlights section, the rest from CSS variables.
- **Herdr**: the native app itself was not observed; the default look is the site's own HTML/CSS recreation, and community tiles are user-customised. Herdr has no official light theme evidenced; "Board" is a sample layout. Mobile screenshots were looked at only briefly.
- **Motion** for all three is described from CSS declarations and class names; nothing was watched animating. Hover, focus and active states are from CSS, not seen.
- **Command palettes** were not visible on any page (Zed shows only the shortcut hint; Herdr shows a floating prompt in the mock). **Tables** appear only as a mono terminal listing (Herdr) and inside FigCraft's sample dashboard design, which is user content, not product UI.
- oklch figures are my conversions from hex, not values published by the sites.
