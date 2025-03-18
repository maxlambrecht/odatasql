package internal

import (
	"fmt"
)

const (
	sqlEq = "="
	sqlNe = "!="
	sqlGt = ">"
	sqlGe = ">="
	sqlLt = "<"
	sqlLe = "<="
)

// opMapping maps OData operators to SQL operators.
var opMapping = map[string]string{
	"eq": sqlEq,
	"ne": sqlNe,
	"gt": sqlGt,
	"ge": sqlGe,
	"lt": sqlLt,
	"le": sqlLe,
}

// validOperators is a set for quick operator validation.
var validOperators = map[string]struct{}{
	"eq": {}, "ne": {}, "gt": {}, "ge": {}, "lt": {}, "le": {},
}

// BuildAST converts an OData filter string into an AST by tokenizing and parsing it.
func BuildAST(filter string) (Node, error) {
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
func parse(tokens []token) (Node, error) {
	p := &parser{tokens: tokens}
	node, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if !p.isAtEnd() {
		return nil, fmt.Errorf("unexpected extra tokens: %v", p.current())
	}
	return node, nil
}

// --- Recursive Descent Parsing ---

func (p *parser) parseExpression() (Node, error) {
	return p.parseOr()
}

// parseOr handles OR expressions: `<andExpr> OR <andExpr>`.
func (p *parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.match(tOpOr) {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryNode{opOr, left, right}
	}
	return left, nil
}

// parseAnd handles AND expressions: `<notExpr> AND <notExpr>`.
func (p *parser) parseAnd() (Node, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.match(tOpAnd) {
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &BinaryNode{opAnd, left, right}
	}
	return left, nil
}

// parseNot handles NOT expressions: `NOT <primaryExpr>`.
func (p *parser) parseNot() (Node, error) {
	if p.match(tOpNot) {
		if p.isAtEnd() {
			return nil, fmt.Errorf("invalid use of NOT: missing expression")
		}
		child, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &NotNode{child}, nil
	}
	return p.parsePrimary()
}

// parsePrimary handles parenthesized expressions and simple conditions.
func (p *parser) parsePrimary() (Node, error) {
	if p.match(tParenOpen) {
		node, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.expect(tParenClose) {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		return &ParenNode{Child: node}, nil
	}
	return p.parseConditionOrIn()
}

// parseConditionOrIn parses conditions like `field eq value` or `field in (value1, value2)`.
func (p *parser) parseConditionOrIn() (Node, error) {
	if !p.check(tIdentifier) {
		return nil, fmt.Errorf("expected field name, got %v", p.current())
	}

	// Extract field name
	fieldTok := p.current()
	p.advance()
	field := toSnakeCase(fieldTok.val)

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

			values = append(values, sanitizeValue(tok.val))
			p.advance()

			if !p.match(tComma) {
				break
			}
		}

		if !p.expect(tParenClose) {
			return nil, fmt.Errorf("missing closing parenthesis in IN list")
		}

		return &InNode{Field: field, Values: values}, nil
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
	if valTok.typ != tString && valTok.typ != tNumber && valTok.typ != tIdentifier {
		return nil, fmt.Errorf("invalid value: %v", valTok)
	}
	p.advance()

	return &ConditionNode{Field: field, Op: sqlOp, Value: sanitizeValue(valTok.val)}, nil
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

// isValidOperator checks if an operator is valid.
func isValidOperator(op string) bool {
	_, exists := validOperators[op]
	return exists
}
