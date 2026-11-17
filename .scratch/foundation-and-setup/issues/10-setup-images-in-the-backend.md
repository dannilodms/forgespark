# 10 - Images of the setup in the backend

Status: ready-for-agent
Blocked by: 06, 07

Read `../spec.md`, ADR-0017, ADR-0010 and, in `CODING_STANDARDS.md`, Files and API conventions.

## What to build

A caller with a valid setup code sends a logo for the consultancy and a photo for the administrator straight to the file storage, and concluding the setup keeps them.

- Garage in the development Compose file, with a step that creates the one private bucket, its key and its cross-origin rule, which accepts the origin of the app only. A host name of its own in the `Caddyfile` leads to it.
- The `files` area, with two use cases: start an upload and confirm an upload. Starting creates the record of the file with a random object key and returns a signed URL that expires in minutes. Confirming checks the object in the storage and marks the file as confirmed.
- The rules of an image: a PNG, a JPEG or a WebP of at most 2 MB. Starting refuses another declared type or a larger declared size. Confirming reads the size of the object and its first bytes, and removes an object that is over the limit or is not one of the three types.
- The `s3` adapter on the AWS SDK for Go v2, using only what the S3 API defines, with the URL signed for the public host name.
- The identifier encoder in `platform`, with the prefix and the alphabet of the file. The identifier of a file leaves the backend as its prefix and its Sqids encoding, and a wrong prefix is rejected on the way in.
- The use case that starts an upload of the setup: it checks the setup code and asks the `files` area to start an upload for the logo of the consultancy or for the photo of the administrator.
- `POST /api/setup/uploads`, with the setup code in the body, answering 201 with the identifier of the file, the signed URL and when it expires. It falls under the limit of ticket 07.
- Concluding the setup accepts the identifier of a logo and of a photo. It confirms them after validating the data and before its transaction, and the consultancy and the administrator keep them.
- A file started under a setup code that has since expired is still accepted when the setup is concluded with a new code.

The area has no REST operation of its own and no download yet. The job that removes uploads never confirmed comes with River, in the second effort.

## Acceptance criteria

- [x] With a valid code, starting an upload answers a signed URL, and a `PUT` to it from the origin of the app stores the object. A `PUT` from another origin is refused by the storage.
- [x] A declared type outside the three, or a declared size over 2 MB, is refused when the upload starts.
- [x] An object that is larger than declared and over the limit, or whose first bytes are not a PNG, a JPEG or a WebP, is refused at the confirmation and removed from the storage.
- [x] Concluding with a logo and a photo saves the consultancy and the administrator with them. Concluding without them still works.
- [x] Concluding with the identifier of a file that was never sent, or that does not exist, answers 422 naming the field, and saves nothing.
- [x] A code that is not valid starts no upload.
- [x] The object key is random and holds neither the name of the file nor an identifier.
- [x] The identifier of a file in a response has its prefix, at least 8 characters after it, and never the integer.
- [x] The adapter tests run against Garage and cover the signed URL for the public host name, the confirmation and the removal.
- [x] The AWS SDK for Go v2 works against Garage with path-style addresses and the checksum settings it needs. What had to be set is noted in ADR-0017.
- [ ] The spec, the generated code and the lint are up to date, and the pipeline starts Garage for the adapter tests.

## Comments

### What was built

