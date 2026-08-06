package command

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// A Node is one word of a command and what may follow it: more words, or the
// arguments it takes before it runs. The whole set is a tree of these, so a node
// can be held in a variable and hung under two parents.
type Node struct {
	word  string
	says  string
	takes []Argument
	under []*Node
	runs  func(Call) error
}

// Word starts a node. Nothing else builds one, so a tree is always a tree.
func Word(word string) *Node {
	return &Node{
		word: word,
	}
}

// Says is the one line shown beside the word in help
func (n *Node) Says(says string) *Node {
	n.says = says

	return n
}

// Takes are the arguments read after this word, in order. Optional ones come
// last, because the line is read left to right and nothing can be skipped.
func (n *Node) Takes(takes ...Argument) *Node {
	n.takes = append(n.takes, takes...)

	return n
}

// Under hangs child words below this one, which is what makes a subcommand a
// node rather than a string that happens to hold a space
func (n *Node) Under(under ...*Node) *Node {
	n.under = append(n.under, under...)

	return n
}

func (n *Node) Runs(run func(Call) error) *Node {
	n.runs = run

	return n
}

func (n *Node) Name() string {
	return n.word
}

func (n *Node) Describe() string {
	return n.says
}

func (n *Node) Children() []*Node {
	return n.under
}

// Usage is the line a player is shown. The path already names this node, since
// only the walk down the tree knows how it was reached.
func (n *Node) Usage(path string) string {
	said := strings.Builder{}
	said.WriteString(strings.TrimSpace(path))

	for _, argument := range n.takes {
		if argument.Spare() {
			said.WriteString(" [" + argument.Label() + "]")

			continue
		}

		said.WriteString(" <" + argument.Label() + ">")
	}

	return said.String()
}

var errNotRunnable = errors.New("that word takes a subcommand")

// A Tree is the whole set of commands, and the only thing a caller talks to.
type Tree struct {
	roots []*Node
}

func Rooted(roots ...*Node) *Tree {
	return &Tree{
		roots: roots,
	}
}

func (t *Tree) Roots() []*Node {
	return t.roots
}

// Run reads a line and runs what it names. It reports whether the line named a
// command at all, so a caller can pass an unclaimed line somewhere else.
func (t *Tree) Run(line string, call Call) (bool, error) {
	tokens := Read(line)
	call.tokens = tokens
	call.held = map[string]any{}

	stream := Over(tokens)

	node, path, found := t.walk(stream)
	if !found {
		return false, nil
	}

	if node.runs == nil {
		return true, fmt.Errorf("%s: %w — try %s", path, errNotRunnable, t.children(node, path))
	}

	if err := take(node, stream, &call); err != nil {
		return true, &Failure{
			Usage: node.Usage(path),
			Err:   err,
		}
	}

	return true, node.runs(call)
}

// walk follows words down the tree for as long as they match, and hands back the
// deepest node it reached with the path taken to it
func (t *Tree) walk(stream *Stream) (*Node, string, bool) {
	token, held := stream.Take()
	if !held {
		return nil, "", false
	}

	node := pick(t.roots, token.Text)
	if node == nil {
		return nil, "", false
	}

	path := strings.Builder{}
	path.WriteString(node.word)

	for len(node.under) > 0 {
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
		path.WriteString(" " + child.word)
	}

	return node, path.String(), true
}

func pick(among []*Node, word string) *Node {
	for _, node := range among {
		if strings.EqualFold(node.word, word) {
			return node
		}
	}

	return nil
}

// take reads every declared argument in order. An optional one that finds the
// line already spent is left at its fallback rather than failing.
func take(node *Node, stream *Stream, call *Call) error {
	for _, argument := range node.takes {
		if stream.Done() && argument.Spare() {
			continue
		}

		if err := argument.take(stream, call); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tree) children(node *Node, path string) string {
	words := make([]string, 0, len(node.under))
	for _, child := range node.under {
		words = append(words, child.word)
	}

	sort.Strings(words)

	return path + " " + strings.Join(words, "|")
}

// A Failure is a line that named a command and then would not parse. It carries
// the usage so the player is told what was wanted in the same breath.
type Failure struct {
	Usage string
	Err   error
}

func (f *Failure) Error() string {
	return f.Err.Error()
}

func (f *Failure) Unwrap() error {
	return f.Err
}

// Help is every command in the tree, deepest paths included, sorted
func (t *Tree) Help() []string {
	var said []string

	var walk func(node *Node, path string)
	walk = func(node *Node, path string) {
		here := strings.TrimSpace(path + " " + node.word)

		if node.runs != nil {
			line := node.Usage(here)
			if node.says != "" {
				line += "  —  " + node.says
			}

			said = append(said, line)
		}

		for _, child := range node.under {
			walk(child, here)
		}
	}

	for _, root := range t.roots {
		walk(root, "")
	}

	sort.Strings(said)

	return said
}
