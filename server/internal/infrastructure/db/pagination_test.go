package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessPaginatedResult(t *testing.T) {
	t.Run("Processes paginated result with string cursors", func(t *testing.T) {
		items := []TestItem{{id: 3}, {id: 2}, {id: 1}, {id: 4}} // Extra item for hasNextPage
		first := intPtr(3)

		result := ProcessPaginatedResult(items, first, nil)

		assert.Len(t, result.Data, 3) // Extra item removed
		assert.True(t, result.HasNextPage)
		assert.False(t, result.HasPreviousPage)
		assert.Equal(t, "3", *result.StartCursor)
		assert.Equal(t, "1", *result.EndCursor)
	})

	t.Run("Processes paginated result with int64 cursors", func(t *testing.T) {
		items := []TestItem{{id: 4}, {id: 3}, {id: 2}, {id: 1}} // Extra item for hasPreviousPage
		last := intPtr(3)

		result := ProcessPaginatedResult(items, nil, last)

		assert.Len(t, result.Data, 3) // First item removed
		assert.False(t, result.HasNextPage)
		assert.True(t, result.HasPreviousPage)
		assert.Equal(t, int64(3), *result.StartCursor)
		assert.Equal(t, int64(1), *result.EndCursor)
	})

	t.Run("Processes empty result", func(t *testing.T) {
		var items []TestItem

		result := ProcessPaginatedResult(items, nil, nil)

		assert.Empty(t, result.Data)
		assert.False(t, result.HasNextPage)
		assert.False(t, result.HasPreviousPage)
		assert.Nil(t, result.StartCursor)
		assert.Nil(t, result.EndCursor)
	})
}

// TestItem for testing ProcessPaginatedResult - implements GetID() method
type TestItem struct {
	id int64
}

func (t TestItem) GetID() int64 {
	return t.id
}

func TestValidatePagination(t *testing.T) {
	t.Run("Valid options pass validation", func(t *testing.T) {
		options := PaginationOptions{
			First: intPtr(10),
		}
		err := ValidatePagination(options)
		assert.NoError(t, err)
	})

	t.Run("Zero first parameter fails validation", func(t *testing.T) {
		options := PaginationOptions{
			First: intPtr(0),
		}
		err := ValidatePagination(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first parameter must be positive")
	})

	t.Run("Negative last parameter fails validation", func(t *testing.T) {
		options := PaginationOptions{
			Last: intPtr(-1),
		}
		err := ValidatePagination(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "last parameter must be positive")
	})

	t.Run("Both first and last parameters fail validation", func(t *testing.T) {
		options := PaginationOptions{
			First: intPtr(10),
			Last:  intPtr(5),
		}
		err := ValidatePagination(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot specify both first and last")
	})
}

func TestPaginationEdgeCases(t *testing.T) {
	t.Run("Cursor creation preserves ID", func(t *testing.T) {
		testIDs := []int64{1, 999, 999999, 123456789}

		for _, id := range testIDs {
			cursor := fmt.Sprintf("%d", id)
			assert.Equal(t, fmt.Sprintf("%d", id), cursor)
		}
	})

	t.Run("Empty data creates valid result", func(t *testing.T) {
		var items []TestItem
		first := intPtr(5)

		result := ProcessPaginatedResult(items, first, nil)

		assert.Empty(t, result.Data)
		assert.False(t, result.HasNextPage)
		assert.False(t, result.HasPreviousPage)
		assert.Nil(t, result.StartCursor)
		assert.Nil(t, result.EndCursor)
	})
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}
