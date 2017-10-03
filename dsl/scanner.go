package dsl

import (
	"bytes"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Scanner is a marionette DSL tokenizer.
type Scanner struct {
	data []byte
}

// NewScanner returns a new instance of Scanner.
func NewScanner(data []byte) *Scanner {
	data = bytes.Replace(data, []byte{0}, []byte("\uFFFD"), -1)
	data = bytes.Replace(data, []byte{'\f'}, []byte{'\n'}, -1)
	data = bytes.Replace(data, []byte{'\r', '\n'}, []byte{'\n'}, -1)
	return &Scanner{data: data}
}

// Scan returns the next token from the reader.
func (s *Scanner) Scan() (tok Token, lit string, pos Pos) {
	for {
		// Read next code point.
		ch := s.read()
		pos := s.pos()

		// If whitespace code point found, then consume all contiguous whitespace.
		if isWhitespace(ch) {
			return s.scanWhitespace()
		} else if isDigit(ch) {
			s.unread(1)
			return s.scanNumeric(pos)
		} else if isNameStart(ch) {
			return s.scanIdent()
		}

		// Check against individual code points next.
		switch ch {
		case eof:
			return EOF, "", pos
		case '"', '\'':
			return s.scanString()
		case ',':
			return COMMA, string(ch), pos
		case '-':
			s.unread(1)
			return s.scanNumeric(pos)
		case ':':
			return COLON, string(ch), pos
		case '(':
			return LPAREN, string(ch), pos
		case ')':
			return RPAREN, string(ch), pos
		case '.':
			return DOT, string(ch), pos
		default:
			return ILLEGAL, string(ch), pos
		}
	}
}

// ScanIgnoreWhitespace returns the next non-whitespace token.
func (s *Scanner) ScanIgnoreWhitespace() (tok Token, lit string, pos Pos) {
	for {
		if tok, lit := s.Scan(); tok != WS {
			return tok, lit
		}
	}
}

// Peek returns the next token without moving the scanner forward.
func (s *Scanner) Peek() (tok Token, lit string, pos Pos) {
	panic("TODO")
}

// PeekIgnoreWhitespace returns the next non-whitespace token without moving the scanner forward.
func (s *Scanner) PeekIgnoreWhitespace() (tok Token, lit string, pos Pos) {
	for {
		if tok, lit := s.Peek(); tok != WS {
			return tok, lit
		}
	}
}

// scanWhitespace consumes the current code point and all subsequent whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit []byte) {
	pos := s.pos()
	var buf bytes.Buffer
	_, _ = buf.WriteRune(s.curr())
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread(1)
			break
		}
		_, _ = buf.WriteRune(ch)
	}
	return &Token{Tok: WhitespaceToken, Value: buf.String(), Pos: pos}
}

// scanString consumes a quoted string. (§4.3.4)
//
// This assumes that the current token is a single or double quote.
// This function consumes all code points and escaped code points up until
// a matching, unescaped ending quote.
// An EOF closes out a string but does not return an error.
// A newline will close a string and returns a bad-string token.
func (s *Scanner) scanString() (tok Token, lit []byte) {
	pos, ending := s.pos(), s.curr()
	var buf bytes.Buffer
	for {
		ch := s.read()
		if ch == eof || ch == ending {
			return &Token{Tok: StringToken, Value: buf.String(), Ending: ending, Pos: pos}
		} else if ch == '\n' {
			s.unread(1)
			return &Token{Tok: BadStringToken, Pos: pos}
		} else if ch == '\\' {
			// If the next code point is EOF then do nothing.
			// If it is a newline then consume it.
			if next := s.read(); next == eof {
				continue
			} else if next == '\n' {
				_, _ = buf.WriteRune(next)
				continue
			}
			s.unread(1)

			// If it is an escape then consume the escaped code point.
			if s.peekEscape() {
				_, _ = buf.WriteRune(s.scanEscape())
				continue
			}
		}

		// Append anything else to the buffer.
		_, _ = buf.WriteRune(ch)
	}
}

// scanNumeric consumes a numeric token.
//
// This assumes that the current token is a +, -, . or digit.
func (s *Scanner) scanNumeric(pos Pos) (tok Token, lit []byte) {
	num, typ, repr := s.scanNumber()

	// If the number is immediately followed by an identifier then scan dimension.
	if s.read(); s.peekIdent() {
		unit := s.scanName()
		return &Token{Tok: DimensionToken, Type: typ, Value: repr + unit, Number: num, Unit: unit, Pos: pos}
	} else {
		s.unread(1)
	}

	// If the number is followed by a percent sign then return a percentage.
	if ch := s.read(); ch == '%' {
		return &Token{Tok: PercentageToken, Type: typ, Value: repr + "%", Number: num, Pos: pos}
	} else {
		s.unread(1)
	}

	// Otherwise return a number token.
	return &Token{Tok: NumberToken, Type: typ, Value: repr, Number: num, Pos: pos}
}

