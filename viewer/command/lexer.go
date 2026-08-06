package command

import (
	"fmt"
	"strings"
	"unicode"
)

// Read breaks a line into tokens. It never fails on an unfinished line: a lone
// opening quote is a quoted token that runs to the end, because half a line is
// exactly what completion is asked about while the player is still typing.
func Read(line string) Tokens {
	lexer := lexer{
		line: []rune(line),
	}

	return Tokens{
		line:  line,
		found: lexer.run(),
	}
}

type lexer struct {
	line []rune
	at   int
}

func (l *lexer) run() []Token {
	var found []Token

	for {
		l.skipBlanks()
		if l.done() {
			return found
		}

		found = append(found, l.next())
	}
}

func (l *lexer) next() Token {
	switch l.line[l.at] {
	case '"':
		return l.quoted()
	case '~':
		return l.offset(Relative)
	case '^':
		return l.offset(Local)
	case '@':
		return l.marked(Selector)
	case '[':
		return l.single(Opens)
	case ']':
		return l.single(Closes)
	case '=':
		return l.single(Equals)
	case ',':
		return l.single(Comma)
	default:
		return l.bare()
	}
}

func (l *lexer) single(shape Shape) Token {
	from := l.at
	l.at++

	return Token{
		Shape: shape,
		Text:  string(l.line[from:l.at]),
		From:  from,
		To:    l.at,
	}
}

// a quoted run keeps its escapes resolved, so a parameter never sees a backslash
// that was only there to carry the quote
func (l *lexer) quoted() Token {
	from := l.at
	l.at++

	var text strings.Builder

	for !l.done() {
		letter := l.line[l.at]

		if letter == '\\' && l.at+1 < len(l.line) {
			text.WriteRune(l.line[l.at+1])
			l.at += 2

			continue
		}
		if letter == '"' {
			l.at++

			break
		}

		text.WriteRune(letter)
		l.at++
	}

	return Token{
		Shape: Quoted,
		Text:  text.String(),
		From:  from,
		To:    l.at,
	}
}

// ~ and ^ carry an optional signed number: "~" is here, "~-2" is two back
func (l *lexer) offset(shape Shape) Token {
	from := l.at
	l.at++

	if !l.done() && (l.line[l.at] == '-' || l.line[l.at] == '+') {
		l.at++
	}
	for !l.done() && (unicode.IsDigit(l.line[l.at]) || l.line[l.at] == '.') {
		l.at++
	}

	return Token{
		Shape: shape,
		Text:  string(l.line[from:l.at]),
		From:  from,
		To:    l.at,
	}
}

func (l *lexer) marked(shape Shape) Token {
	from := l.at
	l.at++

	for !l.done() && isBare(l.line[l.at]) {
		l.at++
	}

	return Token{
		Shape: shape,
		Text:  string(l.line[from:l.at]),
		From:  from,
		To:    l.at,
	}
}

// a bare run is a word unless every character of it reads as a number
func (l *lexer) bare() Token {
	from := l.at
	for !l.done() && isBare(l.line[l.at]) {
		l.at++
	}

	text := string(l.line[from:l.at])

	shape := Bare
	if numeric(text) {
		shape = Numeral
	}

	return Token{
		Shape: shape,
		Text:  text,
		From:  from,
		To:    l.at,
	}
}

func (l *lexer) skipBlanks() {
	for !l.done() && unicode.IsSpace(l.line[l.at]) {
		l.at++
	}
}

func (l *lexer) done() bool {
	return l.at >= len(l.line)
}

func isBare(letter rune) bool {
	if unicode.IsSpace(letter) {
		return false
	}

	return !strings.ContainsRune(`"~^@[]=,`, letter)
}

func numeric(text string) bool {
	if text == "" || text == "-" || text == "+" || text == "." {
		return false
	}

	body := strings.TrimLeft(text, "-+")
	dots := 0

	for _, letter := range body {
		switch {
		case letter == '.':
			dots++
			if dots > 1 {
				return false
			}
		case !unicode.IsDigit(letter):
			return false
		}
	}

	return body != ""
}

// Tokens is a read line, kept beside the text it came from so a span can be
// turned back into something the player recognises.
type Tokens struct {
	line  string
	found []Token
}

func (t Tokens) Len() int {
	return len(t.found)
}

func (t Tokens) At(index int) (Token, bool) {
	if index < 0 || index >= len(t.found) {
		return Token{}, false
	}

	return t.found[index], true
}

func (t Tokens) All() []Token {
	return t.found
}

func (t Tokens) Line() string {
	return t.line
}

// Under answers which token the cursor sits in, and whether it sits in one at
// all. A cursor in the blank after a token is starting the next argument, not
// still editing the last, and completion turns on exactly that difference.
func (t Tokens) Under(cursor int) (int, bool) {
	for index, token := range t.found {
		if cursor >= token.From && cursor <= token.To {
			return index, true
		}
	}

	return len(t.found), false
}

// Caret draws the line with a marker under a span, which is how an error shows
// what it could not read
func (t Tokens) Caret(from, to int) string {
	if to <= from {
		to = from + 1
	}

	return t.line + "\n" + strings.Repeat(" ", from) + strings.Repeat("^", to-from)
}

func (t Tokens) String() string {
	shown := make([]string, 0, len(t.found))
	for _, token := range t.found {
		shown = append(shown, fmt.Sprintf("%s(%q)", token.Shape, token.Text))
	}

	return strings.Join(shown, " ")
}
