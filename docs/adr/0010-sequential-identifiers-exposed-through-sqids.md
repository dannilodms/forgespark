---
status: accepted
---

# Sequential identifiers, exposed through Sqids with a type prefix

Every aggregate is identified by a sequential integer taken from a database sequence. That integer never leaves the backend as such. Outside, in the REST API, in MCP tools and in URLs, an identifier is a type prefix followed by the Sqids encoding of the integer, at least 8 characters long: `prj_Xk9fQ2mB`. Each aggregate type has its own prefix and its own shuffled Sqids alphabet.

- **Short identifiers suit the consumers.** A person can read one aloud, and it costs an agent fewer tokens than a 36-character UUID.
- **The prefix carries the type.** An inbound adapter refuses an identifier with the wrong prefix before decoding it. The alphabet alone does not give that: an identifier decoded with another type's alphabet can yield a valid number of a different record.
- **The identity exists before the aggregate.** The repository reserves the next number from the sequence, and the use case creates the aggregate with it. An in-memory repository uses a counter. A reserved number that is never saved leaves a gap, which does not matter.
- **Alphabets and prefixes are constants in the code.** Every environment and every test produce the same identifiers, and no configuration value can be lost and invalidate them.
- **Encoding happens at the edge.** Inbound adapters decode on the way in and encode on the way out. The domain, the database and foreign keys use the integer.

Sqids is not a security measure, and its documentation says so: it is not encryption, and anyone who knows the alphabet can decode the numbers. Nothing here relies on identifiers being unguessable. Every operation requires authentication, and the use case checks whether the actor may touch that record.

## Considered Options

- **UUIDv7.** Time-ordered, native in PostgreSQL 18, and generated without a round trip to the database. It also hides how many records exist, which a decoded Sqid does not. Not chosen: the maintainer preferred short identifiers over those properties.
- **TypeID, a UUIDv7 with a type prefix.** The type is visible, but the identifier is longer still and the Go library's second version is alpha.
- **Alphabets in secret configuration.** Keeps them out of the repository, but losing or changing a value invalidates every identifier ever handed out, and environments would disagree.
- **The database assigning the number on insert.** One round trip fewer, but the aggregate would exist without an identity and its identifier type would need a "not assigned yet" state.
- **No prefix.** Shorter, but a wrong-type identifier could reach another record instead of failing.

## Consequences

- A prefix or an alphabet can never change once identifiers have been handed out: every link, bookmark and reference stored by an agent would break.
- Creating an aggregate needs the database, to reserve the number.
- Anyone who reads the repository can decode an identifier and see roughly how many records of a type exist.
- Sequences belong to one database. Merging data from two instances would collide, which a single instance does not face.
