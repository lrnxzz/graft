package command

import (
	"errors"
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
	Spare() bool
	Offer(call Call, typed string) []string

	take(stream *Stream, call *Call) error
}

// A Param is one typed argument. It knows how to read itself out of the token
// stream and how to offer what it would accept, which is the same knowledge the
// chat needs to complete a half-typed line.
type Param[T any] struct {
	label string
	spare bool
	fallb T

	read  func(*Stream, Call) (T, error)
	offer func(Call, string) []string
}

func (p Param[T]) Label() string {
	return p.label
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

// word is the common single-token parameter: take one token, refuse it as
// wanted when parse will not have it
func word[T any](label, wanted string, parse func(string) (T, error)) Param[T] {
	read := func(stream *Stream, _ Call) (T, error) {
		var zero T

		token, held := stream.Take()
		if !held {
			return zero, missing(label)
		}

		value, err := parse(token.Text)
		if err != nil {
			return zero, refuse(label, token, wanted)
		}

		return value, nil
	}

	return Param[T]{
		label: label,
		read:  read,
	}
}

// Text is any single word, or a quoted run held together
func Text(label string) Param[string] {
	return word(label, "a word", func(text string) (string, error) {
		return text, nil
	})
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
		read:  read,
	}
}

func Whole(label string) Param[int] {
	return word(label, "a whole number", strconv.Atoi)
}

func Number(label string) Param[float64] {
	return word(label, "a number", func(text string) (float64, error) {
		return strconv.ParseFloat(text, 64)
	})
}

func Flag(label string) Param[bool] {
	param := word(label, "yes or no", func(text string) (bool, error) {
		switch strings.ToLower(text) {
		case "true", "yes", "on":
			return true, nil
		case "false", "no", "off":
			return false, nil
		}

		return false, errors.New("refused")
	})

	param.offer = func(_ Call, typed string) []string {
		return matching([]string{"yes", "no"}, typed)
	}

	return param
}

func Span(label string) Param[time.Duration] {
	return word(label, "a length of time such as 30s or 2m", time.ParseDuration)
}

// OneOf accepts nothing but the words it was given, and offers them
func OneOf(label string, choices ...string) Param[string] {
	param := word(label, "one of "+strings.Join(choices, ", "), func(text string) (string, error) {
		for _, choice := range choices {
			if strings.EqualFold(text, choice) {
				return choice, nil
			}
		}

		return "", errors.New("refused")
	})

	param.offer = func(_ Call, typed string) []string {
		return matching(choices, typed)
	}

	return param
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

// matching keeps only what the half-typed word could still become
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
