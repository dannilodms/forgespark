package s3

// ListInPagesOf makes the listing of the storage ask for this many keys a
// page, so that a test proves it follows every page without filling the
// bucket with the thousand objects of a page of the storage's own size.
func ListInPagesOf(storage *Storage, size int32) {
	storage.listPageSize = size
}
