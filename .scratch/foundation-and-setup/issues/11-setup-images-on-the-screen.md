# 11 - Images on the setup screen

Status: ready-for-agent
Blocked by: 09, 10

Read `../spec.md`, `DESIGN.md` and, in `CODING_STANDARDS.md`, State and Components in the frontend rules.

## What to build

On the setup screen the person chooses a logo for the consultancy and a photo for the administrator, sees each one, and can replace or remove it before concluding.

- A field for the logo on the step of the consultancy and a field for the photo on the step of the administrator. Both are optional.
- An image is checked when it is chosen: a PNG, a JPEG or a WebP of at most 2 MB. Another type or a larger file is refused on the field, with the reason.
- An accepted image is sent to the storage right then, through the upload of the setup, and shown from the browser's own copy. The field shows that it is being sent, and an upload that fails shows an error on the field with a way to try again.
- The person can replace or remove an image already chosen.
- The review shows both images.
- Concluding sends the identifiers of the files that are on the screen at that moment.
- A code that expires while an image is being sent returns to the step of the code. With a new code, an image already sent stays, and one that was not is sent again.
- The upload of several steps is one reducer whose state is a union of named steps, as State asks.
- Every text in the three supported languages.

## Acceptance criteria

- [ ] A person completes the setup with a logo and a photo, and both are kept by the consultancy and by the administrator.
- [ ] A file of another type, and a file over 2 MB, are refused on the field with the reason, and nothing is sent.
- [ ] The image chosen shows on its step and on the review.
- [ ] Replacing an image sends the new one, and concluding keeps only the new one. Removing it concludes with none.
- [ ] The setup concludes with neither image.
- [ ] An upload that fails shows an error on its field, and trying again works.
- [ ] Concluding is not possible while an image is still being sent.
- [ ] Correct in light and dark, in the three languages, at desktop width and at 375px, and fully operable by keyboard.
- [ ] Screen tests cover an image accepted, each refusal, a replacement, a removal and a failed upload.
