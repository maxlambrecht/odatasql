package internal

import (
	"fmt"
	"strings"
)

const (
	opAnd = "AND"
	opOr  = "OR"
	opNot = "NOT"
	opIn  = "IN"
)

// Node represents any part of the parsed expression.
type Node interface {
	// ToSQL generates the SQL snippet for the node.
	// The level parameter indicates nesting for internal use.
	ToSQL(level int) string
}

// BinaryNode represents an expression combining two subexpressions with "AND" or "OR".
type BinaryNode struct {
	Op          string // "AND" or "OR"
	Left, Right Node
}

// ToSQL converts a BinaryNode to its SQL representation.
func (b *BinaryNode) ToSQL(level int) string {
	left := b.Left.ToSQL(level + 1)
	right := b.Right.ToSQL(level + 1)
	// For binary nodes, if not wrapped explicitly then add parentheses for nested expressions.
	if level > 0 {
		return fmt.Sprintf("(%s %s %s)", left, b.Op, right)
	}
	return fmt.Sprintf("%s %s %s", left, b.Op, right)
}

// NotNode represents a "NOT" operation.
type NotNode struct {
	Child Node
}

func (n *NotNode) ToSQL(level int) string {
	child := n.Child.ToSQL(level + 1)
	// For a NOT node, always add parentheses for nested expressions.
	if level > 0 {
		return fmt.Sprintf("(%s %s)", opNot, child)
	}
	return fmt.Sprintf("%s %s", opNot, child)
}

// ConditionNode represents a simple binary condition like "field = value".
type ConditionNode struct {
	Field, Op, Value string
}

func (c *ConditionNode) ToSQL(_ int) string {
	return fmt.Sprintf("%s %s %s", c.Field, c.Op, c.Value)
}

// InNode represents an IN operator condition.
type InNode struct {
	Field  string
	Values []string
}

func (i *InNode) ToSQL(_ int) string {
	return fmt.Sprintf("%s %s (%s)", i.Field, opIn, strings.Join(i.Values, ", "))
}

// ParenNode represents an expression that was explicitly parenthesized in the input.
type ParenNode struct {
	Child Node
}

func (p *ParenNode) ToSQL(level int) string {
	// Always emit the surrounding parentheses regardless of level.
	// We call Child.ToSQL with level 0 so that inner nodes don't remove their grouping.
	return fmt.Sprintf("(%s)", p.Child.ToSQL(0))
}
