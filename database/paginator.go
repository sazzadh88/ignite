package database

import "fmt"

// Paginator represents pagination results.
type Paginator struct {
	// Items contains the current page's data
	Items []map[string]any
	// Total is the total number of records
	Total int64
	// PerPage is the number of items per page
	PerPage int
	// CurrentPage is the current page number (1-indexed)
	CurrentPage int
	// LastPage is the last page number
	LastPage int
}

// HasMorePages returns true if there are more pages available.
func (p *Paginator) HasMorePages() bool {
	return p.CurrentPage < p.LastPage
}

// NextPageURL returns a URL representation for the next page.
// In a real implementation, this would build a full URL with query params.
func (p *Paginator) NextPageURL() string {
	if !p.HasMorePages() {
		return ""
	}
	return fmt.Sprintf("?page=%d", p.CurrentPage+1)
}

// PrevPageURL returns a URL representation for the previous page.
func (p *Paginator) PrevPageURL() string {
	if p.CurrentPage <= 1 {
		return ""
	}
	return fmt.Sprintf("?page=%d", p.CurrentPage-1)
}
