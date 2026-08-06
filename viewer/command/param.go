package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// An Argument is what a node takes, with its type rubbed off so a node can hold
// several of different types. The typed side is Param, which is what a handler
// reads back through.
type Argument interface {
	Label() string
	Shape() string
	Spare() bool
	Offer(call Call, typed string) []string

	take(stream *Stream, call *Call) error
}

// A Param is one typed argument. It knows how to read itself out of the token
// stream and how to offer what it would accept, which is the same knowledge the
// chat needs to complete a half-typed line.
type Param[T any] struct {
	label string
	shape string
	spare bool
	fallb T

	read  func(*Stream, Call) (T, error)
	offer func(Call, string) []string
}

func (p Param[T]) Label() string {
	return p.label
}

func (p Param[T]) Shape() string {
	return p.shape
}

func (p Param[T]) Spare() bool {
	return p.spare
}

// Or makes the argument optional, standing in with this value when the line ends
// before it
func (p Param[T]) Or(fallback T) Param[T] {
	p.spare = true
	p.fallb = fallback

	return p
}

// Of reads what this parameter parsed. It is typed by the parameter, so a
// handler never casts and never names the argument twice.
func (p Param[T]) Of(call Call) T {
	held, found := call.held[p.label]
	if !found {
		return p.fallb
	}

	value, is := held.(T)
	if !is {
		return p.fallb
	}

	return value
}

func (p Param[T]) Offer(call Call, typed string) []string {
	if p.offer == nil {
		return nil
	}

	return p.offer(call, typed)
}

func (p Param[T]) take(stream *Stream, call *Call) error {
	value, err := p.read(stream, *call)
	if err != nil {
		return err
	}

	call.held[p.label] = value

	return nil
}

// Text is any single word, or a quoted run held together
func Text(label string) Param[string] {
	read := func(stream *Stream, _ Call) (string, error) {
		token, held := stream.Take()
		if !held {
			return "", missing(label)
		}

		return token.Text, nil
	}

	return Param[string]{
		label: label,
		shape: "text",
		read:  read,
	}
}

// Rest is everything left of the line, spacing and all
func Rest(label string) Param[string] {
	read := func(stream *Stream, _ Call) (string, error) {
		text := stream.Rest()
		if text == "" {
			return "", missing(label)
		}

		return text, nil
	}

	return Param[string]{
		label: label,
		shape: "text...",
		read:  read,
	}
}

func Whole(label string) Param[int] {
	read := func(stream *Stream, _ Call) (int, error) {
		token, held := stream.Take()
		if !held {
			return 0, missing(label)
		}

		value, err := strconv.Atoi(token.Text)
		if err != nil {
			return 0, refuse(label, token, "a whole number")
		}

		return value, nil
	}

	return Param[int]{
		label: label,
		shape: "number",
		read:  read,
	}
}

func Number(label string) Param[float64] {
	read := func(stream *Stream, _ Call) (float64, error) {
		token, held := stream.Take()
		if !held {
			return 0, missing(label)
		}

		value, err := strconv.ParseFloat(token.Text, 64)
		if err != nil {
			return 0, refuse(label, token, "a number")
		}

		return value, nil
	}

	return Param[float64]{
		label: label,
		shape: "number",
		read:  read,
	}
}

var truths = [...]string{"true", "yes", "on"}

func Flag(label string) Param[bool] {
	read := func(stream *Stream, _ Call) (bool, error) {
		token, held := stream.Take()
		if !held {
			return false, missing(label)
		}

		spoken := strings.ToLower(token.Text)
		for _, yes := range truths {
			if spoken == yes {
				return true, nil
			}
		}

		for _, no := range [...]string{"false", "no", "off"} {
			if spoken == no {
				return false, nil
			}
		}

		return false, refuse(label, token, "yes or no")
	}

	offer := func(_ Call, typed string) []string {
		return matching([]string{"yes", "no"}, typed)
	}

	return Param[bool]{
		label: label,
		shape: "yes|no",
		read:  read,
		offer: offer,
	}
}

func Span(label string) Param[time.Duration] {
	read := func(stream *Stream, _ Call) (time.Duration, error) {
		token, held := stream.Take()
		if !held {
			return 0, missing(label)
		}

		value, err := time.ParseDuration(token.Text)
		if err != nil {
			return 0, refuse(label, token, "a length of time such as 30s or 2m")
		}

		return value, nil
	}

	return Param[time.Duration]{
		label: label,
		shape: "duration",
		read:  read,
	}
}

// OneOf accepts nothing but the words it was given, and offers them
func OneOf(label string, choices ...string) Param[string] {
	read := func(stream *Stream, _ Call) (string, error) {
		token, held := stream.Take()
		if !held {
			return "", missing(label)
		}

		for _, choice := range choices {
			if strings.EqualFold(token.Text, choice) {
				return choice, nil
			}
		}

		return "", refuse(label, token, "one of "+strings.Join(choices, ", "))
	}

	offer := func(_ Call, typed string) []string {
		return matching(choices, typed)
	}

	return Param[string]{
		label: label,
		shape: strings.Join(choices, "|"),
		read:  read,
		offer: offer,
	}
}

func missing(label string) error {
	return fmt.Errorf("%s is missing", label)
}

func refuse(label string, token Token, wanted string) error {
	return &Refusal{
		Label:  label,
		Got:    token.Text,
		Wanted: wanted,
		From:   token.From,
		To:     token.To,
	}
}

// A Refusal is an argument the parser could read but not accept, and it carries
// where it sat so the chat can point at it
type Refusal struct {
	Label  string
	Got    string
	Wanted string
	Near   string
	From   int
	To     int
}

func (r *Refusal) Error() string {
	said := fmt.Sprintf("%s: %q is not %s", r.Label, r.Got, r.Wanted)
	if r.Near != "" {
		said += fmt.Sprintf(" — did you mean %s?", r.Near)
	}

	return said
}

func matching(choices []string, typed string) []string {
	if typed == "" {
		return choices
	}

	lowered := strings.ToLower(typed)

	var near []string
	for _, choice := range choices {
		if strings.HasPrefix(strings.ToLower(choice), lowered) {
			near = append(near, choice)
		}
	}

	return near
}
