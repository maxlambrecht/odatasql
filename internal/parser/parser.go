package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/maxlambrecht/odatasql/internal/ast"
)

const (
	sqlEq = "="
	sqlNe = "!="
	sqlGt = ">"
	sqlGe = ">="
	sqlLt = "<"
	sqlLe = "<="
)

const maxNestingDepth = 10

var camelToSnakeRegex = regexp.MustCompile(`([a-z0-9])([A-Z])`)

var reservedSQLKeywords = map[string]struct{}{
	"select": {}, "insert": {}, "update": {}, "delete": {}, "drop": {}, "alter": {},
	"from": {}, "where": {}, "join": {}, "order": {}, "group": {}, "having": {},
	"limit": {}, "offset": {}, "union": {}, "except": {}, "intersect": {},
}

// opMapping maps OData operators to SQL operators.
var opMapping = map[string]string{
	"eq": sqlEq,
	"ne": sqlNe,
	"gt": sqlGt,
	"ge": sqlGe,
	"lt": sqlLt,
	"le": sqlLe,
}

// BuildAST converts an OData filter string into an AST by tokenizing and parsing it.
func BuildAST(filter string) (ast.Node, error) {
	tokens, err := tokenize(filter)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}
	return parse(tokens)
}

// --- Parser Struct & Entry Point ---

type parser struct {
	tokens []token
	pos    int
}

// parse starts the parsing process and returns the root node of the AST.
func parse(tokens []token) (ast.Node, error) {
	p := &parser{tokens: tokens}
	node, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if !p.isAtEnd() {
		return nil, fmt.Errorf("unexpected extra tokens: %v", p.current())
	}
	return node, nil
}

// --- Recursive Descent Parsing ---

func (p *parser) parseExpression(depth int) (ast.Node, error) {
	if depth > maxNestingDepth {
		return nil, fmt.Errorf("exceeded maximum nesting depth of %d", maxNestingDepth)
	}
	return p.parseOr(depth)
}

// parseOr handles OR expressions: `<andExpr> OR <andExpr>`.
func (p *parser) parseOr(depth int) (ast.Node, error) {
	if depth > maxNestingDepth {
		return nil, fmt.Errorf("exceeded maximum nesting depth of %d", maxNestingDepth)
	}

	left, err := p.parseAnd(depth)
	if err != nil {
		return nil, err
	}

	for p.match(tOpOr) {
		if p.isAtEnd() {
			return nil, fmt.Errorf("expected expression after OR, but found end of input")
		}
		right, err := p.parseAnd(depth)
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryNode{Op: ast.OpOr, Left: left, Right: right}
	}
	return left, nil
}

// parseAnd handles AND expressions: `<notExpr> AND <notExpr>`.
func (p *parser) parseAnd(depth int) (ast.Node, error) {
	if depth > maxNestingDepth {
		return nil, fmt.Errorf("exceeded maximum nesting depth of %d", maxNestingDepth)
	}

	left, err := p.parseNot(depth)
	if err != nil {
		return nil, err
	}
	for p.match(tOpAnd) {
		if p.isAtEnd() {
			return nil, fmt.Errorf("expected expression after AND, but found end of input")
		}

		right, err := p.parseNot(depth)
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryNode{Op: ast.OpAnd, Left: left, Right: right}
	}
	return left, nil
}

// parseNot handles NOT expressions: `NOT <primaryExpr>`.
func (p *parser) parseNot(depth int) (ast.Node, error) {
	if depth > maxNestingDepth {
		return nil, fmt.Errorf("exceeded maximum nesting depth of %d", maxNestingDepth)
	}

	if p.match(tOpNot) {
		if p.isAtEnd() {
			return nil, fmt.Errorf("invalid use of NOT: missing expression")
		}
		child, err := p.parseNot(depth)
		if err != nil {
			return nil, err
		}
		return &ast.NotNode{Child: child}, nil
	}
	return p.parsePrimary(depth)
}