- **Garage in development.** `compose.yaml` runs `dxflrs/garage:v2.5.0` (tag in `deploy/compose.base.yaml`) on one node, with `server --single-node --default-bucket`: on every start it creates, if missing, the layout, the bucket `FILE_STORAGE_BUCKET` and the key `FILE_STORAGE_ACCESS_KEY_ID`/`FILE_STORAGE_SECRET_ACCESS_KEY` with read, write and owner rights. Its configuration is a `configs` entry with its content (no root domain, so addresses are path-style only; it listens on IPv4), its data the volume `garage-data`, its S3 API published at `127.0.0.1:${GARAGE_PORT}` (3900). The image has no shell, so its health check is `/garage bucket info`. A one-shot service, `garage-cors` (`curlimages/curl:8.23.0`), signs `PutBucketCors` with `--aws-sigv4`: the origin of the app (`APP_ORIGIN`) may `PUT` and `GET`, with the header `content-type`, `ETag` exposed. Caddy depends on it completing, which is what lets `docker compose up --wait` accept a service that exited.
- **The host name of the files.** `http://files.localhost` in `deploy/Caddyfile` forwards to `{$FILES_UPSTREAM:garage:3900}`, keeping the `Host` the signature covers.
- **The identifier encoder**, `internal/platform/identifier`: a `Type` with its prefix and its shuffled alphabet (`File`, `fil_`), Sqids with at least 8 characters. `Encode` refuses a number that is not positive; `Decode` refuses another prefix before decoding, and an encoding the encoder would not have written (more than one number, not canonical). `Pattern()` gives `^fil_[0-9A-Za-z]{8,}$` to the contract.
- **The `files` area.** `domain`: the file (identifier, random object key of 128 bits in hexadecimal, purpose `consultancy_logo` or `user_photo`, the declared type and size, when its upload started and when it was confirmed), `DeclareImage` (a PNG, a JPEG or a WebP of 1 byte to 2 MiB, naming `media_type` and `size`), `File.Confirm` (the stored size against 2 MiB and the first 12 bytes against the three signatures; a second confirmation keeps the first time), and five broken rules: `file_type_unsupported`, `file_too_large` (`max_size`), `file_empty`, `file_not_found`, `file_not_uploaded`. `application`: `StartUpload` (checks the declaration, saves the file in a transaction, signs the URL for five minutes) and `ConfirmUpload` (loads the file with a lock, refuses one of another purpose as not found, inspects the object, removes it when the file refuses it), with `Storage`, `FileRepository`, `Clock` and `Transactions`, and `vocabulary.go` for another area. `applicationtest`: in-memory implementations, the clock and the transactions copied from the consultancy area.
- **The adapters.** `adapters/s3`: the AWS SDK for Go v2 (`service/s3` v1.114.10), two clients with path-style addresses, one for the calls of the API at `FILE_STORAGE_URL` and one that signs at `FILE_STORAGE_PUBLIC_URL`; `SignUpload` signs a `PUT` with the declared `Content-Type` and `Content-Length`; `Inspect` is a `HEAD` and a ranged `GET` of the first 12 bytes; `Remove` a `DeleteObject`; a 10-second timeout on each call; the calls traced through `smithyoteltracing`. `adapters/postgres`: `FileRepository` on `files.sql` (`LoadFile` is `FOR UPDATE`), with its sqlc entry.
- **The migration** `00006`: the table `files` (a unique object key, `CHECK`s on the purpose and the type), its grants, and the columns `consultancy.logo_file_id` and `users.photo_file_id`, foreign keys to `files`.
- **The setup.** `StartSetupUpload` in the consultancy area checks the setup code (without a lock) and calls `StartUpload`. `ConcludeSetup` takes `TypedConsultancy.LogoFileID` and `TypedUser.PhotoFileID`; after the fields and a check of the code without a lock, it confirms each image chosen and names `consultancy.logo_file_id` or `administrator.photo_file_id` with the rule refused; then its transaction, as before, saves them with the consultancy and the administrator. `findWorkingSetupCode` is the one check of a typed code that opening, starting an upload and concluding share.
- **REST.** `POST /api/setup/uploads` (`start_setup_upload`), body `StartSetupUploadRequest` (`code`, `purpose`, `media_type`, `size`, the `enum` and the `maximum` taken from the files domain), answers 201 with `SetupUpload` (`file_id`, `upload_url`, `expires_at`), 422 `setup_code_invalid` and the three rules of a declaration, 429 of Caddy. `conclude_setup` takes `logo_file_id` and `photo_file_id` with the pattern of a file, decodes them (one that decodes to no number is `file_not_found` on its field), and lists the four rules of an image under 422. `apps/api/openapi.yaml` regenerated.
- **Configuration.** Six variables the API requires: `FILE_STORAGE_URL` and `FILE_STORAGE_PUBLIC_URL` (origins), `FILE_STORAGE_REGION`, `FILE_STORAGE_BUCKET`, `FILE_STORAGE_ACCESS_KEY_ID` and `FILE_STORAGE_SECRET_ACCESS_KEY`, a `config.Secret` that prints as a placeholder. `.env.example` gains them, `GARAGE_PORT` and `GARAGE_RPC_SECRET`. `serveAPI` builds the storage and passes it to `mountREST`.
- **The tests.** Use cases on the in-memory implementations (`files/application`, `consultancy/application`); the `postgres` adapters on pgtestdb; the `s3` adapter on the Garage of development; the binary through HTTP (`cmd/api/setup_uploads_test.go`); Caddy in front of Garage (`caddy_test.go`); a new scenario of the test of sensitive values.
- **The documents.** ADR-0017 (what the SDK needed, and the size of a signed upload); `CODING_STANDARDS.md` (Layout, Identifiers, Spans, Files, Tests, Development); both READMEs (the variables, Garage, `files.localhost`, `task services`, `task api:test`); a comment in `ci.yml`.

