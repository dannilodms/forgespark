import { defineConfig } from "vite-plus"

// Format, lint and type checks for the whole JavaScript workspace. `vp check`
// reads only this file: a package does not carry its own `lint` or `fmt` block.
export default defineConfig({
  fmt: {
    // The style the shadcn CLI writes, so an added component needs no reformat.
    semi: false,
    trailingComma: "es5",
    printWidth: 80,
    ignorePatterns: [
      "**/routeTree.gen.ts",
      "pnpm-lock.yaml",
      // Documents and agent files are not part of the JavaScript workspace.
      "**/*.md",
      ".claude/**",
      ".github/**",
      ".scratch/**",
      "docs/**",
      "deploy/**",
      "skills-lock.json",
    ],
  },
  lint: {
    plugins: ["typescript", "react"],
    ignorePatterns: ["**/routeTree.gen.ts", "**/dist/**", ".claude/**"],
    options: {
      typeAware: true,
      typeCheck: true,
    },
  },
})
