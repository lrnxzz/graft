package plugin

import "github.com/dop251/goja"

// A Word is one name a plugin can write, and what saying it builds. Components,
// goals, markers and their combinators all land on the same javascript object,
// so the one catalogue is what keeps two corners from claiming the same name.
type Word struct {
	Name  string
	Build func(*Runtime, goja.FunctionCall) goja.Value
}

func Words() []Word {
	var words []Word
	words = append(words, componentWords()...)
	words = append(words, goalWords()...)
	words = append(words, markerWords()...)

	return words
}

func (r *Runtime) spoken(build func(*Runtime, goja.FunctionCall) goja.Value) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		return build(r, call)
	}
}
