---
status: accepted
---

# PostgreSQL backed up by pgBackRest, and production files in Backblaze B2

Two things in production cannot be lost: the database and the files. PostgreSQL stays a container on the production server and is backed up continuously by pgBackRest to a bucket at another provider. The files leave the server: in production they live in Backblaze B2, and Garage (ADR-0017) remains the file storage of development and staging only.

- **Continuous backup.** pgBackRest archives the write-ahead log as it is written, and takes a full backup every week and a differential one every day. The database can be restored to any moment of the last 30 days.
- **At another provider.** The backups go to a bucket at Backblaze B2, not at the provider of the server. Losing the server, or the account that owns it, does not take the backups with it.
- **Encrypted before they leave.** pgBackRest encrypts the backups with a passphrase from the deploy configuration. The passphrase and the key that encrypts credential values (ADR-0018) are never in the bucket.
- **The PostgreSQL image is built here.** It is the official image plus pgBackRest, which has to run beside the server to archive the log. This is the change of image that ADR-0015 said would need a decision.
- **A timer on the server starts the backups.** The API and the worker have no part in it: their database role cannot read the files of the server.
- **A restore is tested every month.** A scheduled workflow creates a server from the same cloud-init file (ADR-0019), restores the latest backup there, checks it and destroys the server. Production data is never restored on a GitHub runner, whose logs are public (ADR-0021).
- **A backup that stops is noticed.** Each backup ends by signalling a monitor outside the server, which warns by e-mail when the signal does not come (ADR-0012).
- **Production files are in a second bucket**, with object versioning. A file deleted or overwritten can be recovered for 30 days. The browser sends and fetches the bytes straight to and from Backblaze B2, with the signed URLs of ADR-0017.
- **The dump before a migration stays** (ADR-0015). The last three are kept on the server.

## Considered Options

- **A managed PostgreSQL.** The provider would take the backups and do the restore, for a monthly fee. The maintainer chose to keep the database on the server and own the restore.
- **A daily `pg_dump` only.** The simplest routine, which can lose a day of data.
- **WAL-G.** One binary and a shorter configuration, with fewer integrity checks and less documentation.
- **Garage on the production server, copied to another provider.** It keeps one file storage in every environment, but the server's disk would have to hold every file, and Garage has no object versioning.
- **Cloudflare R2.** The best documentation and free egress, but no object versioning: only a retention rule on the bucket.
- **OVHcloud Object Storage or Scaleway.** Both document everything this system uses, for a little more. They are the fallback if the tests below fail.
- **Hetzner Object Storage.** One account for everything, but a fixed monthly fee, and the backups would sit at the provider of the server.
- **Providers with a fixed fee or a one-terabyte minimum** (DigitalOcean, Akamai, UpCloud, Storj, IDrive e2, Wasabi, Vultr). Several times the price at this size.
- **The restore test on the production server.** No new server, but it uses the disk and the CPU of production and does not prove a recovery after the server is lost.
- **No automatic restore test.** One workflow less, and a backup that was never restored is not a guarantee.

## Consequences

- Three things were read in documentation and not tried. The work that builds the deploy tries them, before production holds a file or a backup. The AWS SDK for Go v2 sends a checksum by default since 2025; Backblaze does not say it accepts it, and there are reports that it does not, with a documented setting that turns the checksum off. No provider documents a signed URL for each part of an upload in parts. pgBackRest names no S3-compatible provider in its documentation.
- If Backblaze B2 fails one of those tests, the storage moves to Scaleway or OVHcloud. For the backend that is a change of configuration (ADR-0017).
- The log is archived in segments, when one fills or after a timeout. Losing the server loses what was written after the last segment archived.
- A restore is the maintainer's work. Its steps are in the README, and the monthly test is what keeps them true.
- The browser reaches the file storage at Backblaze's host name in production, and at a host name of Caddy in development and staging. The Content-Security-Policy and the bucket's cross-origin rule differ by environment.
- A Backblaze account keeps its data in one region, chosen when the account is created.
- Backblaze documents its lifecycle rules, which remove the old versions of a file after 30 days, through its own API. Whether they can be set through the S3 API was not confirmed.
- With versioning on, the job that removes uploads never confirmed (ADR-0017) leaves an old version behind for 30 days, and so does every deletion. A file is gone for good only after that.
- The image of PostgreSQL has a Dockerfile, and its version has two parts to update: PostgreSQL and pgBackRest.
