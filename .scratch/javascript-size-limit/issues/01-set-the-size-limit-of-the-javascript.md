# 01: Set the size limit of the JavaScript

**What to build:** The JavaScript of `apps/web`, initial and per route, gets a size limit that fails the build. It is measured on the Brotli copies that `task web:build` writes, which is what the browser downloads. This closes the rule in Enforcement of the frontend rules of `CODING_STANDARDS.md`, which sets the limit and leaves its numbers and its tool open, and the line on it in Deferred there, which points here. The numbers and the tool are chosen in this ticket: staging already serves the built app as production will, with the copies they measure. The sizes when the images of staging were built, on 2026-11-20: the entry `index-*.js` 122 KB and `form-fields-*.js` 134 KB in Brotli, 444 KB and 501 KB minified, where Vite warns above 500 KB; the CSS 16 KB.

**Blocked by:** None (can start immediately). It is in a folder of its own so that it outlives the effort of staging, whose folder is removed when that effort ends.

**Status:** ready-for-agent

- [ ] A tool is chosen that measures the Brotli copies of the build, the initial JavaScript and that of each route, with the reason for it. Its version has the home Versions in `CODING_STANDARDS.md` gives it, and `task versions:check` passes.
- [ ] The numbers are chosen from the sizes of the build at that time, with the reason for each.
- [ ] A build whose JavaScript passes a limit fails `task check:quick` and the app job of the pipeline, naming what passed which limit and by how much.
- [ ] Enforcement in the frontend rules of `CODING_STANDARDS.md` points at the check, and Deferred there no longer names the limit.
