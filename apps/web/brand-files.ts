import { readFile } from "node:fs/promises"
import type { ServerResponse } from "node:http"
import { resolve } from "node:path"

import type { Connect, Plugin } from "vite-plus"

// The files of the brand the app serves at the root of its origin, as
// `index.html` names them, with the media types Caddy gives them on the
// servers. Their one home is docs/design/brand (Brand mark in DESIGN.md): the
// app holds no copy of them.
const brandFiles = new Map([
  ["favicon.ico", "image/vnd.microsoft.icon"],
  ["favicon.svg", "image/svg+xml"],
  ["apple-touch-icon.png", "image/png"],
  ["icon-192.png", "image/png"],
  ["icon-512.png", "image/png"],
  ["icon-maskable-512.png", "image/png"],
])

// The web app manifest, which names the app and its icons to a browser that
// installs it. Its name and description are the brand's, in English in every
// language (Brand mark in DESIGN.md), and its colours are `background` of
// the light theme.
const manifestName = "manifest.webmanifest"
const manifest = JSON.stringify({
  id: "/",
  name: "ForgeSpark",
  short_name: "ForgeSpark",
  description: "From ideas to execution.",
  lang: "en",
  start_url: "/",
  scope: "/",
  display: "standalone",
  background_color: "#eceff8",
  theme_color: "#eceff8",
  icons: [
    { src: "/icon-192.png", sizes: "192x192", type: "image/png" },
    { src: "/icon-512.png", sizes: "512x512", type: "image/png" },
    {
      src: "/icon-maskable-512.png",
      sizes: "512x512",
      type: "image/png",
      purpose: "maskable",
    },
  ],
})

// Serves the files of the brand and the manifest at the root: read from
// docs/design/brand by the development server, and written beside
// `index.html` by the build, without a hash, so Caddy revalidates them
// (Caching in Delivery, CODING_STANDARDS.md).
export function brandFilesAtTheRoot(): Plugin {
  let brandFolder = ""
  return {
    name: "brand-files-at-the-root",
    configResolved(config) {
      brandFolder = resolve(config.root, "../../docs/design/brand")
    },
    configureServer(server) {
      server.middlewares.use(serveBrandFiles(() => brandFolder))
    },
    async generateBundle() {
      const files = await Promise.all(
        [...brandFiles.keys()].map(async (name) => ({
          name,
          content: await readFile(resolve(brandFolder, name)),
        }))
      )
      for (const { name, content } of files) {
        this.emitFile({ type: "asset", fileName: name, source: content })
      }
      this.emitFile({ type: "asset", fileName: manifestName, source: manifest })
    },
  }
}

function serveBrandFiles(
  brandFolder: () => string
): Connect.NextHandleFunction {
  return (request, response, next) => {
    const name = new URL(request.url ?? "/", "http://app").pathname.slice(1)
    if (name === manifestName) {
      send(response, "application/manifest+json", manifest)
      return
    }
    const type = brandFiles.get(name)
    if (type === undefined) {
      next()
      return
    }
    readFile(resolve(brandFolder(), name)).then(
      (content) => send(response, type, content),
      next
    )
  }
}

function send(response: ServerResponse, type: string, body: string | Buffer) {
  response.setHeader("Content-Type", type)
  response.setHeader("Cache-Control", "no-cache")
  response.end(body)
}
