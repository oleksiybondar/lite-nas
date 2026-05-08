package dto

// PaginationMeta defines the page-based pagination metadata for list
// responses.
//
// The pagination object itself should be optional on a response. When it is
// present, all fields in this structure should be populated consistently.
type PaginationMeta struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalItems int  `json:"total_items"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}
