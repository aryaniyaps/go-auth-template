package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCursorForID(t *testing.T) {
	t.Run("Valid ID returns cursor", func(t *testing.T) {
		cursor := GetCursorForID(123)
		require.NotNil(t, cursor)
		assert.NotEmpty(t, *cursor)
	})

	t.Run("Zero ID returns nil", func(t *testing.T) {
		cursor := GetCursorForID(0)
		assert.Nil(t, cursor)
	})

	t.Run("Negative ID returns nil", func(t *testing.T) {
		cursor := GetCursorForID(-1)
		assert.Nil(t, cursor)
	})
}

func TestParseCursor(t *testing.T) {
	t.Run("Valid cursor parses correctly", func(t *testing.T) {
		originalID := int64(123)
		cursor := GetCursorForID(originalID)
		require.NotNil(t, cursor)

		parsedID, err := ParseCursor(*cursor)
		require.NoError(t, err)
		assert.Equal(t, originalID, parsedID)
	})

	t.Run("Empty cursor returns zero", func(t *testing.T) {
		parsedID, err := ParseCursor("")
		require.NoError(t, err)
		assert.Equal(t, int64(0), parsedID)
	})

	t.Run("Invalid cursor returns error", func(t *testing.T) {
		parsedID, err := ParseCursor("invalid-base64!")
		assert.Error(t, err)
		assert.Equal(t, int64(0), parsedID)
	})

	t.Run("Invalid cursor ID returns error", func(t *testing.T) {
		invalidCursor := "invalid!"
		parsedID, err := ParseCursor(invalidCursor)
		assert.Error(t, err)
		assert.Equal(t, int64(0), parsedID)
	})
}

func TestNewPaginatedResult(t *testing.T) {
	t.Run("Creates paginated result with all fields", func(t *testing.T) {
		data := []string{"item1", "item2"}
		startCursor := GetCursorForID(1)
		endCursor := GetCursorForID(2)

		result := NewPaginatedResult(data, true, false, startCursor, endCursor)

		assert.Equal(t, data, result.Data)
		assert.True(t, result.HasNextPage)
		assert.False(t, result.HasPreviousPage)
		assert.Equal(t, startCursor, result.StartCursor)
		assert.Equal(t, endCursor, result.EndCursor)
	})

	t.Run("Creates paginated result with nil cursors", func(t *testing.T) {
		data := []string{"item1"}

		result := NewPaginatedResult(data, false, false, nil, nil)

		assert.Equal(t, data, result.Data)
		assert.False(t, result.HasNextPage)
		assert.False(t, result.HasPreviousPage)
		assert.Nil(t, result.StartCursor)
		assert.Nil(t, result.EndCursor)
	})
}

func TestValidatePaginationOptions(t *testing.T) {
	t.Run("Valid options pass validation", func(t *testing.T) {
		options := PaginationOptions{
			First: intPtr(10),
		}
		err := ValidatePaginationOptions(options)
		assert.NoError(t, err)
	})

	t.Run("Zero first parameter fails validation", func(t *testing.T) {
		options := PaginationOptions{
			First: intPtr(0),
		}
		err := ValidatePaginationOptions(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first parameter must be positive")
	})

	t.Run("Negative last parameter fails validation", func(t *testing.T) {
		options := PaginationOptions{
			Last: intPtr(-1),
		}
		err := ValidatePaginationOptions(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "last parameter must be positive")
	})

	t.Run("Both first and last parameters fail validation", func(t *testing.T) {
		options := PaginationOptions{
			First: intPtr(10),
			Last:  intPtr(5),
		}
		err := ValidatePaginationOptions(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot specify both first and last")
	})
}

func TestPaginationEdgeCases(t *testing.T) {
	t.Run("Cursor roundtrip preserves ID", func(t *testing.T) {
		testIDs := []int64{1, 999, 999999, 123456789}

		for _, id := range testIDs {
			cursor := GetCursorForID(id)
			require.NotNil(t, cursor)

			parsedID, err := ParseCursor(*cursor)
			require.NoError(t, err)
			assert.Equal(t, id, parsedID)
		}
	})

	t.Run("Empty data creates valid result", func(t *testing.T) {
		var data []string

		result := NewPaginatedResult(data, false, false, nil, nil)

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