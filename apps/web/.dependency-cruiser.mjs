// The dependency rule of the frontend (Dependency rule in CODING_STANDARDS.md):
// routes → a domain area → shared → ui. `task web:imports` runs it, and so
// does `task check`. A finding names the rule, the file and the import.
//
// Where a file is, by its path under src:
// - routes: `routes/`
// - a domain area: any other folder directly under `src`, such as `consultancy/`
// - shared: `shared/`, with the generated client in `shared/api/generated/`
// - ui: `ui/`, the design system
// - the test support of the screen tests: `testing/`
// - the root: the files directly in `src` (`main.tsx`, `app.tsx`, the route
//   tree), which put the app together and may import anything.

// Non-capturing, so that $1 in a rule is the name of the area.
const notAnArea = "(?:routes|shared|ui|testing)/"
const anArea = `^src/(?!${notAnArea})[^/]+/`
const anAreaByName = `^src/(?!${notAnArea})([^/]+)/`
const aTest = "\\.test\\.tsx?$"

/** @type {import("dependency-cruiser").IConfiguration} */
export default {
  forbidden: [
    {
      name: "ui-imports-only-ui",
      comment:
        "ui is the design system: it imports third-party libraries and itself, nothing else of the app.",
      severity: "error",
      from: { path: "^src/ui/" },
      to: { path: "^src/", pathNot: "^src/ui/" },
    },
    {
      name: "shared-imports-no-area",
      comment:
        "shared imports ui, third-party libraries and itself: no area, no route, no test support and not the root.",
      severity: "error",
      from: { path: "^src/shared/", pathNot: aTest },
      to: { path: "^src/", pathNot: "^src/(shared|ui)/" },
    },
    {
      name: "routes-import-the-index-of-areas-and-shared",
      comment:
        "A route file declares a path and opens a screen: it imports the index.ts of an area and shared.",
      severity: "error",
      from: { path: "^src/routes/" },
      to: {
        path: "^src/",
        pathNot: ["^src/(routes|shared)/", `${anArea}index\\.ts$`],
      },
    },
    {
      name: "an-area-imports-another-through-its-index",
      comment:
        "An area uses another area only through the index.ts of that area.",
      severity: "error",
      from: { path: anAreaByName },
      to: {
        path: anArea,
        pathNot: ["^src/$1/", `${anArea}index\\.ts$`],
      },
    },
    {
      name: "an-area-imports-no-route-and-not-the-root",
      comment:
        "An area imports its own files, the index.ts of other areas, shared and ui. Its tests also import the test support.",
      severity: "error",
      from: { path: anArea },
      to: { path: "^src/(routes/|[^/]+\\.tsx?$)" },
    },
    {
      name: "only-the-data-module-uses-the-generated-client",
      comment:
        "The client orval generates is used by the data module of an area. The exceptions: the fetch it calls (shared/api), the list of supported languages, which comes from the contract (shared/locales/languages.ts), and the tests with their support, which answer with the simulated responses.",
      severity: "error",
      from: {
        path: "^src/",
        pathNot: [
          `${anArea}data/`,
          "^src/shared/api/",
          "^src/shared/locales/languages\\.ts$",
          "^src/testing/",
          aTest,
        ],
      },
      to: { path: "^src/shared/api/generated/" },
    },
    {
      name: "the-test-support-is-for-tests",
      comment: "Only a test and the test support import src/testing.",
      severity: "error",
      from: { path: "^src/", pathNot: ["^src/testing/", aTest] },
      to: { path: "^src/testing/" },
    },
  ],
  options: {
    doNotFollow: { path: "node_modules" },
    tsConfig: { fileName: "tsconfig.json" },
    // An import of a type is a dependency too.
    tsPreCompilationDeps: true,
    enhancedResolveOptions: {
      extensions: [".ts", ".tsx", ".js", ".json"],
    },
  },
}
