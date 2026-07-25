// Package pagination implements simple offset-based pagination where the
// page token is the string form of the offset to resume from.
package pagination

import "strconv"

const DefaultPageSize = 20

// Page holds the resolved offset/limit for a single page request.
type Page struct {
	Offset int
	Limit  int
}

// Parse turns a (pageSize, pageToken) request pair into an offset/limit page.
// An invalid or empty pageToken is treated as offset 0.
func Parse(pageSize int32, pageToken string) Page {
	limit := int(pageSize)
	if limit <= 0 {
		limit = DefaultPageSize
	}

	offset := 0
	if pageToken != "" {
		if v, err := strconv.Atoi(pageToken); err == nil && v > 0 {
			offset = v
		}
	}

	return Page{Offset: offset, Limit: limit}
}

// NextToken returns the token for the following page, or "" if this page
// wasn't full (i.e. there is nothing more to fetch).
func (p Page) NextToken(rowsReturned int) string {
	if rowsReturned < p.Limit {
		return ""
	}
	return strconv.Itoa(p.Offset + rowsReturned)
}
