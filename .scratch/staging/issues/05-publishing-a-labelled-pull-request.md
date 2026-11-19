# 05: Publishing a labelled pull request

**What to build:** A pull request from this repository with the `staging` label is published to staging once its checks pass. The pipeline pushes the images of the commit to the GitHub container registry and runs the tool's publish command. That command logs in over SSH and writes the variables file, after checking that every name of `.env.example` has a staging value. It then starts the data again, applies the migrations, runs the seed and starts the API and the worker. A comment on the pull request gives the address of the app and of Mailpit. When staging is off, the job skips with a notice and passes. See the spec, sections The publish and The pipeline.

**Blocked by:** 02, 04.

**Status:** ready-for-agent

- [ ] The tool has a publish command. Its inputs are the commit, the pull request, the address, the host key, the deploy key and the password of Mailpit.
- [ ] The remote server is an interface (write a file, run a command), with an in-memory implementation that records both in order, and an SSH implementation that trusts only the given host key.
- [ ] Every name of `.env.example` has a staging value of one of four kinds: fixed, derived from the address, drawn at random on each publish, or deliberately empty. A name with no decision stops the publish before any remote step. A test adds a name and proves it.
- [ ] The steps run in the order the spec gives:
  1. copy the files;
  2. write the variables file, readable by the deploy user alone;
  3. pull the images;
  4. stop the API and the worker;
  5. remove the volumes of PostgreSQL and Garage;
  6. start PostgreSQL, Garage and its rule, Mailpit and Caddy;
  7. run the migration, then the seed;
  8. start the API and the worker;
  9. wait for `/up`.

  A test with the fake server proves this order.
- [ ] Neither the command nor the job prints a value, the variables file or the configuration of a container. A test checks what the command prints.
- [ ] A job of `ci.yml` runs after `Checks`, only for a pull request from this repository with the label, in the GitHub Environment `staging`. It pushes the three images tagged with the commit and runs the publish. The required check does not change.
- [ ] When the address variable of `staging` is empty, the job skips with a notice and passes.
- [ ] Publishes share one concurrency group, and a running one is never cancelled.
- [ ] The job asks for writing packages and pull requests and only reading the contents. actionlint and zizmor pass.
- [ ] The published pull request gets a comment with the address of the app and of Mailpit.
- [ ] For real: the pull request of this effort, labelled, is published to a staging that is on. The app opens on the login screen over HTTPS, a login code asked for a user of the seed arrives in Mailpit, and logging in with it reaches the shell.
- [ ] `CODING_STANDARDS.md` (Publishing, Secrets, Seed) matches what was built.
