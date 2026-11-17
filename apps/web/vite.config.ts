import babel from "@rolldown/plugin-babel"
import tailwindcss from "@tailwindcss/vite"
import { tanstackRouter } from "@tanstack/router-plugin/vite"
import react, { reactCompilerPreset } from "@vitejs/plugin-react"
import { defineConfig } from "vite-plus"

// The build, and the format, lint and type checks that `vp check` runs.
export default defineConfig({
  plugins: [
    // The router plugin has to come before the React plugin.
    tanstackRouter({ target: "react", autoCodeSplitting: true }),
    react(),
    babel({ presets: [reactCompilerPreset()] }),
    tailwindcss(),
  ],
  fmt: {
    // The style the shadcn CLI writes, so an added component needs no reformat.
    semi: false,
    trailingComma: "es5",
    printWidth: 80,
    ignorePatterns: ["**/routeTree.gen.ts", "pnpm-lock.yaml"],
  },
  lint: {
    plugins: ["typescript", "react"],
    ignorePatterns: ["**/routeTree.gen.ts", "**/dist/**"],
    options: {
      typeAware: true,
      typeCheck: true,
    },
  },
})
