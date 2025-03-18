package internal

import (
	"fmt"
	"strconv"
	"strings"
)

type tokenType int

const (
	tIdentifier tokenType = iota
	tString
	tNumber
	tParenOpen
	tParenClose
	tComma
	tOpIn
	tOpNot
	tOpAnd
	tOpOr
	tOpEq
	tOpNe
	tOpGt
	tOpGe
	tOpLt
	tOpLe
)

const (
	parenOpen  = "("
	parenClose = ")"
	comma      = ","
)

type token struct {
	typ tokenType
	val string
}

var keywordTokens = map[string]tokenType{
	"in":  tOpIn,
	"not": tOpNot,
	"and": tOpAnd,
	"or":  tOpOr,
	"eq":  tOpEq,
	"ne":  tOpNe,
	"gt":  tOpGt,
	"ge":  tOpGe,
	"lt":  tOpLt,
	"le":  tOpLe,
}

func tokenize(input string) ([]token, error) {
	var tokens []token
	s := strings.TrimSpace(input)
	i := 0
	for i < len(s) {
		ch := s[i]
		if isWhitespace(ch) {
			i++
			continue
		}
		switch ch {
		case '(':
			tokens = append(tokens, token{tParenOpen, parenOpen})
			i++
		case ')':
			tokens = append(tokens, token{tParenClose, parenClose})
			i++
		case ',':
			tokens = append(tokens, token{tComma, comma})
			i++
		case '\'':
			str, consumed, err := readQuotedString(s[i:])
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token{tString, str})
			i += consumed
		default:
			start := i
			for i < len(s) && !isDelimiter(s[i]) {
				i++
			}
			tokens = append(tokens, classifyWord(s[start:i]))
		}
	}
	return tokens, nil
}

func classifyWord(w string) token {
	lower := strings.ToLower(w)

	if tokType, exists := keywordTokens[lower]; exists {
		return token{tokType, lower}
	}

	if _, err := strconv.ParseFloat(w, 64); err == nil {
		return token{tNumber, w}
	}

	return token{tIdentifier, w}
}

// isWhitespace checks if a character is a whitespace character.
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isDelimiter checks if a character is a delimiter (whitespace, parentheses, comma, or single quote).
func isDelimiter(ch byte) bool {
	return isWhitespace(ch) || ch == '(' || ch == ')' || ch == ',' || ch == '\''
}

// readQuotedString extracts a properly formatted quoted string.
func readQuotedString(input string) (string, int, error) {
	if len(input) < 2 || input[0] != '\'' {
		return "", 0, fmt.Errorf("invalid quoted string")
	}

	var sb strings.Builder
	sb.WriteByte('\'') // Keep the opening quote
	i := 1

	for i < len(input) {
		ch := input[i]

		// Handle escaped single quotes (OData style: '')
		if ch == '\'' {
			if i+1 < len(input) && input[i+1] == '\'' {
				sb.WriteByte('\'') // Append single quote as part of string
				i += 2
				continue
			}
			// End of string
			sb.WriteByte('\'')
			return sb.String(), i + 1, nil
		}

		sb.WriteByte(ch)
		i++
	}

	return "", 0, fmt.Errorf("unclosed string literal: %q", input)
}
