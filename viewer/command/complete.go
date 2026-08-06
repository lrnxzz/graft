package command

import "strings"

// An Offer is what the chat may put in place of what is being typed. The span is
// what it replaces, so accepting one is a splice rather than a guess at where the
// word began.
type Offer struct {
	Words []string
	From  int
	To    int
}

func (o Offer) Empty() bool {
	return len(o.Words) == 0
}

// Complete answers what could follow the cursor. It runs the same walk and the
// same parameters as a real call, stopping where the cursor is, so what is
// offered is exactly what would have been accepted.
func (t *Tree) Complete(line string, cursor int, call Call) Offer {
	if cursor > len(line) {
		cursor = len(line)
	}

	tokens := Read(line)
	call.tokens = tokens
	call.held = map[string]any{}

	index, inside := tokens.Under(cursor)

	typed, from, to := "", cursor, cursor
	if inside {
		token, _ := tokens.At(index)
		typed, from, to = token.Text, token.From, token.To
	}

	// still on the first word: every command is a candidate
	if index == 0 {
		return spread(words(t.roots), typed, from, to)
	}

	stream := Over(tokens)

	node, reached := t.descend(stream, index)
	if node == nil {
		return Offer{
			From: from,
			To:   to,
		}
	}

	// the cursor is on the word right after a grouping node, so the children are
	// what may come next
	if len(node.under) > 0 && reached == index {
		return spread(words(node.under), typed, from, to)
	}

	return spread(pending(node, stream, &call, index, typed), typed, from, to)
}

// descend follows child words while they match and while they sit before the
// cursor, so the node reached is the one the cursor belongs to
func (t *Tree) descend(stream *Stream, index int) (*Node, int) {
	if index == 0 {
		return nil, 0
	}

	token, held := stream.Take()
	if !held {
		return nil, 0
	}

	node := pick(t.roots, token.Text)
	if node == nil {
		return nil, 0
	}

	for len(node.under) > 0 && stream.Mark() < index {
		mark := stream.Mark()

		next, is := stream.Take()
		if !is {
			break
		}

		child := pick(node.under, next.Text)
		if child == nil {
			stream.Reset(mark)

			break
		}

		node = child
	}

	return node, stream.Mark()
}

// pending walks the declared arguments until the one the cursor sits on, so a
// parameter that eats three tokens is counted for what it really took
func pending(node *Node, stream *Stream, call *Call, index int, typed string) []string {
	for _, argument := range node.takes {
		if stream.Mark() >= index {
			return argument.Offer(*call, typed)
		}

		mark := stream.Mark()
		if err := argument.take(stream, call); err != nil {
			stream.Reset(mark)

			return argument.Offer(*call, typed)
		}

		// a parameter that took nothing would spin the walk forever
		if stream.Mark() == mark {
			break
		}
	}

	return nil
}

func words(among []*Node) []string {
	said := make([]string, 0, len(among))
	for _, node := range among {
		said = append(said, node.word)
	}

	return said
}

// spread keeps only what the half-typed word could still become
func spread(offers []string, typed string, from, to int) Offer {
	kept := offers
	if typed != "" {
		kept = nil
		lowered := strings.ToLower(typed)

		for _, offer := range offers {
			if strings.HasPrefix(strings.ToLower(offer), lowered) {
				kept = append(kept, offer)
			}
		}
	}

	return Offer{
		Words: kept,
		From:  from,
		To:    to,
	}
}
