/**
 * Sends a file to the file storage, at a URL the API signed for it (Files in
 * CODING_STANDARDS.md). The type and the size were declared when the upload
 * started and are signed with the URL, so the `PUT` carries the file's own:
 * its `Content-Type`, and its bytes, whose count the browser sends as
 * `Content-Length`. Nothing else is sent: no cookie, no other header.
 *
 * It resolves once the storage keeps the file, and rejects when the storage
 * refuses it, or as fetch does when the network fails.
 */
export async function sendToTheStorage(url: string, file: File) {
  const response = await fetch(url, {
    method: "PUT",
    headers: { "Content-Type": file.type },
    body: file,
  })
  if (!response.ok) {
    throw new Error(`The file storage answered ${response.status}`)
  }
}
