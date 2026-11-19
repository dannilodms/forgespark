# 03: The images of a commit

**What to build:** A server runs images, not tools on the machine. Every commit gets an image of the API, whose binary carries its commands (serve, migrate, worker), and an image of Caddy with the built web app inside. The checks of every pull request build both and publish nothing, as they already do for the seed's image. See the spec, section Images.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [ ] An image of the API runs the binary with each of its commands: serve, `migrate`, `worker`. `issue-setup-code` stays reachable for the installer.
- [ ] The web app is built with the same command the checks use, and its output goes into an image of Caddy that keeps the rate-limit plugin. Node keeps its one home in `apps/web/package.json`: no Dockerfile names a Node version.
- [ ] Task targets build each image locally and publish nothing, as `seed:image` does.
- [ ] The backend job of the pipeline (or the job that builds the web app) builds both images on every pull request, so that a change that breaks either fails before anything is published.
- [ ] `task versions:check` accepts every new pinned version, and every new image tag is exact.
- [ ] The development Compose file and `task dev` behave as before: development still serves the web app through Vite.
- [ ] `CODING_STANDARDS.md` (Checks on a pull request, Publishing, Versions) names the images and where their versions live.
