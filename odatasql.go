package odatasql

import (
	"fmt"
	"strings"

	"github.com/maxlambrecht/odatasql/internal"
)

// FilterToSQL transforms an OData filter string into a SQL WHERE clause.
// It maintains explicit parentheses and ensures correct operator precedence.
//
// Example:
//
//	sql, err := FilterToSQL("name eq 'Alice' and age gt 30")
//	// sql = "name = 'Alice' AND age > 30"
//
// Returns:
//   - A SQL WHERE clause as a string.
//   - An error if the input is invalid.
func FilterToSQL(filter string) (string, error) {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return "", nil
	}

	ast, err := internal.BuildAST(filter)
	if err != nil {
		return "", fmt.Errorf("invalid OData filter %q: %w", filter, err)
	}

	return ast.ToSQL(0), nil
}
