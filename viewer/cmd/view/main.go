package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"runtime"
	"strings"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/pathfinder"
	"github.com/lrnxzz/go-craft/viewer"
)

const (
	settleDelay = 3 * time.Second
	manualPoll  = 500 * time.Millisecond
)

func init() {
	runtime.LockOSThread()
}

type course struct {
	targets []gocraft.Position
	delay   time.Duration
	repeat  bool
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	screenshot := flag.String("screenshot", "", "render a single frame to this PNG and exit")
	username := flag.String("username", "gocraft_view", "bot username")
	goal := flag.String("goal", "", "pathfind to this x,y,z block while rendering")
	loop := flag.Bool("loop", false, "repeat the goal list forever")
	wait := flag.Duration("wait", 0, "delay before starting to navigate")
	flag.Parse()

	address := flag.Arg(0)
	if address == "" {
		return errors.New("usage: view [flags] <host[:port]>")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, err := agent.Join(ctx, address, *username)
	if err != nil {
		return err
	}

	go func() {
		if err := bot.Run(ctx); err != nil {
			log.Println("run:", err)
		}
	}()

	if err := bot.Ready(ctx); err != nil {
		return err
	}
	time.Sleep(settleDelay)

	view, err := viewer.New(bot, *screenshot == "")
	if err != nil {
		return err
	}

	if *goal != "" {
		targets, err := parseGoals(*goal)
		if err != nil {
			return err
		}

		legs := course{
			targets: targets,
			delay:   *wait,
			repeat:  *loop,
		}
		go walk(ctx, bot, view, legs)
	}

	if *screenshot != "" {
		if err := view.Screenshot(*screenshot); err != nil {
			return err
		}
		log.Printf("wrote %s", *screenshot)

		return nil
	}

	view.Run(ctx)

	return nil
}

func walk(ctx context.Context, bot *agent.Agent, view *viewer.Viewer, legs course) {
	time.Sleep(legs.delay)

	for {
		for _, target := range legs.targets {
			for view.Manual() {
				time.Sleep(manualPoll)
			}

			navigate(ctx, bot, target)
		}

		if !legs.repeat {
			return
		}
	}
}

func navigate(ctx context.Context, bot *agent.Agent, target gocraft.Position) {
	log.Printf("navigating from %v to %v", bot.Snapshot().Position.Floor(), target)

	arrived, err := bot.Navigate(ctx, pathfinder.GoalAt(target))
	if err != nil {
		log.Printf("navigation: %v", err)

		return
	}

	log.Printf("arrived at %v", arrived)
}

func parseGoals(goals string) ([]gocraft.Position, error) {
	var targets []gocraft.Position
	for _, leg := range strings.Split(goals, ";") {
		target, err := gocraft.ParsePosition(leg)
		if err != nil {
			return nil, err
		}

		targets = append(targets, target)
	}

	return targets, nil
}
