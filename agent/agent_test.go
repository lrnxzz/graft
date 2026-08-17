package agent_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/codec/v765/items"
	"github.com/lrnxzz/go-craft/pathfinder"
	"github.com/lrnxzz/go-craft/rcon"
)

// joined connects a bot and leaves its session running, which is the same
// opening in front of every test here. It returns the channel the session ends
// on as well, because a test that waits on the bot has to be able to lose the
// race to a connection that died instead.
func joined(ctx context.Context, t *testing.T, username string) (*agent.Agent, <-chan error) {
	t.Helper()

	addr := os.Getenv("GOCRAFT_IT_ADDR")
	if addr == "" {
		t.Skip("set GOCRAFT_IT_ADDR to a running 1.20.4 server to run this integration test")
	}

	bot, err := agent.Join(ctx, addr, username)
	if err != nil {
		t.Fatalf("join: %v", err)
	}

	finished := make(chan error, 1)
	go func() {
		finished <- bot.Run(ctx)
	}()

	select {
	case <-bot.Spawned():
	case err := <-finished:
		t.Fatalf("bot never reached play / received its spawn position: %v", err)
	}

	return bot, finished
}

func TestAgentWalksOnServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	bot, finished := joined(ctx, t, "gocraft_walk")

	start := bot.Player().Position
	bot.Look(0, 0)
	bot.SetControls(gocraft.Controls{Forward: true})

	if err := <-finished; err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("run: %v", err)
	}

	end := bot.Player().Position
	distance := end.Distance(start)

	t.Logf("walked from %v to %v (%.1f blocks), on_ground=%v, chunks=%d",
		start, end, distance, bot.Player().OnGround, bot.World().Loaded())

	if bot.World().Loaded() == 0 {
		t.Error("no chunks were ingested")
	}
	if distance < 5 {
		t.Errorf("bot barely moved (%.1f blocks); walking did not take effect", distance)
	}
	if !bot.Player().OnGround {
		t.Error("bot should stay on the ground while walking on flat terrain")
	}
}

func TestAgentMovesToTarget(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bot, finished := joined(ctx, t, "gocraft_goto")

	target := bot.Player().Position.Offset(6, 0, 10)

	walked := make(chan error, 1)
	go func() {
		_, err := bot.Navigate(ctx, pathfinder.GoalAt(target.Floor()))
		walked <- err
	}()

	select {
	case err := <-walked:
		if err != nil {
			t.Fatalf("navigate: %v", err)
		}
	case err := <-finished:
		t.Fatalf("run: %v", err)
	}

	end := bot.Player().Position
	gap := end.Horizontal().Distance(target.Horizontal())

	t.Logf("target %v, arrived at %v (gap %.2f blocks)", target, end, gap)

	if gap > 1.5 {
		t.Errorf("bot did not reach the target: gap %.2f blocks", gap)
	}
}

const (
	digReach = 4.5
	digPitch = 75
)

// breaking a block is only half client-side: the swing has to be accepted and
// the server has to answer with the block change. A client that mines happily
// against a server that refused it looks identical from inside the bot, so this
// reads the state back out of the world rather than trusting Dig returning.
func TestDiggingBreaksABlockOnServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bot, _ := joined(ctx, t, "gocraft_dig")

	if err := bot.Ready(ctx); err != nil {
		t.Fatalf("terrain: %v", err)
	}

	// looking down puts the ground in the crosshair on any world, so nothing has
	// to be placed in front of the bot first
	bot.Look(0, digPitch)
	time.Sleep(300 * time.Millisecond)

	hit, sighted := bot.Target(digReach)
	if !sighted {
		t.Fatal("no block in the crosshair after looking down")
	}

	before, _ := bot.World().BlockAt(hit.Block)

	dug, err := bot.Dig(ctx, digReach)
	if err != nil {
		t.Fatalf("dig: %v", err)
	}

	// the block change comes back on its own packet, a round trip after the
	// swing the server accepted
	time.Sleep(time.Second)

	after, _ := bot.World().BlockAt(dug.Block)

	t.Logf("broke %v: state before=%d after=%d", dug.Block, before, after)

	if after != 0 {
		t.Errorf("%v is still state %d; the server refused the dig", dug.Block, after)
	}
}

