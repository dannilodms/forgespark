# 11 - Images on the setup screen

Status: resolved
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

- [x] A person completes the setup with a logo and a photo, and both are kept by the consultancy and by the administrator.
- [x] A file of another type, and a file over 2 MB, are refused on the field with the reason, and nothing is sent.
- [x] The image chosen shows on its step and on the review.
- [x] Replacing an image sends the new one, and concluding keeps only the new one. Removing it concludes with none.
- [x] The setup concludes with neither image.
- [x] An upload that fails shows an error on its field, and trying again works.
- [x] Concluding is not possible while an image is still being sent.
- [x] Correct in light and dark, in the three languages, at desktop width and at 375px, and fully operable by keyboard.
- [x] Screen tests cover an image accepted, each refusal, a replacement, a removal and a failed upload.

## Comments

### What was built

- **The upload of an image** (`consultancy/components/setup-image.ts`): one reducer whose state is a union of named steps, `none`, `sending`, `sent`, `failed` and `waitingForCode`, in `useSetupImage`, which chooses, removes, sends again and resumes with a new setup code. The screen holds two, the logo and the photo, beside the flow and the form, so an image keeps going while the person moves between steps and the review reads both.
- **Sending** (`sendSetupImage` in `consultancy/data/setup.ts`, `sendToTheStorage` in `shared/api/storage.ts`): `POST /api/setup/uploads` with the code as typed, the purpose and the type and size of the file, then a `PUT` of the file to the URL answered, with its own `Content-Type`. The image shows from the browser's own copy (`URL.createObjectURL`), which is let go of when the image is replaced or removed and when the screen closes.
- **The check when chosen** (`refusalOf`): the type and the size the contract declares (`StartSetupUploadRequestMediaType`, `startSetupUploadBodySizeMax`), in the codes the API answers with: `file_type_unsupported`, `file_empty`, `file_too_large` with `max_size`. A refused file is not sent and leaves the image chosen before.
- **The field** (`consultancy/components/image-field.tsx`), at the end of the step of the consultancy (logo) and of the administrator (photo): a thumbnail, the name of the file and its status (a sign and a word: sending, sent, not sent, waiting for a new setup code), the button that opens the browser's file picker ("Choose a logo", then "Replace"), Remove, the types and the size it takes, and the problem of the field. A failed upload shows why, the trace identifier, and Try again.
- **The file is a field of the form**: `consultancy.logo_file_id` and `administrator.photo_file_id`, set once the image on the screen is stored and emptied when another is chosen or it is removed, so concluding sends the files on the screen. A result that arrives for an image replaced or removed meanwhile is dropped. A 422 of concluding on either field (`file_not_found`, `file_not_uploaded`, `file_too_large`, `file_type_unsupported`) shows on the field, on its step, and goes when another image is chosen.
- **A setup code that stops working while an image is sent** returns to the step of the code; the image waits, and a new code sends it again. An image already sent stays, and concluding sends it with the new code.
- **The review** shows both images, each with its status, and concluding waits, with a sentence that says why, while an image is being sent or was not sent.
- **`{{max_size, fileSize}}`**: a formatter of i18next that writes a size in bytes in megabytes, in the language of the screen, with `Intl.NumberFormat`.
- **Texts** in English, Portuguese and Spanish: the field, its statuses, its failure, the two sentences of the review, and the five field codes of files.
- **Documents.** `CODING_STANDARDS.md`: a Files item in Forms and validation.

### Decided while building

- **The upload lives in the screen, the file in the form.** The upload outlives its step, as the preview had to; the identifier of the file is a field of the form, so concluding builds the body as it did and a 422 lands on the field with no mapping.
- **Concluding also waits for an image that failed**, not only for one being sent: concluding sends the files of the images on the screen, and a failed image has none. The review says what is missing.
- **A refusal is the state of the field**, not a step of the upload: the file refused never starts an upload, and the image before it stays.
- **The picker is a hidden input**, named by the label of the field, opened by a button that keeps the focus through a choice; Remove and Try again give the focus back to it.
- **No component entered `ui`**: the field is made of Button, Field and the status indicator. `shadcn add` was not run, and `DESIGN.md` is unchanged.
- **2 MiB is written "2 MB"**, as `Intl` names the unit; the API's limit is 2097152 bytes.

### What was checked

- **The screen tests**, 77 in the app (19 new in `setup-images.test.tsx`): an image chosen, sent with the exact body and `PUT`, shown from a `blob:` copy with its status; the logo and the photo on the review and concluded by their files; a PDF, a PNG over 2 MB and an empty file refused with the reason and nothing sent; a refusal that leaves the image before; a replacement, concluding with the new file only; a replacement while the first is still being sent, whatever the first answers; a removal, concluding with no file and the focus on the button; a 500 with its trace identifier and Try again; a `PUT` refused by the storage and Try again with a new upload; a 429; concluding disabled while an image is being sent and while one was not sent; a code that stops working on an upload, back to the step of the code with everything kept, and the image sent with the new code; an image already sent kept with the new code; a 422 of concluding on the logo on its field, gone when another image is chosen; the field operated with the keys, Enter opening the picker; the texts in Portuguese with the size. The tests of ticket 09 conclude with neither image, and still pass with the exact body.
- **That the tests bite**: without the check of the image on the screen, the test of a replacement while the first is sent concluded with the first file and failed.
- **The real system, through the shared Caddy and Garage, in a real Chromium** (`fs-11-manual.sh`, a database of its own, dropped after): a text file and a PNG over 2 MB were refused on the field and started no upload; a PNG logo was sent (`POST /api/setup/uploads` 201, then `PUT files.localhost/...` with `Content-Type=image/png`, 200, from the origin `http://localhost`), then replaced by a WebP (sent the same way); on the step of the administrator a second code was issued, which ended the first; the JPEG photo met `422` on its upload, the screen went back to the step of the code, the second code was typed, and the photo was sent with it, the typed name kept; the review showed both images from `blob:` copies at their real sizes; concluding answered 204 and led to `/login?notice=setup_concluded`. The database then held the consultancy with the WebP logo and the administrator with the photo, both files confirmed, and the first PNG not confirmed; a signed `GET` found the three objects in Garage at the sizes declared.
- **The look**, by screenshots of Vitest's Browser Mode, looked at: the field empty in the three languages, the logo sent in the three languages with a long file name, the photo sent in the three, the logo being sent, failed (Spanish) and refused (English), the review with both images in the three languages and the review waiting for an image, each in light and in dark, at 1280px and at 375px, with no horizontal overflow measured on any.
- **`task check`** passes on the branch. `feat/foundation-and-setup` had not moved from `82c0f7b` (ticket 09), where the branch started, so the merge was empty. The first two runs failed on one test of `cmd/api` this ticket does not touch (`TestTheAPIExportsTheSpanOfARequestAndOfItsQueriesToTheAddressOfTheConfiguration`: readiness answered 503 past its 2-second deadline while every package was tested at once); the test and its package passed alone, and the third run passed whole.
- **The main checkout** holds no change: `git status --short` there printed nothing.

### What was not checked

- **A setup code that expires with time while an image is sent.** The API gives the same answer for an expired code and for one a new code ended; the real check used the second.
- **A refusal of an image by the real API when concluding** (an object that is not an image): the screen checks the type first, so the 422 on an image field was simulated in the screen tests only.
- **The browser's own file picker**, which a test cannot drive: the tests see that Enter asks for it and choose through the input, and the real check chose files the same way.
- **A screen reader**, and browsers other than Chromium.
- **The pipeline.** Nothing was pushed.
- **The translations** were written by the agent and read once.
