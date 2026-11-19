# 10: Periodic cleanup of the storage

**What to build:** Every hour the worker removes from the file storage the uploads started more than 24 hours ago and never confirmed, and the objects that no confirmed file names, so abandoned uploads and the leftovers of a failed removal do not stay for ever. Rules: Files and Jobs in `CODING_STANDARDS.md`, ADR-0017.

**Blocked by:** 01 (River and the worker command).

**Status:** resolved

- [x] The `s3` adapter lists the bucket with the last-modified date of each object and removes an object by key; its tests run against Garage under keys of their own.
- [x] A use case removes each object older than 24 hours that no confirmed file names, and marks as abandoned or removes the file records of uploads started more than 24 hours ago and never confirmed; a confirmed file is never touched.
- [x] A periodic job declared in the worker runs the use case every hour, in the light queue with a longer timeout; it tolerates a missed run.
- [x] A second run over the same state changes nothing; use case tests cover objects named, unnamed, young and old, and the second run, with a fixed clock.
- [x] The job's arguments carry no personal data, and the logs carry no key of an object.

## Comments

### What was built

- `files/domain`: `ListedObject` (the key and the time the storage says the object was last written), `AbandonedBefore(now)` (a day before now) and `ListedObject.IsLeftBehind`.
- `files/application`: the use case `CleanUpStorage`. It deletes, in a transaction, the files whose upload started more than a day ago and was never confirmed, then lists the storage and removes, page by page, each object more than a day old that no confirmed file names (one query per page). The interfaces gained `FileRepository.RemoveAbandoned`, `FileRepository.ConfirmedKeys` and `Storage.List`. The in-memory storage now takes a clock and writes each object at its time (a copy at the time of the copy), lists in pages of two and has `FailListingWith`; its two callers (the files and consultancy test harnesses) pass their clock.
- `files/adapters/s3`: `List`, on the `ListObjectsV2` paginator, calling a function with each page. `keyless.go`: every error of the adapter, and every text the SDK records on the span of a call (`exception.message`, `api.error_message`), has each URL cut to the address of the storage and each object key replaced by `[key]`; the error still wraps the original for `errors.Is`/`errors.As`. `export_test.go` sets the page size for the paging test.
- `files/adapters/postgres`: `RemoveAbandonedFiles` and `ConfirmedObjectKeys` (sqlc). Migration `00020_let_the_api_role_remove_abandoned_files.sql`: `GRANT DELETE ON files TO api` and a partial index on `started_at` of the unconfirmed files.
- `files/adapters/river`: `CleanUpStorageArgs` (empty, kind `clean_up_storage`, light queue) and `CleanUpStorageWorker`, whose `Work` calls the use case and whose `Timeout` is ten minutes.
- `platform/jobs`: `AddPeriodic(kinds, worker, interval)` registers a job worker and a River periodic job of its kind together, with `RunOnStart`, and arguments that are the zero value of their type. `RunnerSettings.PeriodicJobs` is gone: `Kinds` carries the periodic jobs.
- `cmd/api`: `mountJobs` schedules the cleanup every hour, one line beside the delivery of e-mail of ticket 02, and `cleanUpStorage` wires it; `fileStorageOf` builds the storage for `serve` and `worker` alike; `config.Worker` requires the six `FILE_STORAGE_` variables beside the database and the sender of e-mail. The seed's `noFileStorage` gained `List`.
- Records: `CODING_STANDARDS.md` (Jobs: Running, Periodic work; Files: Access, Never confirmed; Deferred: the first record removed), ADR-0016 (the run at the start), ADR-0017 (the job exists; what `ListObjectsV2` and `CopyObject` did against Garage), `README.md` and `README.pt-BR.md` (the storage variables the worker requires, what the worker does in development).

### Decided while building

