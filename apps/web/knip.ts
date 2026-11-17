import type { KnipConfig } from "knip"

// Knip fails `task web:knip`, and so `task check`, on an unused file, export
// or dependency (Enforcement in CODING_STANDARDS.md).
export default {
  // The design system is kept whole, as DESIGN.md lists it: a component no
  // screen uses yet is not dead code.
  entry: ["src/ui/**/*.{ts,tsx}", "scripts/*.mjs"],
  project: ["src/**/*.{ts,tsx,css}", "scripts/*.mjs"],
  ignoreIssues: {
    // orval writes a function for every operation and status, used or not.
    "src/shared/api/generated/**": ["exports", "types"],
    // `#ui/lib/*` is where the shadcn CLI writes `cn` for a new component
    // (components.json), even while no component imports it that way.
    "package.json": ["subpaths"],
  },
} satisfies KnipConfig