### Decided while building

- **Garage creates its bucket and key itself.** Since v2.3.0, `--default-bucket` does it from `GARAGE_DEFAULT_*`, which `compose.yaml` fills with the variables of the API, so the key in `.env.example` is the key Garage has. The CORS rule is the only step, and curl signs it, so no AWS CLI image is pulled. The admin API is not turned on.
- **The configuration of Garage lives in `compose.yaml`**, as the one of PostgreSQL does, not as a file of `deploy/`: no worktree creates Garage again by mounting its own path, so an upload in flight of another worktree survives. Staging, which also runs Garage, will need its own.
- **The URL is signed for five minutes**, and the declared `Content-Type` and `Content-Length` are signed with it.
- **2 MB is 2 MiB**, 2097152 bytes, written in the contract as `maximum`.
- **The setup code is checked, without a lock, before the images are confirmed.** Otherwise anyone could make the API read and remove objects with a wrong code. The transaction checks it again, locked.
- **Confirming is checked again every time.** A conclusion that failed after its images were confirmed is tried again with the same images; the second confirmation reads the object again and keeps the first time.
- **An identifier with the right prefix that decodes to no number is `file_not_found`**, on its field, as a file that does not exist; the contract answers a wrong prefix with `invalid`. The adapter answers it before the use case runs, so other fields are not named in the same answer.
- **The image rules belong to the files area**; the consultancy area re-exports them in `application/files.go`, because its REST adapter may reach another area only through its own `application`.
- **The in-memory clock and transactions are copied** into the files area, as Tests in `CODING_STANDARDS.md` puts them in each area.
- **The storage variables are required by the API**, like `DATABASE_URL`, rather than optional with a `not_configured` part: the API serves the setup, which uploads.
- **`purpose` is `user_photo` for the administrator**, the same purpose ticket 12 uses for a new administrator.
- **No `Location` on the 201**: there is no resource of a file to point at until the files area has operations of its own.
- **The calls to the storage are traced** through the SDK's own OpenTelemetry adapter; `otelaws` is deprecated and would bring four other service clients.

### What was checked

