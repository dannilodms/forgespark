# My Software Factory Manager

The operational hub of a software consultancy: one place to manage clients, their projects and their software development requests, from first contact to delivery and maintenance.

## Language

### Credentials and access

**Credential**:
An access the consultancy holds to a system of a client, such as a login and password, an API key or a server. It always belongs to a client or to one of that client's projects.
_Avoid_: Secret, password, login

**Access token**:
What a user creates so that an agent can enter the system and act on their behalf. It has a name and a set of scopes chosen when it is created, and it can be revoked. An action taken with it is recorded as that user's, together with the access token used.
_Avoid_: API key, personal access token

**Scope**:
What an access token is allowed to do: read or write in one area of the system. Reading the value of a credential is a scope of its own.
_Avoid_: Permission, role

**Audit log**:
The permanent record of sensitive operations: reading the value of a credential, logging in, and creating or revoking an access token. Nothing in it is edited or deleted. It is not the history of work that screens show.
_Avoid_: Activity history, history, log

### Languages and translation

**Supported language**:
A language the system offers for its interface, its messages and its configuration content.
_Avoid_: Locale, idiom

**Default language**:
The supported language used when a person's own language is not supported or a version is missing. It is English.
_Avoid_: Fallback language, base language

**Preferred language**:
The supported language a user has chosen to read the system in.
_Avoid_: User locale, display language

**Configuration content**:
Content registered once and reused by everyone, such as templates and the names of categories and statuses. It has one version per supported language.
_Avoid_: Master data, reference data, settings

**Free text**:
Text a person writes in the course of work, such as a description, a comment or a message. It exists only in its original language.
_Avoid_: User content, user-generated content

**Original language**:
The language in which a piece of configuration content or free text was first written.
_Avoid_: Source language, native language

**Automatic draft**:
A version of configuration content proposed by the system and not yet reviewed by a person.
_Avoid_: Machine translation, suggestion

**On-demand translation**:
An automatic rendering of free text into the reader's language, shown at their request. It never replaces the original.
_Avoid_: Auto-translation, translated version