- **The file of an abandoned upload is deleted, not marked.** Abandoned follows from time (never confirmed, started more than a day ago), and Domain modelling says a condition that follows from time is computed and never stored. Nothing refers to such a file: the setup confirms an image before it saves the consultancy or the user that names it. It is the first record the system removes, so Deferred in `CODING_STANDARDS.md` now says so and keeps the question open for business records.
- **One day for both rules, and "more than" is strict:** an upload started exactly a day ago, and an object written exactly a day ago, are kept until the next run.
- **The files go first, then the objects.** A late confirmation holds its file locked; the `DELETE` waits for it and then finds it confirmed (proven by `TestRemovingTheAbandonedFilesWaitsForAConfirmationThatHoldsOne`). So by the time the storage is listed, no confirmation still reads an object of an abandoned upload.
- **The age of an object is the storage's last-modified time**, and a copy is written when it is made (proven against Garage by `TestACopyIsListedWithTheTimeItWasMade`). The object of a file being confirmed is therefore young while the confirmation commits and named once it is old, so a confirmed file is never touched. An old object that only an unconfirmed file names is the object of an abandoned upload, and is removed.
- **The cleanup stops at the first failure** and returns it; River tries again (five attempts) and the next hourly run starts over, so nothing is lost.
- **Every periodic job runs when the worker starts** (`RunOnStart`), so a worker restarted more often than its interval still runs it; that is how "tolerates a missed run" is met. No uniqueness option: a second run changes nothing, so a duplicate after a new election or a retry overlapping the next run is harmless.
- **Timeout of ten minutes** (River's default is one): the job lists the whole bucket and removes objects one by one. Each call to the storage keeps its own ten-second timeout, and at the shutdown the job is cut and made available again like any other.
- **No key in logs or spans, for every call of the adapter**, not only the cleanup's: a URL is cut to the address of the storage (the path holds a key; Garage's continuation token is the last key in base64), and anything shaped like an object key (32 hexadecimal digits, as `domain.NewObjectKey` makes one) becomes `[key]`, which also covers a key the storage repeats in its answer. The API's own 5xx lines for `Inspect` and `Copy` lose the key too.
- **No test runs the cleanup on the bucket of development.** A cleanup there with a temporary database would remove objects the development database names, so the tests of the `worker` command give it a stand-in storage (`standInStorage` in `cmd/api/worker_test.go`), the use case is tested on the in-memory fakes, and the `s3` tests list the bucket but look only at their own keys.
- **No line for a successful run**, as Telemetry says of a successful job; a failed one is logged by River (`Job errored`), without a key.

### What was checked

- Use case, in-memory with a fixed clock (`files/application/clean_up_storage_test.go`): an unnamed object kept for a day and removed after; a confirmed file and its object untouched after 30 days; an abandoned upload removed with its object; an upload of a day kept and still confirmable; the object of an abandoned upload written after the start removed at the next run; the object left behind by a failed removal after a confirmation removed; a state with named, unnamed, young and old objects over several pages; a second run that changes nothing; an unlistable storage (files removed, objects kept for the next run); a database that does not answer (nothing removed); a failed removal retried by the next run.
- `s3` against Garage: the listing over several pages with each key once and its time in UTC; an object removed is listed no more; a copy listed with its own time. Against unreachable or stand-in storages: the error of `Inspect`, `Copy`, `Remove` and `List` names the storage and no key; a later page that fails names neither the key nor the token; a refusal that repeats the key does not; the spans of a failed call carry no key.
- `postgres` as the API role: abandoned files removed and the young, the confirmed and the one started exactly at the instant kept; the keys of confirmed files found among those given; the removal waits for a confirmation that holds the file and keeps it. Without migration `00020` the removal fails with `permission denied`.
- `river`: the arguments marshal to `{}`; the light queue and a timeout above River's default; the worker runs the use case through `rivertest`.
- `platform/jobs`: a periodic job runs when the worker starts and again after its interval, in the queue its arguments name; the test reads when River saved each job, because when a job runs also depends on how soon a busy worker fetches it.
- `cmd/api`: the worker starts on the database, the sender of e-mail and a storage, and stops with status 0; it removes an old object of the storage and keeps a young one as soon as it starts; it refuses to start without the database, the sender of e-mail or the file storage, naming each variable. The test of sensitive values gained a scenario in which the removal of an object fails with a cut connection: neither the logs (with River's `Job errored` line) nor the spans carry the key; it fails when the adapter stops cutting the key.
- `go build`, `go vet`, `go mod tidy -diff`, `go test -count=1 ./...` of `apps/api`, `task api:lint`, `task api:generate:check`, `task api:spec:check`, `task api:spec:lint`, `task api:lists:check`, `task versions:check`, before and after `feat/login` with ticket 02 was brought in. Ticket 02's scenario of sensitive values loads the worker's configuration, so it now gives it a stand-in storage too. In the whole runs, tests of `platform/database`, `cmd/api` and `cmd/seed` failed at the 30-second limit of creating or dropping a database under load; each passed when run again alone, as Tests in `CODING_STANDARDS.md` says to do.

### What was not checked

- `task dev` and `task worker` were not run (ports of the main checkout), and the worker was never run against the Garage of development, on purpose (see above).
- Backblaze B2: the listing, its paging, the time of a copy and its continuation token were tried on Garage only (the deploy effort, ADR-0020).
- A bucket of thousands of objects was not measured; the ten-minute timeout is an estimate.
- Nothing of the frontend changed, so `task web:check` was not run.

### Found, and left for a decision

- **One bucket in development for several databases.** The development worker removes, a day later, the objects only the end-to-end database or a test database names (the logo of the end-to-end setup, the leftovers of a killed test). Harmless today, since the end-to-end database is recreated on every run; a bucket or a key prefix per database would change it. Recorded in Files.
- **The rule for keys is a shape.** Any run of 32 hexadecimal digits in an error or a span of the storage becomes `[key]`, so an identifier of that shape from the storage (a request identifier, say) would be hidden too. If the form of an object key changes, `keyless.go` changes with it.
- `docs/machine-tools.yaml` and `docs/providers.md` need no change.