// scanNumber consumes a number.
func (s *Scanner) scanNumber() (num float64, typ, repr string) {
	var buf bytes.Buffer
	typ = "integer"

	// If initial code point is + or - then store it.
	if ch := s.read(); ch == '+' || ch == '-' {
		_, _ = buf.WriteRune(ch)
	} else {
		s.unread(1)
	}

	// Read as many digits as possible.
	_, _ = buf.WriteString(s.scanDigits())

	// If next code points are a full stop and digit then consume them.
	if ch0 := s.read(); ch0 == '.' {
		if ch1 := s.read(); isDigit(ch1) {
			typ = "number"
			_, _ = buf.WriteRune(ch0)
			_, _ = buf.WriteRune(ch1)
			_, _ = buf.WriteString(s.scanDigits())
		} else {
			s.unread(2)
		}
	} else {
		s.unread(1)
	}

	// Consume scientific notation (e0, e+0, e-0, E0, E+0, E-0).
	if ch0 := s.read(); ch0 == 'e' || ch0 == 'E' {
		if ch1 := s.read(); ch1 == '+' || ch1 == '-' {
			if ch2 := s.read(); isDigit(ch2) {
				typ = "number"
				_, _ = buf.WriteRune(ch0)
				_, _ = buf.WriteRune(ch1)
				_, _ = buf.WriteRune(ch2)
			} else {
				s.unread(3)
			}
		} else if isDigit(ch1) {
			typ = "number"
			_, _ = buf.WriteRune(ch0)
			_, _ = buf.WriteRune(ch1)
		} else {
			s.unread(2)
		}
	} else {
		s.unread(1)
	}

	// Parse number.
	num, _ = strconv.ParseFloat(buf.String(), 64)
	repr = buf.String()
	return
}

// scanDigits consume a contiguous series of digits.
func (s *Scanner) scanDigits() string {
	var buf bytes.Buffer
	for {
		if ch := s.read(); isDigit(ch) {
			_, _ = buf.WriteRune(ch)
		} else {
			s.unread(1)
			break
		}
	}
	return buf.String()
}

func (s *Scanner) scanName() string {
	var buf bytes.Buffer
	s.unread(1)
	for {
		if ch := s.read(); isName(ch) {
			_, _ = buf.WriteRune(ch)
		} else if s.peekEscape() {
			_, _ = buf.WriteRune(s.scanEscape())
		} else {
			s.unread(1)
			return buf.String()
		}
	}
}

// scanIdent consumes a ident-like token.
// This function can return an ident, function, url, or bad-url.
func (s *Scanner) scanIdent() (tok Token, lit []byte) {
	pos := s.pos()
	v := s.scanName()

	// Check if this is the start of a url token.
	if strings.ToLower(v) == "url" {
		if ch := s.read(); ch == '(' {
			return s.scanURL(pos)
		}
		s.unread(1)
	} else if ch := s.read(); ch == '(' {
		return &Token{Tok: FunctionToken, Value: v, Pos: pos}
	}
	s.unread(1)

	return &Token{Tok: IdentToken, Value: v, Pos: pos}
}

func (s *Scanner) scanEscape() rune {
	var buf bytes.Buffer
	ch := s.read()
	if isHexDigit(ch) {
		_, _ = buf.WriteRune(ch)
		for i := 0; i < 5; i++ {
			if next := s.read(); next == eof || isWhitespace(next) {
				break
			} else if !isHexDigit(next) {
				s.unread(1)
				break
			} else {
				_, _ = buf.WriteRune(next)
			}
		}
		v, _ := strconv.ParseInt(buf.String(), 16, 0)
		return rune(v)
	} else if ch == eof {
		return '\uFFFD'
	} else {
		return ch
	}
}

func (s *Scanner) peekEscape() bool {
	// If the current code point is not a backslash then this is not an escape.
	if s.curr() != '\\' {
		return false
	}

	// If the next code point is a newline then this is not an escape.
	next := s.read()
	s.unread(1)
	return next != '\n'
}

func (s *Scanner) peekIdent() bool {
	if s.curr() == '-' {
		ch := s.read()
		s.unread(1)
		return isNameStart(ch) || s.peekEscape()
	} else if isNameStart(s.curr()) {
		return true
	} else if s.curr() == '\\' && s.peekEscape() {
		return true
	}
	return false
}

func (s *Scanner) peekNumber() bool {
	// If this is a plus or minus followed by a digit or a dot+digit, return true.
	// If this is a dot followed by a digit then return true.
	switch s.curr() {
	case '+', '-':
		ch0, ch1 := s.read(), s.read()
		s.unread(2)
		if isDigit(ch0) || (ch0 == '.' && isDigit(ch1)) {
			return true
		}
	case '.':
		ch0 := s.read()
		s.unread(1)
		if isDigit(ch0) {
			return true
		}
	}

	// Note: We don't check for digits here because its done in Scan().

	// Anything else is not a number.
	return false
}

func (s *Scanner) read() rune {
	if s.i >= len(s.data) {
		return eof
	}
	ch, sz := utf8.DecodeRune(s.data)
	s.i += sz

	// Track scanner position.
	if ch == '\n' {
		s.pos.Line++
		s.pos.Char = 0
	} else {
		s.pos.Char++
	}

	return ch
}

func (s *Scanner) peak() rune {
	if s.i >= len(s.data) {
		return eof
	}
	ch, _ := utf8.DecodeRune(s.data)
	return ch
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

// isHexDigit returns true if the rune is a hex digit.
func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// isNonASCII returns true if the rune is greater than U+0080.
func isNonASCII(ch rune) bool {
	return ch >= '\u0080'
}

// isNameStart returns true if the rune can start a name.
func isNameStart(ch rune) bool {
	return isLetter(ch) || isNonASCII(ch) || ch == '_'
}

// isName returns true if the character is a name code point.
func isName(ch rune) bool {
	return isNameStart(ch) || isDigit(ch) || ch == '-'
}

// eof represents an EOF file byte.
var eof rune = -1
