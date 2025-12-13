package db

import (
	"fmt"
	"strconv"

	"github.com/uptrace/bun"
)

// PaginationOptions contains pagination parameters
type PaginationOptions struct {
	First  *int    `json:"first,omitempty"`
	Last   *int    `json:"last,omitempty"`
	After  *string `json:"after,omitempty"`
	Before *string `json:"before,omitempty"`
}

// PaginatedResult represents a paginated result set with cursor-based navigation
type PaginatedResult[T any, C any] struct {
	Data            []T  `json:"data"`
	HasNextPage     bool `json:"has_next_page"`
	HasPreviousPage bool `json:"has_previous_page"`
	StartCursor     *C   `json:"start_cursor,omitempty"`
	EndCursor       *C   `json:"end_cursor,omitempty"`
}

// ApplyPagination applies cursor-based pagination to a Bun query using integer cursors
func ApplyPagination(query *bun.SelectQuery, options PaginationOptions) *bun.SelectQuery {
	// Apply "after" cursor (forward pagination)
	if options.After != nil {
		if afterID, err := strconv.ParseInt(*options.After, 10, 64); err == nil && afterID > 0 {
			query = query.Where("id < ?", afterID)
		}
	}

	// Apply "before" cursor (backward pagination)
	if options.Before != nil {
		if beforeID, err := strconv.ParseInt(*options.Before, 10, 64); err == nil && beforeID > 0 {
			query = query.Where("id > ?", beforeID)
		}
	}

	// Order by ID descending for newest first
	query = query.Order("id DESC")

	// Apply limit with +1 to check for additional pages
	if options.First != nil && *options.First > 0 {
		query = query.Limit(*options.First + 1)
	} else if options.Last != nil && *options.Last > 0 {
		query = query.Limit(*options.Last + 1)
	}

	return query
}

// ProcessPaginatedResult processes a paginated query result and creates pagination metadata
// Works with any type T that embeds bun.BaseModel (has ID field)
// C is the cursor type (can be string, int64, or any custom cursor type)
func ProcessPaginatedResult[T interface{ GetID() C }, C any](items []T, first, last *int) PaginatedResult[T, C] {
	result := PaginatedResult[T, C]{
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     nil,
		EndCursor:       nil,
	}

	if len(items) == 0 {
		return result
	}

	// Check if we have an extra item to determine pagination state
	processedItems := items
	if first != nil && len(items) > *first {
		result.HasNextPage = true
		processedItems = items[:len(items)-1] // Remove the extra item
	} else if last != nil && len(items) > *last {
		result.HasPreviousPage = true
		processedItems = items[1:] // Remove the first item
	}

	// Create cursors if we have items
	if len(processedItems) > 0 {
		startCursor := processedItems[0].GetID()
		endCursor := processedItems[len(processedItems)-1].GetID()
		result.StartCursor = &startCursor
		result.EndCursor = &endCursor
	}

	result.Data = processedItems
	return result
}

// ValidatePagination validates pagination parameters
func ValidatePagination(options PaginationOptions) error {
	if options.First != nil && *options.First <= 0 {
		return fmt.Errorf("first parameter must be positive")
	}

	if options.Last != nil && *options.Last <= 0 {
		return fmt.Errorf("last parameter must be positive")
	}

	if options.First != nil && options.Last != nil {
		return fmt.Errorf("cannot specify both first and last parameters")
	}

	return nil
}
