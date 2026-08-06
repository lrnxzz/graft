package command

import "strings"

// A Stream is a place in the token list. Every parameter reads from one and
// leaves it past whatever it took, so a parameter that wants three tokens for a
// coordinate is no different from one that wants a single word.
type Stream struct {
	tokens Tokens
	at     int
}

func Over(tokens Tokens) *Stream {
	return &Stream{
		tokens: tokens,
	}
}

func (s *Stream) Done() bool {
	return s.at >= s.tokens.Len()
}

func (s *Stream) Peek() (Token, bool) {
	return s.tokens.At(s.at)
}

func (s *Stream) Take() (Token, bool) {
	token, held := s.tokens.At(s.at)
	if held {
		s.at++
	}

	return token, held
}

// Skip steps over a separator that is allowed but not required, so 10,64,-3 and
// 10 64 -3 are the same three numbers
func (s *Stream) Skip(shape Shape) bool {
	token, held := s.tokens.At(s.at)
	if !held || token.Shape != shape {
		return false
	}

	s.at++

	return true
}

// Rest is everything left as the player wrote it, which is what a trailing
// message parameter wants: the spacing back the way it was typed
func (s *Stream) Rest() string {
	token, held := s.tokens.At(s.at)
	if !held {
		return ""
	}

	last, _ := s.tokens.At(s.tokens.Len() - 1)
	s.at = s.tokens.Len()

	return strings.TrimSpace(s.tokens.Line()[token.From:last.To])
}

func (s *Stream) Mark() int {
	return s.at
}

func (s *Stream) Reset(mark int) {
	s.at = mark
}

// Span is where the stream is standing, for an error that wants to point
func (s *Stream) Span() (int, int) {
	token, held := s.tokens.At(s.at)
	if held {
		return token.From, token.To
	}

	return len(s.tokens.Line()), len(s.tokens.Line()) + 1
}

func (s *Stream) Tokens() Tokens {
	return s.tokens
}
