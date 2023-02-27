// Package pagination provides generic definitions for implementing `Timestamp_ID` Continuation Token
// pagination algorithm. This algorithm basically performs the query below:
//
// SELECT ... FROM ... WHERE (timestampColumn > TS OR (timestampColumn = TS AND idColumn > ID)) ORDER BY tsColumn asc, idColumn asc;
//
// - `ID`: The ID of the last element in the previous page.
// - `TS`: The timestamp of the last element in the previous page. Usually mapped to a column like `created` or `modified`.
//
// How actual "Pages" (sub-sets of objects) are model is not in the scope of this package, and is something each consumer
// should take care of. An example of a page could be:
//
//	type Page struct {
//		Items []MyObjects
//		Total uint64
//		Next  pagination.ContinuationToken
//	}
package pagination

// Pager encapsulates information on how pagination should be performed, this is
// used as an input for listing method of repository implementations.
type Pager struct {
	// Page a string representation of a ContinuationToken. This should be the
	// token from a previous page.
	Page string

	// Size indicates how many items per page should be fetched.
	Size int32

	// Query is an optional value that can be used to apply arbitrary filtering
	// conditions on the set being paginated.
	Query string
}