- **The use cases**: starting saves the file as declared and unconfirmed, signs a URL for its key, the declared type and size, for five minutes; keys come only from the random source; each of the three types passes; a GIF, an SVG, a PDF, `IMAGE/PNG`, a type with parameters and none are refused; 2 MiB passes and one byte more is refused with `max_size`; 0 and -1 are `file_empty`; a type and a size refused are named together; another purpose saves nothing; a refused save signs nothing. Confirming: each type confirmed at the time of the clock; 2 MiB passes; 2 MiB + 1 refused, removed, unconfirmed; a PDF, text, a RIFF that is no WebP, an empty object and the first four bytes of a PNG refused, removed; never sent is `file_not_uploaded`; 99, 0 and -1 `file_not_found`; a photo given for a logo `file_not_found`, not removed; a second confirmation keeps the first time; a storage that does not answer confirms nothing. Through the setup: with a valid code the file and its URL; a wrong, invalid, expired or ended code starts nothing and signs nothing; the rules of a declaration named; a wrong code answered before them; concluding with a logo and a photo keeps both and confirms them; without them neither; never sent and not existing named on their fields with nothing saved; a photo is not taken for a logo; an image too large and a file that is no image named and removed; an image sent under an expired code accepted with a new one; images confirmed by a conclusion that failed accepted on the next; a wrong code confirms and removes nothing; a field refused confirms nothing.
- **The `s3` adapter against Garage**: the URL has the scheme and host of `FILE_STORAGE_PUBLIC_URL`, the bucket in its path and `X-Amz-Expires=300`; a `PUT` with that `Host` stores the object, and `Inspect` reads its size and first 12 bytes; a larger, a smaller and another type are answered 403 and store nothing; an expired URL is answered 400 and stores nothing; the preflight of `APP_ORIGIN` is allowed for `PUT` and `content-type`, the one of another origin answered 403; `Inspect` of 3 MiB reads its size and 12 bytes, of 4 bytes all of them, of an empty object none, of a missing key nothing; `Remove` leaves nothing, and of a missing key is no failure; a storage on port 1 fails with an error that does not carry the secret key.
- **What the SDK needs**, with a probe written and removed: without path-style, `PutObject` is answered `Invalid create bucket XML query` and `GetObject` `NoSuchBucket` with the key as the bucket; with it, `PutObject` with a seekable and an unseekable body and a ranged `GetObject` pass with the default checksums; the signed URL is `X-Amz-SignedHeaders=content-length;content-type;host` with no checksum parameter.
- **The `postgres` adapters**: a file loaded as saved, a confirmation loaded with its time, nothing for an unknown identifier, the lock (`55P03` from a second transaction), nothing kept from a transaction that fails, two files never share a key (`23505`); the consultancy and the user saved with the file of their logo and photo, and `23503` for a file that does not exist; the full rows of both with the new columns empty.
- **The binary**: 201 with exactly `file_id`, `upload_url`, `expires_at`; the identifier `fil_` and 8 or more characters, decoding to the row saved, and the integer absent from the body; the URL under `FILE_STORAGE_PUBLIC_URL` and the bucket; the expiry five minutes ahead. Two images sent to their URLs and the setup concluded with them: the consultancy and the administrator point at confirmed files of their purposes. A wrong code starts no upload and saves no file. A GIF, over 2 MiB, no byte, another purpose, no type, a name of file and a fraction answered 422 at their field. Never sent and not existing answered with the exact body and nothing stored changed. Another prefix, no prefix, too short (`invalid`) and an identifier that decodes to no number (`file_not_found`). An image too large and a file that is no image, stored past the URL, named and removed from Garage, nothing saved. The variables missing name the six. The spec: three `POST` of the setup each declare the 429 of Caddy.
- **Through Caddy**, a container of the image on the network of the services: the preflight of the app allowed with its origin, of another origin 403, and the `PUT` to the signed URL, with `Host: files.localhost`, stored (200).
- **Sensitive values**: a scenario starts two uploads, sends one image, concludes with a never-sent photo (422) and with the logo (204), then fails a start with a 5xx (`REVOKE INSERT ON files`); the logs and spans, the spans of `HeadObject` included, carry neither the secret key of the storage, nor the signature of either URL, nor either setup code. The first scenario now carries a secret key of the storage too.
- **Configuration**: the six read, each named when missing, the two origins refused when not one, the secret absent from `%v`, `%+v` and `%#v`, `migrate` and `issue-setup-code` need none of them.
- **The identifier**: 1 to `MaxInt64` round-trip with `fil_` and 8 or more characters, two numbers two identifiers, 0 and negatives refused, eight wrong forms refused, the pattern, and the encodings of 1 and 1000 pinned.
- **The Compose file** in a project of its own (`fs10trial`, removed with its volume): Garage healthy with its bucket and key, the step setting the rule, a restart keeping them, the preflights through Caddy; `up --wait` failing (status 1) with the step alone and passing once Caddy waits for it. Then the shared services from this worktree: PostgreSQL left running, Caddy and Garage created.
- **The versions**: `dxflrs/garage:v2.5.0` and `curlimages/curl:8.23.0` exist and are the latest releases; the Go modules are the latest (`go list -m -versions`); `task versions:check` passes.
- **The workflow file**: actionlint 1.7.12 from its image, no finding.
- **`task check`** passes on the branch. The first run failed in `internal/platform/database` and `cmd/api` only, every failure a `CREATE DATABASE` or `DROP DATABASE` past its 30 seconds, or a readiness of 503, under load; the two packages passed run alone, and the next `task check` passed whole. `feat/foundation-and-setup` had nothing new to merge.

### What was not checked

- **The pipeline.** Nothing was pushed: the backend job never started Garage on the runner. Garage was tried on Docker Desktop only; its configuration binds IPv4 so that a runner without IPv6 in its containers does not stop it, untried. This is why the last criterion is left open.
- **The client generated by orval** (`apps/web`) is stale against the new spec; ticket 08 owns it and the merge regenerates it, as the run asked.
- **A browser.** No browser sent a file: the preflights and the `PUT` were sent by Go with the headers a browser sends.
- **A request without a browser from another origin** is not refused: Garage refuses the preflight, and a browser then sends nothing. A `PUT` sent by a script with any `Origin` and a valid URL is stored, as the signature allows.
- **The removal when the storage fails to remove.** It answers 500, read from the code; no test makes `DeleteObject` fail.

### Found, and left for a decision

- **`Location` on the 201** of `start_setup_upload`, which API conventions ask of a creation, waits for an operation that reads a file.
- **A confirmed image can be replaced** through its signed URL within its five minutes, after it was confirmed. Garage checks the signed size and type, so it stays an image of the same type and size; its bytes are not checked again until another confirmation.
- **Images of an abandoned setup stay in the bucket**, the tests of the binary included, until the cleanup job of the second effort.
- **The health of the storage** (an optional part, ADR-0012) is not built: the API requires the variables and does not probe Garage at the start.
