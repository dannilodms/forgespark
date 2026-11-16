// Fails when the gallery leaves something of packages/ui out:
//
// - a component that no block of the gallery imports;
// - a style variable that the Foundations route does not register.
//
// Run from anywhere: `vp node apps/prototype/scripts/gallery-coverage.mjs`.
import { readFileSync, readdirSync } from "node:fs"
import { basename, extname, join, resolve } from "node:path"

const root = resolve(import.meta.dirname, "../../..")
const componentsDir = join(root, "packages/ui/src/components")
const blocksDir = join(root, "apps/prototype/src/gallery/blocks")
const stylesheet = join(root, "packages/ui/src/styles/globals.css")
const registry = join(root, "apps/prototype/src/gallery/style-variables.ts")

const components = readdirSync(componentsDir)
  .filter((file) => extname(file) === ".tsx")
  .map((file) => basename(file, ".tsx"))

const blocks = readdirSync(blocksDir)
  .map((file) => readFileSync(join(blocksDir, file), "utf8"))
  .join("\n")

const missingComponents = components.filter(
  (name) => !blocks.includes(`@workspace/ui/components/${name}"`)
)

const css = readFileSync(stylesheet, "utf8")

/** The custom properties declared in the first block that opens with `opening`. */
function declaredIn(opening) {
  const start = css.indexOf(opening)
  if (start === -1) {
    throw new Error(`${stylesheet} has no "${opening}" block`)
  }
  const block = css.slice(start, css.indexOf("\n}", start))
  return [...block.matchAll(/^\s*--([\w-]+):/gm)].map((match) => match[1])
}

const variables = [...declaredIn(":root {"), ...declaredIn("@theme static {")]
const registered = readFileSync(registry, "utf8")
const missingVariables = variables.filter(
  (name) => !registered.includes(`"${name}"`)
)

console.log(
  `Gallery: ${components.length - missingComponents.length} of ${components.length} components shown, ` +
    `${variables.length - missingVariables.length} of ${variables.length} style variables registered.`
)

if (missingComponents.length > 0) {
  console.error(
    `No block in apps/prototype/src/gallery/blocks imports: ${missingComponents.join(", ")}`
  )
}

if (missingVariables.length > 0) {
  console.error(
    `Not registered in apps/prototype/src/gallery/style-variables.ts: ${missingVariables.join(", ")}`
  )
}

if (missingComponents.length > 0 || missingVariables.length > 0) {
  process.exit(1)
}
