package main

import (
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/plugin"
)

// The bot's events and intents are catalogues rather than switches, so adding one
// is a line here and nothing else. Each entry knows only how to subscribe and how
// to flatten its own payload; the plumbing around that is written once.
//
// The map is built per bot because subscribing needs the agent, and the payload
// of some entries needs the bot to answer a question about the world.
func (b bot) watching() map[string]func(func(map[string]any)) {
	return map[string]func(func(map[string]any)){
		"spawned": raise(b, func(e agent.Spawned) map[string]any {
			return map[string]any{"at": spot(e.At)}
		}),
		"arrived": raise(b, func(e agent.Arrived) map[string]any {
			return map[string]any{
				"at":     spot(e.At.Corner()),
				"reason": reason(e.Reason),
			}
		}),
		"chat": raise(b, func(e agent.ChatReceived) map[string]any {
			return map[string]any{"text": e.Line}
		}),
		"blockChanged": raise(b, func(e agent.BlockChanged) map[string]any {
			return map[string]any{
				"at":    spot(e.At.Corner()),
				"block": b.BlockAt(spot(e.At.Corner())),
			}
		}),
		"health": raise(b, func(e agent.HealthChanged) map[string]any {
			return map[string]any{
				"hp":   e.Health,
				"food": e.Food,
			}
		}),
		"disconnected": raise(b, func(e agent.Disconnected) map[string]any {
			return map[string]any{"reason": e.Reason}
		}),
	}
}

func (b bot) guarding() map[string]func(func(map[string]any) string) {
	return map[string]func(func(map[string]any) string){
		"dig": veto(b, func(e *agent.Digging) map[string]any {
			return map[string]any{
				"block": spot(e.Block.Corner()),
				"tool":  named(e.Tool),
			}
		}),
		"place": veto(b, func(e *agent.Placing) map[string]any {
			return map[string]any{
				"block": spot(e.Block.Corner()),
				"item":  named(e.Item),
			}
		}),
		"chat": veto(b, func(e *agent.Chatting) map[string]any {
			return map[string]any{"text": e.Message}
		}),
		"move": veto(b, func(*agent.Navigating) map[string]any {
			return map[string]any{}
		}),
	}
}

// raise subscribes to one notice and flattens it into what a plugin reads
func raise[N agent.Notice](b bot, flatten func(N) map[string]any) func(func(map[string]any)) {
	return func(handle func(map[string]any)) {
		agent.On(b.agent, func(e N) {
			handle(flatten(e))
		})
	}
}

// veto subscribes to one intent, and a handler that answers with a reason
// refuses it. An empty answer lets it through.
func veto[I agent.Intent](b bot, flatten func(I) map[string]any) func(func(map[string]any) string) {
	return func(handle func(map[string]any) string) {
		agent.Before(b.agent, func(e I) {
			refused := handle(flatten(e))
			if refused == "" {
				return
			}

			e.Refuse(refused)
		})
	}
}

func (b bot) Watch(event string, handle func(map[string]any)) bool {
	subscribe, known := b.watching()[event]
	if !known {
		return false
	}

	subscribe(handle)

	return true
}

func (b bot) Guard(intent string, handle func(map[string]any) string) bool {
	subscribe, known := b.guarding()[intent]
	if !known {
		return false
	}

	subscribe(handle)

	return true
}

func reason(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

var (
	_ plugin.Watcher = bot{}
	_ plugin.Guard   = bot{}
)
