import babel from "@rolldown/plugin-babel"
import tailwindcss from "@tailwindcss/vite"
import { tanstackRouter } from "@tanstack/router-plugin/vite"
import react, { reactCompilerPreset } from "@vitejs/plugin-react"
import { msw } from "msw/vite"
import { defineConfig } from "vite-plus"
import { playwright } from "vite-plus/test/browser-playwright"

// The build, the format, lint and type checks that `vp check` runs, and the
// screen tests that `vp test` runs.
export default defineConfig({
  plugins: [
    // The router plugin has to come before the React plugin.
    tanstackRouter({ target: "react", autoCodeSplitting: true }),
    react(),
    babel({ presets: [reactCompilerPreset()] }),
    tailwindcss(),
    // Serves the worker of MSW, which simulates the API in the screen tests.
    // The build emits nothing of it: the app never imports `virtual:msw`.
    msw(),
  ],
  optimizeDeps: {
    // Every source file, so the dependencies a lazy route or a test imports
    // are found at the start, and not halfway through a run of the screen
    // tests, which reloads the page under them.
    entries: ["src/**/*.{ts,tsx}"],
    // It imports `vitest`, which a pre-bundled copy would hold a second time,
    // with no runner. What it imports from React is pre-bundled instead.
    exclude: ["vitest-browser-react"],
    include: ["react-dom/client"],
  },
  fmt: {
    // The style the shadcn CLI writes, so an added component needs no reformat.
    semi: false,
    trailingComma: "es5",
    printWidth: 80,
    // Generated files: the route tree, the lists of the API (`task
    // api:lists`) and the client of orval.
    ignorePatterns: [
      "**/*.gen.ts",
      "pnpm-lock.yaml",
      "src/shared/api/generated/**",
    ],
  },
  lint: {
    plugins: ["typescript", "react"],
    ignorePatterns: [
      "**/*.gen.ts",
      "**/dist/**",
      "src/shared/api/generated/**",
    ],
    options: {
      typeAware: true,
      typeCheck: true,
    },
  },
  test: {
    // Every test runs in a real Chromium (ADR-0014). `task install`
    // downloads it, at the version the Playwright of the catalog names.
    browser: {
      enabled: true,
      headless: true,
      provider: playwright(),
      instances: [{ browser: "chromium" }],
    },
    // Starts the simulated API for each test file (src/testing).
    setupFiles: ["src/testing/setup.ts"],
  },
})
