package db

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/uptrace/bun"
)

// PaginatedResult represents a paginated result set with cursor-based navigation
type PaginatedResult[T any] struct {
	Data             []T   `json:"data"`
	HasNextPage      bool  `json:"has_next_page"`
	HasPreviousPage  bool  `json:"has_previous_page"`
	StartCursor      *string `json:"start_cursor,omitempty"`
	EndCursor        *string `json:"end_cursor,omitempty"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult[T any](data []T, hasNextPage, hasPreviousPage bool, startCursor, endCursor *string) *PaginatedResult[T] {
	return &PaginatedResult[T]{
		Data:            data,
		HasNextPage:     hasNextPage,
		HasPreviousPage: hasPreviousPage,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}
}

// GetCursorForID converts a database ID to a cursor string
func GetCursorForID(id int64) *string {
	if id <= 0 {
		return nil
	}
	cursor := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", id)))
	return &cursor
}

// ParseCursor converts a cursor string back to a database ID
func ParseCursor(cursor string) (int64, error) {
	if cursor == "" {
		return 0, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, fmt.Errorf("invalid cursor format: %w", err)
	}

	idStr := strings.TrimSpace(string(decoded))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid cursor ID: %w", err)
	}

	if id <= 0 {
		return 0, fmt.Errorf("invalid cursor ID: must be positive")
	}

	return id, nil
}

// PaginationOptions contains pagination parameters
type PaginationOptions struct {
	First *int    `json:"first,omitempty"`
	Last  *int    `json:"last,omitempty"`
	After *string `json:"after,omitempty"`
	Before *string `json:"before,omitempty"`
}

// ApplyCursorPagination applies cursor-based pagination to a Bun query
func ApplyCursorPagination(query *bun.SelectQuery, options PaginationOptions, orderDesc bool) *bun.SelectQuery {
	// Apply "after" cursor (forward pagination)
	if options.After != nil {
		afterID, err := ParseCursor(*options.After)
		if err == nil {
			if orderDesc {
				query = query.Where("id < ?", afterID)
			} else {
				query = query.Where("id > ?", afterID)
			}
		}
	}

	// Apply "before" cursor (backward pagination)
	if options.Before != nil {
		beforeID, err := ParseCursor(*options.Before)
		if err == nil {
			if orderDesc {
				query = query.Where("id > ?", beforeID)
			} else {
				query = query.Where("id < ?", beforeID)
			}
		}
	}

	// Set order
	if orderDesc {
		query = query.Order("id DESC")
	} else {
		query = query.Order("id ASC")
	}

	// Apply limit
	if options.First != nil && *options.First > 0 {
		query = query.Limit(*options.First + 1) // +1 to check for next page
	} else if options.Last != nil && *options.Last > 0 {
		query = query.Limit(*options.Last + 1) // +1 to check for previous page
	}

	return query
}

// ValidatePaginationOptions validates pagination parameters
func ValidatePaginationOptions(options PaginationOptions) error {
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

// Constants for pagination
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)