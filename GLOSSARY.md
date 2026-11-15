# My Software Factory Manager

The operational hub of a software consultancy: one place to manage clients, their projects and their software development requests, from first contact to delivery and maintenance.

## Language

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
