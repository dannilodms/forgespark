import { expect, test } from "playwright/test"

// The favicon, the icon of iOS and the web app manifest, at the root of the
// origin, as the page names them and a browser asks for them through Caddy
// (Brand mark in DESIGN.md). Each answers with the media type Caddy gives it
// on the servers, and so do the icons the manifest names.
test("the page names its favicon, its icon of iOS and its manifest, and each of them answers", async ({
  page,
  request,
}) => {
  await page.goto("/")
  const named = await page
    .locator(
      'link[rel="icon"], link[rel="apple-touch-icon"], link[rel="manifest"]'
    )
    .evaluateAll((links) => links.map((link) => link.getAttribute("href")))
  expect(named).toEqual([
    "/favicon.ico",
    "/favicon.svg",
    "/apple-touch-icon.png",
    "/manifest.webmanifest",
  ])

  for (const [path, type] of [
    ["/favicon.ico", "image/vnd.microsoft.icon"],
    ["/favicon.svg", "image/svg+xml"],
    ["/apple-touch-icon.png", "image/png"],
  ]) {
    const answer = await request.get(path)
    expect(answer.status(), path).toBe(200)
    expect(answer.headers()["content-type"], path).toBe(type)
  }

  const answer = await request.get("/manifest.webmanifest")
  expect(answer.status()).toBe(200)
  expect(answer.headers()["content-type"]).toBe("application/manifest+json")
  const manifest = (await answer.json()) as {
    name: string
    icons: { src: string; type: string; purpose?: string }[]
  }
  expect(manifest).toMatchObject({
    name: "ForgeSpark",
    description: "From ideas to execution.",
    start_url: "/",
    display: "standalone",
  })
  expect(manifest.icons.map((icon) => icon.purpose ?? "any")).toEqual([
    "any",
    "any",
    "maskable",
  ])
  for (const icon of manifest.icons) {
    const image = await request.get(icon.src)
    expect(image.status(), icon.src).toBe(200)
    expect(image.headers()["content-type"], icon.src).toBe(icon.type)
  }
})
