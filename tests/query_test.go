package tests

import (
	"testing"

	"github.com/maxlambrecht/odatasql"
	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// --- Basic Comparisons ---
		{"Basic eq", "name eq 'Bob'", "name = 'Bob'", false},
		{"Basic ne", "status ne 'inactive'", "status != 'inactive'", false},
		{"Basic gt", "age gt 18", "age > 18", false},
		{"Basic ge", "height ge 170", "height >= 170", false},
		{"Basic lt", "score lt 50", "score < 50", false},
		{"Basic le", "price le 99.99", "price <= 99.99", false},

		// --- Logical Operators ---
		{"AND operator", "age gt 18 and status eq 'active'", "age > 18 AND status = 'active'", false},
		{"OR operator", "age lt 18 or status eq 'inactive'", "age < 18 OR status = 'inactive'", false},
		{"NOT operator", "not age lt 18", "NOT age < 18", false},

		// --- Parentheses & Precedence ---
		{"Parentheses grouping", "(age gt 18 and status eq 'active') or premium eq true", "(age > 18 AND status = 'active') OR premium = true", false},
		{"Complex precedence", "not (age gt 18 and status eq 'active') or premium eq true", "(NOT (age > 18 AND status = 'active')) OR premium = true", false},
		{"Nested parentheses", "((name eq 'Bob'))", "((name = 'Bob'))", false},
		{"Deeply nested logic", "((name eq 'Bob' and age gt 25) or (status eq 'active' and premium eq true))", "((name = 'Bob' AND age > 25) OR (status = 'active' AND premium = true))", false},
		{"Double NOT", "not not name eq 'Bob'", "NOT (NOT name = 'Bob')", false},

		// --- IN Operator ---
		{"IN with strings", "color in ('red', 'blue')", "color IN ('red', 'blue')", false},
		{"IN with numbers", "age in (20, 25, 30)", "age IN (20, 25, 30)", false},
		{"IN with single value", "color in ('red')", "color IN ('red')", false},
		{"Malformed IN (empty)", "color in ()", "", true},

		// --- Quoting and String Literals ---
		{"String with spaces", "name eq 'John Doe'", "name = 'John Doe'", false},
		{"Quoted values with single quotes", "nickname eq 'O''Brien'", "nickname = 'O''Brien'", false},

		// --- Boolean and Null Literals ---
		{"Boolean true", "isActive eq true", "is_active = true", false},
		{"Boolean false", "isDeleted eq false", "is_deleted = false", false},
		{"Null equality", "deletedAt eq null", "deleted_at = null", false},
		{"Null inequality", "deletedAt ne null", "deleted_at != null", false},

		// --- Whitespace Variations & Snake Case ---
		{"Extra spaces", "   name   eq    'Alice'   ", "name = 'Alice'", false},
		{"Mixed spacing and operators", "age    gt   30  and    status   eq  'active'", "age > 30 AND status = 'active'", false},
		{"CamelCase field conversion", "FirstName eq 'John'", "first_name = 'John'", false},

		// --- Combined Logical Operators with NOT, AND, OR, IN ---
		{"NOT before AND", "not age gt 30 and status eq 'active'", "(NOT age > 30) AND status = 'active'", false},
		{"NOT before OR", "not age gt 30 or status eq 'active'", "(NOT age > 30) OR status = 'active'", false},
		{"NOT with parentheses", "not (age gt 30 and status eq 'active')", "NOT (age > 30 AND status = 'active')", false},
		{"AND before OR", "age gt 30 and status eq 'active' or premium eq true", "(age > 30 AND status = 'active') OR premium = true", false},
		{"OR before AND", "age gt 30 or status eq 'active' and premium eq true", "age > 30 OR (status = 'active' AND premium = true)", false},
		{"Parentheses force grouping", "(age gt 30 or status eq 'active') and premium eq true", "(age > 30 OR status = 'active') AND premium = true", false},
		{"IN combined with AND", "color in ('red', 'blue') and status eq 'active'", "color IN ('red', 'blue') AND status = 'active'", false},
		{"IN combined with OR", "color in ('red', 'blue') or status eq 'active'", "color IN ('red', 'blue') OR status = 'active'", false},
		{"IN with NOT", "not color in ('red', 'blue')", "NOT color IN ('red', 'blue')", false},
		{"Double NOT", "not not name eq 'Alice'", "NOT (NOT name = 'Alice')", false},
		{"Triple NOT", "not not not name eq 'Alice'", "NOT (NOT (NOT name = 'Alice'))", false},

		// --- Case Sensitivity Tests ---
		{"Case insensitive EQ", "name EQ 'Bob'", "name = 'Bob'", false},
		{"Case insensitive Ne", "status Ne 'inactive'", "status != 'inactive'", false},
		{"Case insensitive Gt", "age Gt 18", "age > 18", false},

		// --- Error Cases ---
		{"Invalid operator", "name xx 'Bob'", "", true},
		{"Unclosed parenthesis", "(age gt 18 and status eq 'active'", "", true},
		{"Unexpected token", "age gt 18 and or status eq 'active'", "", true},
		{"Malformed IN clause", "color in 'red', 'blue')", "", true},
		{"Malformed NOT", "not", "", true},
		{"Empty input", "", "", false},
		{"Leading OR", "or age gt 30", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sql, err := odatasql.FilterToSQL(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "FilterToSQL(%q) expected error", tt.input)
				return
			}

			assert.NoError(t, err, "FilterToSQL(%q) did not expect an error", tt.input)
			assert.Equal(t, tt.expected, sql, "FilterToSQL(%q) = %q, want %q", tt.input, sql, tt.expected)
		})
	}
}
