import { defineConfig } from "orval"

// The client, the Zod schemas and the simulated responses of the API, written
// by orval from the committed spec into src/shared/api/generated (Contract in
// CODING_STANDARDS.md). `task web:generate` writes them, and
// `task web:generate:check` fails when the committed files are stale.
const output = "src/shared/api/generated"
const spec = "../api/openapi.yaml"

export default defineConfig({
  client: {
    input: { target: spec },
    output: {
      mode: "split",
      target: `${output}/client.ts`,
      schemas: `${output}/model`,
      client: "fetch",
      clean: true,
      // A simulated answer for every status the spec declares, so a screen
      // test can answer a 422 or a 429 as the API and Caddy do.
      mock: {
        generators: [{ type: "msw", generateEachHttpStatus: true }],
      },
      override: {
        // Rejects with an ApiError when the answer is a failure, so the
        // functions resolve with the body of a success only.
        mutator: { path: "src/shared/api/request.ts", name: "request" },
        fetch: { includeHttpResponseReturnType: false },
      },
    },
  },
  schemas: {
    input: { target: spec },
    output: {
      mode: "single",
      target: `${output}/schemas.ts`,
      client: "zod",
    },
  },
})
