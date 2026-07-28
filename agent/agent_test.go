package agent_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/pathfinder"
)

func TestAgentWalksOnServer(t *testing.T) {
	addr := os.Getenv("GOCRAFT_IT_ADDR")
	if addr == "" {
		t.Skip("set GOCRAFT_IT_ADDR to a running 1.20.4 server to run this integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	bot, err := agent.Join(ctx, addr, "gocraft_walk")
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
	addr := os.Getenv("GOCRAFT_IT_ADDR")
	if addr == "" {
		t.Skip("set GOCRAFT_IT_ADDR to a running 1.20.4 server to run this integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bot, err := agent.Join(ctx, addr, "gocraft_goto")
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
