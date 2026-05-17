package resources

import "fmt"

// PaginatedResource wraps a collection with pagination metadata.
type PaginatedResource[T any] struct {
	Items       []T
	Total       int
	PerPage     int
	CurrentPage int
	LastPage    int
}

// Paginate creates a paginated resource response with data, meta, and links.
// It returns a map containing:
// - data: array of transformed items
// - meta: pagination metadata (total, per_page, current_page, last_page)
// - links: navigation links (first, last, prev, next)
func Paginate[T any](items []T, transformer Transformer[T], total, perPage, currentPage int) map[string]any {
	lastPage := calculateLastPage(total, perPage)

	return map[string]any{
		"data": Collection(items, transformer),
		"meta": map[string]any{
			"total":        total,
			"per_page":     perPage,
			"current_page": currentPage,
			"last_page":    lastPage,
		},
		"links": buildPaginationLinks(currentPage, lastPage),
	}
}

// calculateLastPage calculates the last page number based on total items and per page.
func calculateLastPage(total, perPage int) int {
	if perPage == 0 {
		return 1
	}
	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	if lastPage == 0 {
		lastPage = 1
	}
	return lastPage
}

// buildPaginationLinks creates navigation links for pagination.
func buildPaginationLinks(currentPage, lastPage int) map[string]any {
	links := map[string]any{
		"first": buildPageURL(1),
		"last":  buildPageURL(lastPage),
	}

	if currentPage > 1 {
		links["prev"] = buildPageURL(currentPage - 1)
	} else {
		links["prev"] = nil
	}

	if currentPage < lastPage {
		links["next"] = buildPageURL(currentPage + 1)
	} else {
		links["next"] = nil
	}

	return links
}

// buildPageURL creates a URL for a specific page.
// In a real implementation, this would use the actual request URL and query parameters.
// For now, it returns a simple page parameter string.
func buildPageURL(page int) string {
	return fmt.Sprintf("?page=%d", page)
}