// parsePrimary handles parenthesized expressions and simple conditions.
func (p *parser) parsePrimary(depth int) (ast.Node, error) {
	if depth > maxNestingDepth {
		return nil, fmt.Errorf("exceeded maximum nesting depth of %d", maxNestingDepth)
	}

	if p.match(tParenOpen) {
		node, err := p.parseExpression(depth + 1)
		if err != nil {
			return nil, err
		}
		if !p.expect(tParenClose) {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		return &ast.ParenNode{Child: node}, nil
	}
	return p.parseConditionOrIn()
}

// parseConditionOrIn parses conditions like `field eq value` or `field in (value1, value2)`.
func (p *parser) parseConditionOrIn() (ast.Node, error) {
	if !p.check(tIdentifier) {
		return nil, fmt.Errorf("expected field name, got %v", p.current())
	}

	// Extract field name
	fieldTok := p.current()
	field := toSnakeCase(fieldTok.val)
	p.advance()

	if isReservedSQLKeyword(field) {
		return nil, fmt.Errorf("invalid field name: %q is a reserved SQL keyword", field)
	}

	// --- Handle IN Operator ---
	if p.match(tOpIn) {
		if !p.expect(tParenOpen) {
			return nil, fmt.Errorf("expected '(' after 'IN'")
		}

		var values []string
		if p.check(tParenClose) {
			return nil, fmt.Errorf("IN operator must have at least one value")
		}

		for {
			if p.check(tParenClose) {
				break
			}
			if p.isAtEnd() {
				return nil, fmt.Errorf("unclosed IN list")
			}

			tok := p.current()
			if tok.typ != tString && tok.typ != tNumber && tok.typ != tIdentifier {
				return nil, fmt.Errorf("invalid value in IN list: %v", tok)
			}

			sanitizedValue, err := validateAndSanitizeValue(tok.val)
			if err != nil {
				return nil, fmt.Errorf("invalid value in condition: %w", err)
			}

			values = append(values, sanitizedValue)
			p.advance()

			if !p.match(tComma) {
				break
			}
		}

		if !p.expect(tParenClose) {
			return nil, fmt.Errorf("missing closing parenthesis in IN list")
		}

		return &ast.InNode{Field: field, Values: values}, nil
	}

	// --- Handle Simple Binary Condition ---
	if p.isAtEnd() {
		return nil, fmt.Errorf("expected operator after field %q", fieldTok.val)
	}

	opTok := p.current()
	if !isValidOperator(opTok.val) {
		return nil, fmt.Errorf("unsupported operator: %s", opTok.val)
	}
	p.advance()

	sqlOp := opMapping[opTok.val]

	if p.isAtEnd() {
		return nil, fmt.Errorf("missing value after operator %q", opTok.val)
	}
	valTok := p.current()
	if valTok.typ != tString && valTok.typ != tNumber && valTok.typ != tIdentifier && valTok.typ != tLiteral {
		return nil, fmt.Errorf("invalid value: %v", valTok)
	}
	p.advance()

	sanitizedValue, err := validateAndSanitizeValue(valTok.val)
	if err != nil {
		return nil, fmt.Errorf("invalid value in condition: %w", err)
	}

	return &ast.ConditionNode{Field: field, Op: sqlOp, Value: sanitizedValue}, nil
}

// --- Parser Helper Functions ---

// match advances if the next token is of the given type.
func (p *parser) match(tt tokenType) bool {
	if p.check(tt) {
		p.advance()
		return true
	}
	return false
}

// check returns true if the next token is of the given type.
func (p *parser) check(tt tokenType) bool {
	return !p.isAtEnd() && p.current().typ == tt
}

// current returns the current token.
func (p *parser) current() token {
	return p.tokens[p.pos]
}

// isAtEnd checks if all tokens have been consumed.
func (p *parser) isAtEnd() bool {
	return p.pos >= len(p.tokens)
}

// advance moves to the next token.
func (p *parser) advance() {
	p.pos++
}

// expect advances if the next token is of the given type, otherwise returns false.
func (p *parser) expect(tt tokenType) bool {
	if p.check(tt) {
		p.advance()
		return true
	}
	return false
}

// validateAndSanitizeValue returns a properly formatted SQL literal for a value.
func validateAndSanitizeValue(value string) (string, error) {
	value = strings.TrimSpace(value)

	// Allow valid numeric values without quotes
	if _, err := strconv.Atoi(value); err == nil {
		return value, nil
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return value, nil
	}

	lower := strings.ToLower(value)

	// Prevent SQL injection attempts by blocking dangerous SQL characters and keywords
	bannedPatterns := []string{";", "--", "/*", "*/"}
	for _, pattern := range bannedPatterns {
		if strings.Contains(lower, pattern) {
			return "", fmt.Errorf("invalid input detected: %q", value)
		}
	}

	if isReservedSQLKeyword(lower) {
		return "", fmt.Errorf("invalid input detected: %q is a reserved SQL keyword", value)
	}

	// Allow SQL booleans and NULL as-is (without quotes)
	if lower == "true" || lower == "false" || lower == "null" {
		return lower, nil
	}

	// Sanitize string values: Escape single quotes inside the value
	value = strings.Trim(value, "'")             // Remove surrounding single quotes
	value = strings.ReplaceAll(value, "'", "''") // Escape inner single quotes

	// Wrap in single quotes for safe SQL usage
	return fmt.Sprintf("'%s'", value), nil
}

func toSnakeCase(s string) string {
	return strings.ToLower(camelToSnakeRegex.ReplaceAllString(s, "${1}_${2}"))
}

func isReservedSQLKeyword(s string) bool {
	_, exists := reservedSQLKeywords[strings.ToLower(s)]
	return exists
}

func isValidOperator(op string) bool {
	_, exists := opMapping[op]
	return exists
}
