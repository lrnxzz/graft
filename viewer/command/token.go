package command

// A Shape is what a run of characters turned out to be. The lexer decides this
// much and no more: what a word means is the parameter's business, so the same
// token stream serves parsing, completion and the caret under an error.
type Shape int

const (
	Bare Shape = iota
	Quoted
	Numeral
	Relative
	Local
	Selector
	Opens
	Closes
	Equals
	Comma
)

var shapes = [...]string{
	Bare:     "word",
	Quoted:   "quoted text",
	Numeral:  "number",
	Relative: "relative offset",
	Local:    "local offset",
	Selector: "selector",
	Opens:    "[",
	Closes:   "]",
	Equals:   "=",
	Comma:    ",",
}

func (s Shape) String() string {
	if int(s) >= len(shapes) {
		return "unknown"
	}

	return shapes[s]
}

// A Token is a run of the line with where it sat. The span is what lets an error
// point at the offending characters and lets completion know which argument the
// cursor is inside.
type Token struct {
	Shape Shape
	Text  string
	From  int
	To    int
}

// Offset is the number a relative or local token carried, and whether it carried
// one at all: "~" is the current value, "~5" is five past it.
func (t Token) Offset() (string, bool) {
	if t.Shape != Relative && t.Shape != Local {
		return "", false
	}
	if len(t.Text) <= 1 {
		return "", false
	}

	return t.Text[1:], true
}
