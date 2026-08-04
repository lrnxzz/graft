package main

import (
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/plugin"
)

// built per bot because subscribing needs the agent, and some payloads need the
// bot to answer a question about the world
func (b bot) watching() map[string]func(func(map[string]any)) {
	return map[string]func(func(map[string]any)){
		"spawned":      raise(b, spawnedAt),
		"arrived":      raise(b, arrivedAt),
		"chat":         raise(b, chatSaid),
		"blockChanged": raise(b, b.blockChanged),
		"health":       raise(b, healthNow),
		"disconnected": raise(b, disconnectedWhy),
	}
}

func (b bot) guarding() map[string]func(func(map[string]any) string) {
	return map[string]func(func(map[string]any) string){
		"dig":   veto(b, digging),
		"place": veto(b, placing),
		"chat":  veto(b, chatting),
		"move":  veto(b, navigating),
	}
}

func spawnedAt(e agent.Spawned) map[string]any {
	return map[string]any{
		"at": spot(e.At),
	}
}

func arrivedAt(e agent.Arrived) map[string]any {
	return map[string]any{
		"at":     spot(e.At.Corner()),
		"reason": reason(e.Reason),
	}
}

func chatSaid(e agent.ChatReceived) map[string]any {
	return map[string]any{
		"text": e.Line,
	}
}

func (b bot) blockChanged(e agent.BlockChanged) map[string]any {
	at := spot(e.At.Corner())

	return map[string]any{
		"at":    at,
		"block": b.BlockAt(at),
	}
}

func healthNow(e agent.HealthChanged) map[string]any {
	return map[string]any{
		"hp":   e.Health,
		"food": e.Food,
	}
}

func disconnectedWhy(e agent.Disconnected) map[string]any {
	return map[string]any{
		"reason": e.Reason,
	}
}

func digging(e *agent.Digging) map[string]any {
	return map[string]any{
		"block": spot(e.Block.Corner()),
		"tool":  named(e.Tool),
	}
}

func placing(e *agent.Placing) map[string]any {
	return map[string]any{
		"block": spot(e.Block.Corner()),
		"item":  named(e.Item),
	}
}

func chatting(e *agent.Chatting) map[string]any {
	return map[string]any{
		"text": e.Message,
	}
}

func navigating(*agent.Navigating) map[string]any {
	return map[string]any{}
}

func raise[N agent.Notice](b bot, flatten func(N) map[string]any) func(func(map[string]any)) {
	return func(handle func(map[string]any)) {
		notice := func(e N) {
			handle(flatten(e))
		}

		agent.On(b.agent, notice)
	}
}

// an empty answer lets the intent through; anything else refuses it
func veto[I agent.Intent](b bot, flatten func(I) map[string]any) func(func(map[string]any) string) {
	return func(handle func(map[string]any) string) {
		intent := func(e I) {
			refused := handle(flatten(e))
			if refused == "" {
				return
			}

			e.Refuse(refused)
		}

		agent.Before(b.agent, intent)
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
