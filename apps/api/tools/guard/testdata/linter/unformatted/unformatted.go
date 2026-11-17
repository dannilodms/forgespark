// Package unformatted is formatted as gofmt asks and not as gofumpt does:
// a function body starts with an empty line.
package unformatted

// Double returns twice the number.
func Double(number int) int {

	return number * 2
}
