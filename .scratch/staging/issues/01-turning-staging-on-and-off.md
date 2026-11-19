# 01: Turning staging on and off

**What to build:** From the maintainer's machine, one command turns staging on: a server comes up at Hetzner, prepared by the cloud-init file, and the command returns once the server answers SSH with the host key the tool gave it and cloud-init has finished. Another command turns it off and destroys it, and a third tells whether it is on, since when and at which address and host names. The tool is a Go command of its own, among the repository's tools, reaching the provider through an interface whose only adapter today is Hetzner's. See the spec, sections The tool and The server.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [ ] A Go command of its own, in the Go module of the API among the repository's tools and built into no image, has the commands to turn on, turn off and give the status of an environment. Task runs them as `staging:on`, `staging:off` and `staging:status`.
- [ ] The choices of staging (provider, region, size, system image, server name) are in a committed file the tool reads: Hetzner, `fsn1`, `cx23` and the newest Ubuntu LTS that Docker supports officially.
- [ ] The provider is an interface (create from a specification, find by name, destroy by name), with an in-memory implementation for the tests and a Hetzner adapter on `hcloud-go`.
- [ ] The Hetzner adapter registers the deploy public key at the project, so Hetzner e-mails no root password, and removes it with the server. It asks for an ordinary IPv4 address that is deleted with the server, labels what it creates with the environment, and creates no firewall at Hetzner.
- [ ] The Hetzner adapter is tested against a stand-in of Hetzner's API that answers as its documentation says. No test calls Hetzner.
- [ ] One cloud-init file prepares the server:
  - Docker from Docker's repository, at the version the file names;
  - the deploy user in Docker's group, with the deploy public key;
  - the host key the tool generated;
  - SSH by key only, with no password and no root;
  - a firewall inside the server that lets in 22, 80 and 443 only;
  - automatic security updates;
  - a signal that it has finished.
- [ ] Readiness is an interface: wait until SSH answers with the expected host key and cloud-init reports it has finished, within about ten minutes. Its SSH implementation is tested against an SSH server the test starts in-process, which also proves that a different host key is refused.
- [ ] Turning on creates one server. Turning on when it is on creates nothing and reports. A failure after the server was created destroys it, removes the local key and exits with an error. Each is proven with the fake provider.
- [ ] A copy of the deploy private key is kept in a git-ignored folder of the main checkout while staging is on, and removed when it is turned off.
- [ ] The token is `HETZNER_TOKEN`, from `.env`. When the tool runs in a worktree it reads the `.env` of the main checkout. The name is not in `.env.example`.
- [ ] No command prints a secret: the tests check what they print.
- [ ] Turning staging on for real works, `ssh deploy@<address>` with the local key runs `docker version`, root and password logins are refused, and turning it off leaves nothing at the provider.
- [ ] `docs/providers.md` describes the Hetzner project as it now is, and `CODING_STANDARDS.md` (Servers, Versions) and the `README.md` say how staging is turned on and off.
