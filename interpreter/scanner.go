package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type scanner struct {
	fileScanner      *bufio.Scanner
	lines            [][]rune
	line             int
	tokenStartColumn int
	currentColumn    int
	tokens           []Token
}

type ScanError struct {
	Line    int
	Column  int
	Message string
}

func (s ScanError) Error() string {
	return fmt.Sprintf("[%d:%d]: %s", s.Line+1, s.Column+1, s.Message)
}

func Scan(source io.Reader) ([]Token, [][]rune, error) {
	fileScanner := bufio.NewScanner(source)

	srcScanner := &scanner{
		fileScanner: fileScanner,
		line:        -1,
	}

	err := srcScanner.scan()

	return srcScanner.tokens, srcScanner.lines, err
}

func (s *scanner) scan() error {
	c, err := s.nextCharacter()
	if err != nil {
		return err
	}

	for c != '\000' {
		switch c {
		case '+':
			s.addToken(PLUS, nil)
		case '-':
			s.addToken(MINUS, nil)
		case '*':
			s.addToken(ASTERISK, nil)
		case '%':
			s.addToken(PERCENT, nil)

		case '(':
			s.addToken(OPEN_PAREN, nil)
		case ')':
			s.addToken(CLOSE_PAREN, nil)

		case '/':
			if s.match('/') {
				s.comment()
			} else {
				s.addToken(SLASH, nil)
			}

		case ' ', '\t':
			s.tokenStartColumn = s.currentColumn + 1
			break

		default:
			if isDigit(c) {
				s.number()
			} else {
				return s.newError(fmt.Sprintf("Unexpected character '%c'.", c))
			}
		}

		c, err = s.nextCharacter()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *scanner) number() {
	for isDigit(s.peek()) {
		s.nextCharacter()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.nextCharacter()
		for isDigit(s.peek()) {
			s.nextCharacter()
		}
	}

	value, _ := strconv.ParseFloat(string(s.lines[s.line][s.tokenStartColumn:s.currentColumn+1]), 64)
	s.addToken(NUMBER, value)
}

func (s *scanner) comment() {
	for s.peek() != '\n' {
		s.nextCharacter()
	}
}

func (s *scanner) nextCharacter() (rune, error) {
	s.currentColumn++
	if s.line == -1 || s.currentColumn == len(s.lines[s.line]) {
		notDone, err := s.nextLine()
		if !notDone {
			return '\000', err
		}
	}

	return s.lines[s.line][s.currentColumn], nil
}

func (s *scanner) peek() rune {
	if s.currentColumn+1 == len(s.lines[s.line]) {
		return '\n'
	}

	return s.lines[s.line][s.currentColumn+1]
}

func (s *scanner) peekNext() rune {
	if s.currentColumn+2 == len(s.lines[s.line]) {
		return '\n'
	}

	return s.lines[s.line][s.currentColumn+2]
}

func (s *scanner) match(char rune) bool {
	if s.peek() != char {
		return false
	}
	s.nextCharacter()
	return true
}

func (s *scanner) nextLine() (bool, error) {
	if !s.fileScanner.Scan() {
		return false, s.fileScanner.Err()
	}
	s.lines = append(s.lines, []rune(s.fileScanner.Text()))
	s.line++
	s.currentColumn = 0
	s.tokenStartColumn = 0

	return true, nil
}

func (s *scanner) addToken(tokenType TokenType, literal any) {
	s.tokens = append(s.tokens, Token{
		Line:    s.line,
		Column:  s.tokenStartColumn,
		Type:    tokenType,
		Lexeme:  string(s.lines[s.line][s.tokenStartColumn : s.currentColumn+1]),
		Literal: literal,
	})
	s.tokenStartColumn = s.currentColumn + 1
}

func (s *scanner) newError(msg string) error {
	return ScanError{
		Line:    s.line,
		Column:  s.currentColumn,
		Message: msg,
	}
}

func isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}
