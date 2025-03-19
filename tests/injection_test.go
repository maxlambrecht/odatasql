package tests

import (
	"testing"

	"github.com/maxlambrecht/odatasql"
	"github.com/stretchr/testify/assert"
)

func TestFilterToSQL_Injection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		// --- Malicious SQL Injection Attempts ---
		{"SQL Injection: DROP TABLE via Value", "id eq '1; DROP TABLE users --'"},
		{"SQL Injection: Standalone Statement", "1; DROP TABLE users --'"},
		{"SQL Injection: Direct DROP TABLE", "DROP TABLE users"},

		// --- Boolean, Numeric, and NULL Injection Attempts ---
		{"Boolean as Field", "true eq false"},
		{"Numeric as Field", "42 eq 42"},
		{"NULL as Field", "null eq null"},
		{"Boolean Comparison with Field", "status eq true eq false"},

		// --- Parentheses Manipulation Attacks ---
		{"Empty Parentheses", "()"},
		{"Parentheses Around Literal", "(42) eq 42"},
		{"Nested Parentheses with Literal", "((true)) eq false"},
		{"Unmatched Opening Parenthesis", "(name eq 'Alice'"},
		{"Unmatched Closing Parenthesis", "name eq 'Alice')"},

		// --- SQL Comment & Concatenation Injection ---
		{"SQL Comment Attempt", "name eq 'Alice' --"},
		{"SQL Concatenation Attempt", "name eq 'Alice' || 'Bob'"},

		// --- Logical Operator Abuse ---
		{"OR Without Right Side", "name eq 'Alice' or"},
		{"AND Without Right Side", "name eq 'Alice' and"},
		{"NOT Without Operand", "not"},

		// --- IN Operator Injection Attempts ---
		{"Empty IN List", "color in ()"},
		{"IN with Boolean", "color in (true, false)"},
		{"IN with NULL", "color in (null, 'red')"},

		// --- SQL Keyword Manipulation ---
		{"Quoted Field Name", "'name' eq 'Alice'"},
		{"SQL Keyword as Field", "SELECT eq 'Alice'"},

		// --- Excessive Nesting Attacks ---
		{"Excessive Nesting", "(((((((((((name eq 'Alice')))))))))))"},

		// --- Always-True/Always-False Attacks ---
		{"Boolean Always True", "name eq 'Alice' or true eq true"},
		{"Boolean Always False", "name eq 'Alice' and false eq false"},
		{"Numeric Always True", "age gt 30 or 1 eq 1"},
		{"NULL Always True", "age gt 30 or null eq null"},

		// --- Arithmetic & Expression Manipulation ---
		{"Arithmetic in Filter", "age eq salary + 1"},
		{"CASE Statement Manipulation", "CASE WHEN 1=1 THEN username ELSE password END"},
		{"Substring Manipulation", "substringof('Alice', name) eq true"},

		// --- Time-Based SQL Injection Attempts ---
		{"Time Delay Attack", "name eq 'Alice' AND SLEEP(5)"},
		{"Benchmark Attack", "name eq 'Alice' AND BENCHMARK(1000000, MD5('test'))"},

		// --- Path Traversal & Encoded SQL Injection ---
		{"Comment Escape Attempt", "name eq 'Alice' /* test */"},
		{"Encoded SQL Injection", "name eq '%27%20or%201=1--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := odatasql.FilterToSQL(tt.input)

			assert.Error(t, err, "Expected an error for input: %q", tt.input)
		})
	}
}
