// Package format provides interfaces to format content into various kinds of
// data
package format

// JSONFormattable is implemented with the method FormatJSON, which must return the ordered
// slice of JSON struct tag names for the type implementing it
type JSONFormattable interface {
	FormatJSON()
}