const (
	clickSettle  = 2 * time.Second
	clickStocked = 10 * time.Second
	clickPoll    = 200 * time.Millisecond
)

// moving a stack is only half client-side: the server validates the click and
// answers one it refused by sending the slots back the way they were, leaving a
// bot convinced it moved an item it never moved. This reads the slots back out
// after the server has had a round trip to disagree.
func TestClickingMovesAStackOnServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	bot, _ := joined(ctx, t, "gocraft_click")

	if err := bot.Ready(ctx); err != nil {
		t.Fatalf("terrain: %v", err)
	}

	prepared(t, fmt.Sprintf("give %s bread 1", bot.Player().Username))

	source, stocked := -1, false
	for deadline := time.Now().Add(clickStocked); time.Now().Before(deadline); {
		if source, stocked = bot.Inventory().FindItem(items.Bread); stocked {
			break
		}

		time.Sleep(clickPoll)
	}
	if !stocked {
		t.Fatal("the bread never arrived in the inventory")
	}

	target, room := bot.Inventory().FirstEmpty()
	if !room {
		t.Fatal("no empty slot to move the bread into")
	}

	if err := bot.ClickSlot(source); err != nil {
		t.Fatalf("pick up: %v", err)
	}
	if bot.Carried().Empty() {
		t.Fatalf("clicking slot %d picked nothing up", source)
	}

	if err := bot.ClickSlot(target); err != nil {
		t.Fatalf("put down: %v", err)
	}

	time.Sleep(clickSettle)

	t.Logf("source=%v target=%v carried=%v",
		bot.Inventory().Slot(source), bot.Inventory().Slot(target), bot.Carried())

	if !bot.Inventory().Slot(target).Is(items.Bread) {
		t.Errorf("slot %d does not hold the bread; the server reverted the move", target)
	}
	if !bot.Carried().Empty() {
		t.Error("the cursor still holds something after putting the stack down")
	}
}

// prepared runs one console command on the test server, which is how a test
// that needs something in the world puts it there. Without rcon it cannot set
// itself up, so it skips rather than failing on an empty inventory.
func prepared(t *testing.T, command string) {
	t.Helper()

	addr := os.Getenv("GOCRAFT_IT_RCON")
	if addr == "" {
		t.Skip("set GOCRAFT_IT_RCON to the server's rcon address to run this integration test")
	}

	password := os.Getenv("GOCRAFT_IT_RCON_PASSWORD")
	if password == "" {
		password = "gocraft"
	}

	console, err := rcon.Dial(addr, password)
	if err != nil {
		t.Fatalf("rcon: %v", err)
	}
	defer func() {
		_ = console.Close()
	}()

	answer, err := console.Run(command)
	if err != nil {
		t.Fatalf("rcon %q: %v", command, err)
	}

	t.Logf("rcon %q answered %q", command, answer)
}

// a line the bot sends comes back to it like anyone else's, which is the only
// proof from inside the client that the server took it
func TestChatSentComesBack(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bot, _ := joined(ctx, t, "gocraft_chat")

	heard := make(chan string, 8)
	agent.On(bot, func(e agent.ChatReceived) {
		// a notice handler runs on the tick, so a send that blocks would stall
		// the bot rather than fail the test
		select {
		case heard <- e.Line:
		default:
		}
	})

	said := fmt.Sprintf("gocraft integration %d", time.Now().UnixNano())
	if err := bot.Chat(said); err != nil {
		t.Fatalf("chat: %v", err)
	}

	waiting, stop := context.WithTimeout(ctx, 10*time.Second)
	defer stop()

	for {
		select {
		case line := <-heard:
			t.Logf("heard %q", line)
			if strings.Contains(line, said) {
				return
			}
		case <-waiting.Done():
			t.Fatalf("the server never echoed %q back", said)
		}
	}
}
