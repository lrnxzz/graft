package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"runtime"
	"strings"
	"time"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/agent"
	"github.com/lrnxzz/graft/host"
	"github.com/lrnxzz/graft/pathfinder"
	"github.com/lrnxzz/graft/viewer"
)

const manualPoll = 500 * time.Millisecond

func init() {
	runtime.LockOSThread()
}

type course struct {
	targets []graft.Position
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
	username := flag.String("username", "graft_view", "bot username")
	goal := flag.String("goal", "", "pathfind to this x,y,z block while rendering")
	loop := flag.Bool("loop", false, "repeat the goal list forever")
	wait := flag.Duration("wait", 0, "delay before starting to navigate")
	flag.Parse()

	address := flag.Arg(0)
	if address == "" {
		return errors.New("usage: view [flags] <host[:port]>")
	}

	targets, err := parseGoals(*goal)
	if err != nil {
		return err
	}

	// the plugin is what a user of the project writes; the host owns everything
	// around it, and the window opens here because this is still main's goroutine
	render := func(ctx context.Context, bot *agent.Agent) error {
		view, err := viewer.New(bot, *screenshot == "")
		if err != nil {
			return err
		}

		if len(targets) > 0 {
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

	return host.Run(context.Background(), address, *username, render)
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

func navigate(ctx context.Context, bot *agent.Agent, target graft.Position) {
	log.Printf("navigating from %v to %v", bot.Snapshot().Position.Floor(), target)

	arrived, err := bot.Navigate(ctx, pathfinder.GoalAt(target))
	if err != nil {
		log.Printf("navigation: %v", err)

		return
	}

	log.Printf("arrived at %v", arrived)
}

func parseGoals(goals string) ([]graft.Position, error) {
	if goals == "" {
		return nil, nil
	}

	var targets []graft.Position
	for _, leg := range strings.Split(goals, ";") {
		target, err := graft.ParsePosition(leg)
		if err != nil {
			return nil, err
		}

		targets = append(targets, target)
	}

	return targets, nil
}
