package command

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
)

// A Call is one line being run: what the arguments parsed to, who is running it,
// and the tokens behind it so an error can point back at the text.
type Call struct {
	held   map[string]any
	tokens Tokens
	bot    *agent.Agent
	tell   func(string)
}

func Calling(bot *agent.Agent, tell func(string)) Call {
	return Call{
		held: map[string]any{},
		bot:  bot,
		tell: tell,
	}
}

func (c Call) Bot() *agent.Agent {
	return c.bot
}

// Tell answers the player. A command that refuses or reports goes through here
// rather than the log, because the player is the one who typed it.
func (c Call) Tell(text string) {
	if c.tell != nil {
		c.tell(text)
	}
}

func (c Call) Tokens() Tokens {
	return c.tokens
}

// Standing is where the bot is, which is what a relative offset is relative to
func (c Call) Standing() gocraft.Position {
	if c.bot == nil {
		return gocraft.Position{}
	}

	return c.bot.Snapshot().Position.Floor()
}
