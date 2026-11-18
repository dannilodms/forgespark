---
status: accepted
---

# Files in Garage, reached directly by the browser

Files live in S3-compatible storage. The server is Garage, which replaces MinIO, and the backend speaks only the S3 API, through the AWS SDK for Go v2. The browser sends and fetches the bytes straight to and from the storage, with short-lived signed URLs that the API issues.

- **MinIO is no longer an option.** Its community repository was archived in April 2026 and gets no fixes. The official images left Docker Hub, and the last one published has a known privilege-escalation flaw. Where an earlier document names MinIO, it means the file storage.
- **Only the S3 API.** The backend uses nothing that one server alone offers, such as bucket notifications. Changing the server is a change of configuration.
- **Garage in development and staging, a provider in production.** In production the files are in Backblaze B2 (ADR-0020). What this decision says about signed URLs, confirmation, parts and keys holds for both.
- **The bytes never pass through the API.** A video or an audio file can have gigabytes. Through the API, an upload would hold a request for minutes, and it would hit the body size limit and the drain deadline of the shutdown. The use case checks the permission and returns a signed URL.
- **The browser confirms an upload.** The API then checks the object in the storage, its size and its type, and only then is the file available. A periodic job (ADR-0016) removes uploads that were never confirmed.
- **A confirmed file is out of reach of its upload URL.** Confirming copies the object inside the storage, with `CopyObject` of the S3 API, to a new random key that no URL was signed for, records that key on the file and removes the object the upload was sent to. Whoever still holds the upload URL can send to it until it expires, and what they send changes nothing the system reads.
- **A large file goes in parts**, each sent with its own signed URL.
- **One private bucket, random keys.** An object key never contains the file's name or an identifier.
- **The storage has an address of its own.** In development and staging Caddy forwards a host name to Garage. In production the browser reaches the provider's own host name. Either way the bucket accepts cross-origin requests from the app's origin only.

## Considered Options

- **SeaweedFS.** Apache licence, object versioning and a panel. It has more processes to operate and depends almost entirely on one maintainer.
- **Silo, the community fork of MinIO.** The same server, with its panel back. Its future depends on a single sponsor.
- **RustFS.** Close to MinIO, with a panel. The youngest of the group, and its support for signed upload forms was not confirmed.
- **AIStor Free, MinIO's own replacement.** Proprietary and limited to one node.
- **minio-go as the client.** A simpler API and still maintained, but it belongs to the company that archived the server and its command-line tool.
- **Uploads and downloads through the API.** The storage would never face the internet, but a file of gigabytes would occupy the API for minutes.
- **The storage telling the API that an upload ended.** Garage has no bucket notifications, and each server that has them does it its own way.

## Consequences

- Garage has no object versioning, no bucket notifications and no official web panel, and a small team maintains it (v2.5.0 when this was decided). Its licence is AGPLv3. It runs as a separate service, so the licence does not reach this repository's code.
- Without versioning, an object deleted or overwritten in Garage is gone. That is accepted in development and staging, which hold no real file. Production keeps its files in a provider with versioning (ADR-0020).
- The AWS SDK for Go v2 was tried against Garage with the setup (Garage v2.5.0, `service/s3` v1.114.10), and the tests of `apps/api/internal/files/adapters/s3` run against it. What it needed:
  - **Path-style addresses** (`UsePathStyle`). With the bucket in the host name, Garage reads the bucket from the path and takes the key for it.
  - **No checksum setting.** The defaults of the SDK (`WhenSupported` for requests and for responses) work against Garage, a `PutObject` with a body that cannot be read twice included, and a signed `PUT` carries no checksum, so the browser sends none. A later version of the SDK that changed this would fail those tests.
  - **A second client for signing**, whose endpoint is the public host name (`FILE_STORAGE_PUBLIC_URL`), beside the one the API calls the storage with (`FILE_STORAGE_URL`). Caddy passes the `Host` header on to Garage, and the signature covers it.
  - **The declared type and size are signed.** Given the `Content-Type` and the `Content-Length`, the SDK signs both (`X-Amz-SignedHeaders=content-length;content-type;host`), and Garage refuses with 403 a `PUT` of another type or size. An expired URL is refused with 400, `Date is too old`, where AWS answers 403.
  - **The cross-origin rule** is set with `PutBucketCors`, signed by curl in a step of the development Compose file. Garage answers the preflight of another origin with 403, so a browser sends nothing; a request sent without a browser is not refused for its `Origin`.
  - **`CopyObject`** works with the defaults of the SDK, the source named as the bucket and the key. A `PUT` to the upload URL after the copy and the removal writes under the key of the upload again, and the copy keeps its bytes: the tests of the `s3` adapter send one.
- The storage is reachable from the internet. What protects a file is the signature and the expiry of each URL: whoever holds a signed URL can use it until it expires, so the expiry is minutes. An upload URL stops reaching the file once the file is confirmed.
- The copy is made in the transaction that checks the file and locks it, before the file names the new key, so the record never names a key with no object, and a second confirmation of the same file waits and finds it confirmed. It does not read the storage again: no URL reaches the object of a confirmed file. The object the upload was sent to is removed after the commit.
- **When the copy succeeds and the removal fails, the confirmation stands** and the operation succeeds. The file names the copy, which no URL reaches, so the object left behind is read by nothing, and its URL stops working minutes after the upload started. Failing the operation would refuse a file that is already confirmed, for an object it no longer needs, and a second try could not find that object either. The failed call shows in the span of the storage. A copy whose transaction does not commit is left behind the same way. No file names either of them, so the periodic job that removes uploads never confirmed also finds the objects no file names, by listing the bucket (`ListObjectsV2`), not from the files alone. Until that job exists they stay in the bucket, as an abandoned upload does.
- `CopyObject` copies up to 5 GB in one call. A larger file, which arrives in parts, is copied in parts (`UploadPartCopy`), chosen with the first upload in parts.
- A signed upload URL bounds the size of what is sent only where the storage checks the signed `Content-Length`: Garage does, and Backblaze B2 is not tried. The API checks the size when the upload is confirmed and removes an object over the limit.
- A signed URL is signed for the public host name of the storage, so the backend has to know that name.
- The file storage stays an optional part of the system health (ADR-0012).
- The library that uploads in parts from the browser is chosen with the first upload screen.
- How a document reaches a person outside the system, as an attachment or as a link, is not decided.
