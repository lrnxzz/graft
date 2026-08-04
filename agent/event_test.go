package agent

import (
	"errors"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func newBus() *Agent {
	return &Agent{
		audience: newAudience(),
	}
}

func TestNoticesReachEveryWatcherOnlyOnDelivery(t *testing.T) {
	bot := newBus()

	var seen []string
	On(bot, func(e ChatReceived) {
		seen = append(seen, "first:"+e.Line)
	})
	On(bot, func(e ChatReceived) {
		seen = append(seen, "second:"+e.Line)
	})

	bot.post(ChatReceived{Line: "oi"})
	if len(seen) != 0 {
		t.Fatalf("a notice reached a watcher before the tick: %v", seen)
	}

	bot.deliver()

	if len(seen) != 2 {
		t.Fatalf("watchers ran %d times, want 2 — a second listener must not replace the first", len(seen))
	}
	if seen[0] != "first:oi" || seen[1] != "second:oi" {
		t.Errorf("delivery = %v, want registration order", seen)
	}
}

func TestNoticesOnlyReachTheirOwnType(t *testing.T) {
	bot := newBus()

	chats := 0
	On(bot, func(ChatReceived) {
		chats++
	})

	bot.post(HealthChanged{Health: 20})
	bot.deliver()

	if chats != 0 {
		t.Errorf("a health notice reached the chat watcher %d times", chats)
	}
}

func TestDeliveryDrainsSoANoticeIsSeenOnce(t *testing.T) {
	bot := newBus()

	seen := 0
	On(bot, func(Spawned) {
		seen++
	})

	bot.post(Spawned{})
	bot.deliver()
	bot.deliver()

	if seen != 1 {
		t.Errorf("the notice was delivered %d times, want 1", seen)
	}
}

func TestAWatcherMayPostFromInsideDelivery(t *testing.T) {
	bot := newBus()

	// the queue is swapped out before handlers run, so a handler that posts is
	// queueing for the next tick rather than deadlocking on the lock it is under
	On(bot, func(Spawned) {
		bot.post(ChatReceived{Line: "spawned"})
	})

	echoed := 0
	On(bot, func(ChatReceived) {
		echoed++
	})

	bot.post(Spawned{})
	bot.deliver()

	if echoed != 0 {
		t.Error("a notice posted during delivery should wait for the next tick")
	}

	bot.deliver()

	if echoed != 1 {
		t.Errorf("the re-posted notice arrived %d times, want 1", echoed)
	}
}

func TestAGuardRefusesAnIntentAndSaysWhy(t *testing.T) {
	bot := newBus()

	Before(bot, func(e *Digging) {
		if e.Block.Y < 0 {
			e.Refuse("below the floor")
		}
	})

	deep := &Digging{Block: gocraft.At(0, -1, 0)}
	err := bot.allowed(deep)
	if err == nil {
		t.Fatal("the guard refused the dig and the intent survived")
	}

	var refusal *Refusal
	if !errors.As(err, &refusal) {
		t.Fatalf("err = %T, want a *Refusal a caller can tell from a failure", err)
	}
	if refusal.Intent != "dig" || refusal.Reason != "below the floor" {
		t.Errorf("refusal = %+v, want the intent and the reason", refusal)
	}

	high := &Digging{Block: gocraft.At(0, 64, 0)}
	if err := bot.allowed(high); err != nil {
		t.Errorf("an unrelated dig was refused: %v", err)
	}
}

func TestEveryGuardRunsEvenAfterOneRefuses(t *testing.T) {
	bot := newBus()

	// a guard that only watches must still see the intent, otherwise a plugin's
	// bookkeeping would depend on the order plugins happened to load in
	ran := 0
	Before(bot, func(e *Placing) {
		ran++
		e.Refuse("first")
	})
	Before(bot, func(*Placing) {
		ran++
	})

	err := bot.allowed(&Placing{})
	if err == nil {
		t.Fatal("the intent was not refused")
	}
	if ran != 2 {
		t.Errorf("guards ran %d times, want 2", ran)
	}
}

func TestAnIntentWithNoGuardsIsAllowed(t *testing.T) {
	bot := newBus()

	if err := bot.allowed(&Chatting{Message: "oi"}); err != nil {
		t.Errorf("an unguarded intent was refused: %v", err)
	}
}
