package main

import (
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/plugin"
)

// built per bot because subscribing needs the agent, and some payloads need the
// bot to answer a question about the world
func (b bot) watching() map[string]func(func(plugin.Notice)) {
	return map[string]func(func(plugin.Notice)){
		plugin.Spawned{}.Event():       raise(b, spawnedAt),
		plugin.Arrived{}.Event():       raise(b, arrivedAt),
		plugin.Said{}.Event():          raise(b, chatSaid),
		plugin.BlockChanged{}.Event():  raise(b, b.blockChanged),
		plugin.HealthChanged{}.Event(): raise(b, healthNow),
		plugin.Disconnected{}.Event():  raise(b, disconnectedWhy),
	}
}

func (b bot) guarding() map[string]func(func(plugin.Intent) string) {
	return map[string]func(func(plugin.Intent) string){
		plugin.Digging{}.Intent():    veto(b, digging),
		plugin.Placing{}.Intent():    veto(b, placing),
		plugin.Chatting{}.Intent():   veto(b, chatting),
		plugin.Navigating{}.Intent(): veto(b, navigating),
	}
}

func spawnedAt(e agent.Spawned) plugin.Notice {
	return plugin.Spawned{
		At: spot(e.At),
	}
}

func arrivedAt(e agent.Arrived) plugin.Notice {
	return plugin.Arrived{
		At:     spot(e.At.Corner()),
		Reason: reason(e.Reason),
	}
}

func chatSaid(e agent.ChatReceived) plugin.Notice {
	return plugin.Said{
		Text: e.Line,
	}
}

func (b bot) blockChanged(e agent.BlockChanged) plugin.Notice {
	at := spot(e.At.Corner())

	return plugin.BlockChanged{
		At:    at,
		Block: b.BlockAt(at),
	}
}

func healthNow(e agent.HealthChanged) plugin.Notice {
	return plugin.HealthChanged{
		Health: e.Health,
		Food:   e.Food,
	}
}

func disconnectedWhy(e agent.Disconnected) plugin.Notice {
	return plugin.Disconnected{
		Reason: e.Reason,
	}
}

func digging(e *agent.Digging) plugin.Intent {
	return plugin.Digging{
		Block: spot(e.Block.Corner()),
		Tool:  named(e.Tool),
	}
}

func placing(e *agent.Placing) plugin.Intent {
	return plugin.Placing{
		Block: spot(e.Block.Corner()),
		Item:  named(e.Item),
	}
}

func chatting(e *agent.Chatting) plugin.Intent {
	return plugin.Chatting{
		Text: e.Message,
	}
}

func navigating(*agent.Navigating) plugin.Intent {
	return plugin.Navigating{}
}

func raise[N agent.Notice](b bot, flatten func(N) plugin.Notice) func(func(plugin.Notice)) {
	return func(handle func(plugin.Notice)) {
		notice := func(e N) {
			handle(flatten(e))
		}

		agent.On(b.agent, notice)
	}
}

// an empty answer lets the intent through; anything else refuses it
func veto[I agent.Intent](b bot, flatten func(I) plugin.Intent) func(func(plugin.Intent) string) {
	return func(handle func(plugin.Intent) string) {
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

func (b bot) Watch(event string, handle func(plugin.Notice)) bool {
	subscribe, known := b.watching()[event]
	if !known {
		return false
	}

	subscribe(handle)

	return true
}

func (b bot) Guard(intent string, handle func(plugin.Intent) string) bool {
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
