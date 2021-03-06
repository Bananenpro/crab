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
			if s.match('=') {
				s.addToken(PLUS_EQUAL, nil)
			} else if s.match('+') {
				s.addToken(PLUS_PLUS, nil)
			} else {
				s.addToken(PLUS, nil)
			}
		case '-':
			if s.match('=') {
				s.addToken(MINUS_EQUAL, nil)
			} else if s.match('-') {
				s.addToken(MINUS_MINUS, nil)
			} else {
				s.addToken(MINUS, nil)
			}
		case '*':
			if s.match('=') {
				s.addToken(ASTERISK_EQUAL, nil)
			} else if s.match('*') {
				if s.match('=') {
					s.addToken(ASTERISK_ASTERISK_EQUAL, nil)
				} else {
					s.addToken(ASTERISK_ASTERISK, nil)
				}
			} else {
				s.addToken(ASTERISK, nil)
			}
		case '%':
			if s.match('=') {
				s.addToken(PERCENT_EQUAL, nil)
			} else {
				s.addToken(PERCENT, nil)
			}

		case '(':
			s.addToken(OPEN_PAREN, nil)
		case ')':
			s.addToken(CLOSE_PAREN, nil)
		case '{':
			s.addToken(OPEN_BRACE, nil)
		case '}':
			s.addToken(CLOSE_BRACE, nil)
		case '[':
			s.addToken(OPEN_BRACKET, nil)
		case ']':
			s.addToken(CLOSE_BRACKET, nil)

		case '=':
			if s.match('=') {
				s.addToken(EQUAL_EQUAL, nil)
			} else {
				s.addToken(EQUAL, nil)
			}
		case '!':
			if s.match('=') {
				s.addToken(BANG_EQUAL, nil)
			} else {
				s.addToken(BANG, nil)
			}
		case '<':
			if s.match('=') {
				s.addToken(LESS_EQUAL, nil)
			} else {
				s.addToken(LESS, nil)
			}
		case '>':
			if s.match('=') {
				s.addToken(GREATER_EQUAL, nil)
			} else {
				s.addToken(GREATER, nil)
			}

		case '&':
			if s.match('&') {
				s.addToken(AND, nil)
			} else {
				return s.newError(fmt.Sprintf("Unexpected character '%c'.", c))
			}
		case '|':
			if s.match('|') {
				s.addToken(OR, nil)
			} else {
				return s.newError(fmt.Sprintf("Unexpected character '%c'.", c))
			}
		case '^':
			if s.match('^') {
				s.addToken(XOR, nil)
			} else {
				return s.newError(fmt.Sprintf("Unexpected character '%c'.", c))
			}

		case '/':
			if s.match('/') {
				s.comment()
			} else if s.match('*') {
				err := s.blockComment()
				if err != nil {
					return err
				}
			} else if s.match('=') {
				s.addToken(SLASH_EQUAL, nil)
			} else {
				s.addToken(SLASH, nil)
			}

		case ';':
			s.addToken(SEMICOLON, nil)
		case ',':
			s.addToken(COMMA, nil)
		case '?':
			s.addToken(QUESTION_MARK, nil)
		case ':':
			s.addToken(COLON, nil)

		case '"':
			err := s.string()
			if err != nil {
				return err
			}

		case ' ', '\t':
			break

		default:
			if isDigit(c) {
				s.number()
			} else if isAlpha(c) {
				s.identifier()
			} else {
				return s.newError(fmt.Sprintf("Unexpected character '%c'.", c))
			}
		}

		c, err = s.nextCharacter()
		if err != nil {
			return err
		}
		s.tokenStartColumn = s.currentColumn
	}

	s.tokens = append(s.tokens, Token{
		Line:   s.line,
		Column: len(s.lines[s.line]),
		Type:   EOF,
		Lexeme: "",
	})

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

func (s *scanner) string() error {
	characters := make([]rune, 0)
	for s.peek() != '"' && s.peek() != '\n' {
		c, _ := s.nextCharacter()
		if c == '\\' {
			c, _ = s.nextCharacter()
			switch c {
			case 'r':
				characters = append(characters, '\r')
			case 'n':
				characters = append(characters, '\n')
			case 't':
				characters = append(characters, '\t')
			case 'e':
				characters = append(characters, '\x1b')
			case '\\':
				characters = append(characters, '\\')
			case '"':
				characters = append(characters, '"')
			default:
				return s.newError("Unknown escape sequence.")
			}
		} else {
			characters = append(characters, c)
		}
	}
	if !s.match('"') {
		return s.newError("Unterminated string.")
	}
	s.addToken(STRING, string(characters))
	return nil
}

func (s *scanner) identifier() {
	for isAlphaNum(s.peek()) {
		s.nextCharacter()
	}

	name := string(s.lines[s.line][s.tokenStartColumn : s.currentColumn+1])

	switch name {
	case "true":
		s.addToken(TRUE, true)
	case "false":
		s.addToken(FALSE, false)
	case "var":
		s.addToken(VAR, nil)
	case "func":
		s.addToken(FUNC, nil)
	case "if":
		s.addToken(IF, nil)
	case "else":
		s.addToken(ELSE, nil)
	case "while":
		s.addToken(WHILE, nil)
	case "for":
		s.addToken(FOR, nil)
	case "break":
		s.addToken(BREAK, nil)
	case "continue":
		s.addToken(CONTINUE, nil)
	case "return":
		s.addToken(RETURN, nil)
	case "try":
		s.addToken(TRY, nil)
	case "catch":
		s.addToken(CATCH, nil)
	case "throw":
		s.addToken(THROW, nil)
	case "throws":
		s.addToken(THROWS, nil)
	default:
		s.addToken(IDENTIFIER, nil)
	}
}

func (s *scanner) comment() {
	for s.peek() != '\n' {
		s.nextCharacter()
	}
}

func (s *scanner) blockComment() error {
	nestingLevel := 1
	for nestingLevel > 0 {
		c, err := s.nextCharacter()
		if c == '\000' || err != nil {
			return err
		}
		if c == '/' && s.match('*') {
			nestingLevel++
			continue
		}
		if c == '*' && s.match('/') {
			nestingLevel--
			continue
		}
	}
	return nil
}

func (s *scanner) nextCharacter() (rune, error) {
	s.currentColumn++
	for s.line == -1 || s.currentColumn >= len(s.lines[s.line]) {
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
}

func isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}

func isAlpha(char rune) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char == '_'
}

func isAlphaNum(char rune) bool {
	return isDigit(char) || isAlpha(char)
}

type ScanError struct {
	Line     int
	LineText []rune
	Column   int
	Message  string
}

func (s ScanError) Error() string {
	return generateErrorText(s.Message, s.LineText, s.Line, s.Column, s.Column+1)
}

func (s *scanner) newError(msg string) error {
	return ScanError{
		Line:     s.line,
		LineText: s.lines[s.line],
		Column:   s.currentColumn,
		Message:  msg,
	}
}
