# 10: Periodic cleanup of the storage

**What to build:** Every hour the worker removes from the file storage the uploads started more than 24 hours ago and never confirmed, and the objects that no confirmed file names, so abandoned uploads and the leftovers of a failed removal do not stay for ever. Rules: Files and Jobs in `CODING_STANDARDS.md`, ADR-0017.

**Blocked by:** 01 (River and the worker command).

**Status:** ready-for-agent

- [ ] The `s3` adapter lists the bucket with the last-modified date of each object and removes an object by key; its tests run against Garage under keys of their own.
- [ ] A use case removes each object older than 24 hours that no confirmed file names, and marks as abandoned or removes the file records of uploads started more than 24 hours ago and never confirmed; a confirmed file is never touched.
- [ ] A periodic job declared in the worker runs the use case every hour, in the light queue with a longer timeout; it tolerates a missed run.
- [ ] A second run over the same state changes nothing; use case tests cover objects named, unnamed, young and old, and the second run, with a fixed clock.
- [ ] The job's arguments carry no personal data, and the logs carry no key of an object.
