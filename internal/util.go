package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var camelToSnakeRegex = regexp.MustCompile(`([a-z0-9])([A-Z])`)

var reservedSQLKeywords = map[string]struct{}{
	"select": {}, "insert": {}, "update": {}, "delete": {}, "drop": {}, "alter": {},
	"from": {}, "where": {}, "join": {}, "order": {}, "group": {}, "having": {},
	"limit": {}, "offset": {}, "union": {}, "except": {}, "intersect": {},
}

// ToSnakeCase converts a camelCase or PascalCase string to snake_case.
func ToSnakeCase(s string) string {
	return strings.ToLower(camelToSnakeRegex.ReplaceAllString(s, "${1}_${2}"))
}

// SanitizeValue returns a properly formatted SQL literal for a value.
func SanitizeValue(value string) string {
	value = strings.TrimSpace(value)

	// Check if the value is a number (integer or float)
	if _, err := strconv.Atoi(value); err == nil {
		return value
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return value
	}

	// Check for booleans and null
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" || lower == "null" {
		return lower
	}

	// OData escapes single quotes as '' -> We must keep it as ''
	value = strings.Trim(value, "'")             // Remove surrounding single quotes
	value = strings.ReplaceAll(value, "'", "''") // Escape single quotes inside

	// Wrap in single quotes for SQL output
	return fmt.Sprintf("'%s'", value)
}

func IsReservedSQLKeyword(s string) bool {
	_, exists := reservedSQLKeywords[strings.ToLower(s)]
	return exists
}
