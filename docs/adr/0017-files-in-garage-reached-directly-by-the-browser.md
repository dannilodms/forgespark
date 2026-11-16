---
status: accepted
---

# Files in Garage, reached directly by the browser

Files live in S3-compatible storage. The server is Garage, which replaces MinIO, and the backend speaks only the S3 API, through the AWS SDK for Go v2. The browser sends and fetches the bytes straight to and from the storage, with short-lived signed URLs that the API issues.

- **MinIO is no longer an option.** Its community repository was archived in April 2026 and gets no fixes. The official images left Docker Hub, and the last one published has a known privilege-escalation flaw. Where an earlier document names MinIO, it means the file storage.
- **Only the S3 API.** The backend uses nothing that one server alone offers, such as bucket notifications. Changing the server is a change of configuration.
- **The bytes never pass through the API.** A video or an audio file can have gigabytes. Through the API, an upload would hold a request for minutes, and it would hit the body size limit and the drain deadline of the shutdown. The use case checks the permission and returns a signed URL.
- **The browser confirms an upload.** The API then checks the object in the storage, its size and its type, and only then is the file available. A periodic job (ADR-0016) removes uploads that were never confirmed.
- **A large file goes in parts**, each sent with its own signed URL.
- **One private bucket, random keys.** An object key never contains the file's name or an identifier.
- **The storage has an address of its own.** Caddy forwards a host name to Garage, and the bucket accepts cross-origin requests from the app's origin only.

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
- Without versioning, an object deleted or overwritten in the storage is gone. The backup routine has to cover the files, and it is not decided yet.
- The AWS SDK for Go v2 was not tried against Garage. S3-compatible servers sometimes need adjustments in the client, such as path-style addresses and checksum settings.
- The storage is reachable from the internet. What protects a file is the signature and the expiry of each URL: whoever holds a signed URL can use it until it expires, so the expiry is minutes.
- A signed upload URL cannot limit the size of what is sent. The API checks the size when the upload is confirmed and removes an object over the limit.
- A signed URL is signed for the public host name of the storage, so the backend has to know that name.
- The file storage stays an optional part of the system health (ADR-0012).
- The library that uploads in parts from the browser is chosen with the first upload screen.
- How a document reaches a person outside the system, as an attachment or as a link, is not decided.
