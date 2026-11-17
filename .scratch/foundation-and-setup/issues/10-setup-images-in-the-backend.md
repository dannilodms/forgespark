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

- [ ] With a valid code, starting an upload answers a signed URL, and a `PUT` to it from the origin of the app stores the object. A `PUT` from another origin is refused by the storage.
- [ ] A declared type outside the three, or a declared size over 2 MB, is refused when the upload starts.
- [ ] An object that is larger than declared and over the limit, or whose first bytes are not a PNG, a JPEG or a WebP, is refused at the confirmation and removed from the storage.
- [ ] Concluding with a logo and a photo saves the consultancy and the administrator with them. Concluding without them still works.
- [ ] Concluding with the identifier of a file that was never sent, or that does not exist, answers 422 naming the field, and saves nothing.
- [ ] A code that is not valid starts no upload.
- [ ] The object key is random and holds neither the name of the file nor an identifier.
- [ ] The identifier of a file in a response has its prefix, at least 8 characters after it, and never the integer.
- [ ] The adapter tests run against Garage and cover the signed URL for the public host name, the confirmation and the removal.
- [ ] The AWS SDK for Go v2 works against Garage with path-style addresses and the checksum settings it needs. What had to be set is noted in ADR-0017.
- [ ] The spec, the generated code and the lint are up to date, and the pipeline starts Garage for the adapter tests.
