package interpreter

import (
	"fmt"
	"math"
)

type interpreter struct {
	lines [][]rune
}

func Interpret(program Expr, lines [][]rune) (any, error) {
	interpreter := &interpreter{
		lines: lines,
	}
	return program.Accept(interpreter)
}

func (i *interpreter) VisitLiteral(expr ExprLiteral) (any, error) {
	return expr.Value, nil
}

func (i *interpreter) VisitGrouping(expr ExprGrouping) (any, error) {
	return expr.Expr.Accept(i)
}

func (i *interpreter) VisitUnary(expr ExprUnary) (any, error) {
	right, err := expr.Right.Accept(i)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case MINUS:
		if isNumber(right) {
			return -right.(float64), nil
		}
		return nil, i.newError(fmt.Sprintf("Operand must be a number."), expr.Operator)
	case BANG:
		return !isTruthy(right), nil
	default:
		return nil, i.newError(fmt.Sprintf("Invalid unary operator '%s'.", expr.Operator.Lexeme), expr.Operator)
	}
}

func (i *interpreter) VisitBinary(expr ExprBinary) (any, error) {
	left, err := expr.Left.Accept(i)
	if err != nil {
		return nil, err
	}
	right, err := expr.Right.Accept(i)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case PLUS:
		if isNumber(left, right) {
			return left.(float64) + right.(float64), nil
		} else if anyString(left, right) {
			return fmt.Sprintf("%v%v", left, right), nil
		}
		return nil, i.newError(fmt.Sprintf("Operands must be either both numbers or at least one of them a string."), expr.Operator)
	case MINUS:
		if isNumber(left, right) {
			return left.(float64) - right.(float64), nil
		}
		return nil, i.newError(fmt.Sprintf("Both operands must be numbers."), expr.Operator)
	case ASTERISK:
		if isNumber(left, right) {
			return left.(float64) * right.(float64), nil
		}
		return nil, i.newError(fmt.Sprintf("Both operands must be numbers."), expr.Operator)
	case SLASH:
		if isNumber(left, right) {
			return left.(float64) / right.(float64), nil
		}
		return nil, i.newError(fmt.Sprintf("Both operands must be numbers."), expr.Operator)
	case PERCENT:
		if isNumber(left, right) {
			return math.Mod(left.(float64), right.(float64)), nil
		}
		return nil, i.newError(fmt.Sprintf("Both operands must be numbers."), expr.Operator)

	case EQUAL_EQUAL:
		return left == right, nil
	case BANG_EQUAL:
		return left != right, nil

	case LESS:
		if isNumber(left, right) {
			return left.(float64) < right.(float64), nil
		}
		return nil, i.newError(fmt.Sprintf("Both operands must be numbers."), expr.Operator)
	case LESS_EQUAL:
		if isNumber(left, right) {
			return left.(float64) <= right.(float64), nil
		}
		return nil, i.newError(fmt.Sprintf("Both operands must be numbers."), expr.Operator)
	case GREATER:
		if isNumber(left, right) {
			return left.(float64) > right.(float64), nil
		}
		return nil, i.newError(fmt.Sprintf("Both operands must be numbers."), expr.Operator)
	case GREATER_EQUAL:
		if isNumber(left, right) {
			return left.(float64) >= right.(float64), nil
		}
		return nil, i.newError(fmt.Sprintf("Both operands must be numbers."), expr.Operator)

	default:
		return nil, i.newError(fmt.Sprintf("Invalid binary operator '%s'.", expr.Operator.Lexeme), expr.Operator)
	}
}

func isNumber(values ...any) bool {
	for _, v := range values {
		if _, ok := v.(float64); !ok {
			return false
		}
	}
	return true
}

func anyString(values ...any) bool {
	for _, v := range values {
		if _, ok := v.(string); ok {
			return true
		}
	}
	return false
}

func isTruthy(value any) bool {
	if v, ok := value.(bool); ok {
		return v
	}

	if v, ok := value.(float64); ok {
		return v != 0
	}

	if v, ok := value.(string); ok {
		return len(v) > 0
	}

	return false
}

type RuntimeError struct {
	Token   Token
	Message string
	Line    []rune
}

func (r RuntimeError) Error() string {
	return generateErrorText(r.Message, r.Line, r.Token.Line, r.Token.Column, r.Token.Column+len([]byte(r.Token.Lexeme)))
}

func (i *interpreter) newError(message string, token Token) error {
	return RuntimeError{
		Token:   token,
		Message: message,
		Line:    i.lines[token.Line],
	}
}