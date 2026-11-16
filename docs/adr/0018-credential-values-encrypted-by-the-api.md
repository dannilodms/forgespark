---
status: accepted
---

# Credential values encrypted by the API, with a key from configuration

The value of a credential (a password, an API key, a private key, a certificate) is encrypted by the API before it reaches PostgreSQL. The key comes from the deploy configuration and is never in the database, so a dump, a backup or a read of the table reveals no value.

- **Encryption in the Go binary.** AES-256-GCM from the standard library. No extension and no new container.
- **Each row names its key.** The identifier of the key is stored beside the encrypted value. The configuration holds the key that encrypts and the older keys that still decrypt, so a key can be changed without stopping the system.
- **A value is bound to its credential.** The identifier of the credential is authenticated together with the value, so an encrypted value copied to another row does not decrypt.
- **A file is a value.** A credential that is a file, such as an SSH key or a `.pfx` certificate, is encrypted the same way and stored in PostgreSQL, under a small size limit. It goes through the API and never to the file storage, the one exception to ADR-0017.
- **Only the use case that reads a value decrypts it.** That use case already records the read in the audit log (ADR-0008, ADR-0012).

## Considered Options

- **pgcrypto.** An extension the official PostgreSQL image already ships. The key travels to the database server on every query and can show up in its logs.
- **A separate vault (OpenBao, Infisical).** Stronger separation, with one more service to operate, to back up and to unseal on every restart.
- **Plain text.** Protected only by the disk and by access to the database. A leaked dump would expose every credential of every client.
- **Credential files in the file storage.** One path for every file, but without the encryption of the application, and a backup of the files would expose them.

## Consequences

- Losing the key loses every credential value. The key has to be kept apart from the database dumps, and the backup routine, not decided yet, has to say where.
- The key is one more secret in the deploy configuration. How a secret reaches a container is decided with the deploy.
- The API process holds the key in memory. Whoever controls the API reads the values.
- A value cannot be searched or compared in SQL.
- The upload of a credential file is a request body, under the body size limit of ADR-0009.
- The routine that encrypts the stored values again under a new key is not written yet. It is written with the first change of key.
