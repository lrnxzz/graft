package command

import (
	"errors"
)

// A Mark is a run of the line and whether the tree made sense of it. The chat
// colours by these, so a wrong block name goes red the moment it stops being a
// prefix of a real one rather than when the line is sent.
type Mark struct {
	From int
	To   int
	Good bool
}

// Reading is the whole line as the tree sees it: what parses, what does not, and
// what finishing the last word would add.
type Reading struct {
	Marks []Mark
	Ghost string
}

// Check reads the line without running anything. A trailing word is judged
// gently: while it could still become something real it is not wrong yet, only
// unfinished, and colouring it red as you type would be noise.
func (t *Tree) Check(line string, call Call) Reading {
	tokens := Read(line)
	if tokens.Len() == 0 {
		return Reading{}
	}

	call.tokens = tokens
	call.held = map[string]any{}

	read := &reading{
		tree:   t,
		call:   &call,
		stream: Over(tokens),
	}

	read.walk()

	return Reading{
		Marks: read.marks,
		Ghost: read.ghost,
	}
}

// reading carries the walk, which needs more between its steps than a return
// value would hold comfortably
type reading struct {
	tree   *Tree
	call   *Call
	stream *Stream

	marks []Mark
	ghost string
}

func (r *reading) walk() {
	node := r.root()
	if node == nil {
		return
	}

	node = r.children(node)
	if node == nil {
		return
	}

	r.arguments(node)
	r.leftovers()
}

func (r *reading) root() *Node {
	token, held := r.stream.Take()
	if !held {
		return nil
	}

	node := pick(r.tree.roots, token.Text)
	if node != nil {
		r.good(token)

		return node
	}

	r.unknown(token, words(r.tree.roots))

	return nil
}

func (r *reading) children(node *Node) *Node {
	for len(node.under) > 0 {
		mark := r.stream.Mark()

		token, held := r.stream.Take()
		if !held {
			return node
		}

		child := pick(node.under, token.Text)
		if child == nil {
			r.stream.Reset(mark)
			r.unknown(token, words(node.under))

			return nil
		}

		r.good(token)
		node = child
	}

	return node
}

func (r *reading) arguments(node *Node) {
	for _, argument := range node.takes {
		if r.stream.Done() {
			return
		}

		mark := r.stream.Mark()

		if err := argument.take(r.stream, r.call); err != nil {
			r.refused(argument, err, mark)

			return
		}

		r.spans(mark, r.stream.Mark(), true)
	}
}

// leftovers are words past everything the node declared, which no amount of
// typing will make right
func (r *reading) leftovers() {
	for {
		token, held := r.stream.Take()
		if !held {
			return
		}

		r.bad(token)
	}
}

// refused is a word the parameter would not take. While the cursor is still on
// it and it could grow into something the parameter wants, it is left alone.
func (r *reading) refused(argument Argument, err error, mark int) {
	r.stream.Reset(mark)

	token, held := r.stream.Take()
	if !held {
		return
	}

	if r.typing(token) && r.growing(argument, token.Text) {
		r.offer(argument, token.Text)

		return
	}

	var refusal *Refusal
	if errors.As(err, &refusal) && refusal.To > refusal.From {
		r.marks = append(r.marks, Mark{
			From: refusal.From,
			To:   refusal.To,
		})

		return
	}

	r.bad(token)
}

// growing reports whether the word could still become one the parameter takes
func (r *reading) growing(argument Argument, typed string) bool {
	return len(argument.Offer(*r.call, typed)) > 0
}

func (r *reading) offer(argument Argument, typed string) {
	offers := argument.Offer(*r.call, typed)
	if len(offers) == 0 {
		return
	}

	r.ghost = offers[0][len(typed):]
}

// unknown is a word no declared child matches. The last word of a line is still
// being typed, so it is only wrong once something follows it.
func (r *reading) unknown(token Token, among []string) {
	if r.typing(token) && prefixed(among, token.Text) != "" {
		r.ghost = prefixed(among, token.Text)[len(token.Text):]
		r.good(token)

		return
	}

	r.bad(token)
}

// typing reports whether the cursor could still be on this word. A word the
// line ends on is being written; one with a space after it has been left.
func (r *reading) typing(token Token) bool {
	return token.To >= len(r.stream.Tokens().Line())
}

func (r *reading) good(token Token) {
	r.marks = append(r.marks, Mark{
		From: token.From,
		To:   token.To,
		Good: true,
	})
}

func (r *reading) bad(token Token) {
	r.marks = append(r.marks, Mark{
		From: token.From,
		To:   token.To,
	})
}

func (r *reading) spans(from, to int, good bool) {
	tokens := r.stream.Tokens()

	for at := from; at < to; at++ {
		token, held := tokens.At(at)
		if !held {
			return
		}

		r.marks = append(r.marks, Mark{
			From: token.From,
			To:   token.To,
			Good: good,
		})
	}
}

func prefixed(among []string, typed string) string {
	for _, word := range among {
		if len(word) > len(typed) && word[:len(typed)] == typed {
			return word
		}
	}

	return ""
}
