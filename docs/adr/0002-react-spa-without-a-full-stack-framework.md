---
status: accepted
---

# React single-page app, without a full-stack framework

The frontend is a React single-page app with TanStack Router, built with Vite+, which supplies the whole frontend toolchain through one CLI (`vp`): dev server and build, tests, lint, format, and the Node.js runtime and package manager. Its static build is served by a Caddy container that also forwards `/api` to the Go backend, so the app and the API share one origin and Caddy handles HTTPS certificates on its own. The file storage is the one other address the browser reaches, for the bytes of a file alone (ADR-0017).

The whole application sits behind login, so there is no SEO or anonymous first-load to serve and server-side rendering buys nothing. A full-stack framework would add a Node server in production whose server-side features duplicate the Go API.

## Considered Options

- **Next.js or TanStack Start.** Rejected for the reason above: a second server process competing with the Go API.
- **Vue or Angular.** The maintainer knows Vue, but agents write the code, and React has by far the largest body of examples and ready-made components for them to draw on.
- **React Router.** Older and better known to models, but TanStack Router types routes and search parameters, which catches agent mistakes at compile time.
- **Embedding the build in the Go binary.** One container instead of two. A separate Caddy container was chosen instead.
