package main

import (
	"context"
	"errors"
	"fmt"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/pathfinder"
	"github.com/lrnxzz/go-craft/viewer"
	"github.com/lrnxzz/go-craft/viewer/command"
)

var (
	target = command.Spot("target")
	ore    = command.Block("ore")
	times  = command.Whole("times").Or(1)
	saying = command.Rest("message")
)

// bench holds what the commands act on. It exists so every handler is a method
// with a name rather than a closure over whatever was in scope.
type bench struct {
	ctx   context.Context
	view  *viewer.Viewer
	route *route
	says  viewer.Speaking
	tree  *command.Tree

	// plugins get whatever the tree does not claim, so a plugin can still add a
	// word without the viewer knowing it exists
	plugins commands
}

func commandsFor(ctx context.Context, view *viewer.Viewer, plugins commands) *bench {
	drawn := &route{
		view: view,
	}

	desk := &bench{
		ctx:   ctx,
		view:  view,
		route: drawn,
		says:  viewer.Says(view.Chat()),

		plugins: plugins,
	}

	desk.tree = command.Rooted(
		command.Word("help").Says("list every command").Runs(desk.help),
		command.Word("goto").Says("walk the bot to a block").Takes(target).Runs(desk.goTo),
		command.Word("mine").Says("mine an ore until the inventory fills").Takes(ore).Runs(desk.mine),
		command.Word("say").Says("speak in chat").Takes(saying).Runs(desk.say),
		command.Word("stop").Says("drop whatever the bot is doing").Runs(desk.stop),

		command.Word("route").Says("draw a path and have the bot follow it").Under(
			command.Word("draw").Says("leave the body to lay out the path").Runs(desk.draw),
			command.Word("list").Says("show the points and whether each can be reached").Runs(desk.list),
			command.Word("walk").Says("follow the path").Takes(times).Runs(desk.walk),
			command.Word("undo").Says("forget the last point").Runs(desk.undo),
			command.Word("clear").Says("forget the whole path").Runs(desk.clear),
		),
	)

	view.AddWorldLayer(drawn)
	view.Settles(drawn.mark)
	view.Intercept(desk.claim)
	view.Offers(desk.offer)

	return desk
}

// claim runs a line the player typed, and reports whether this tree wanted it.
// Anything unclaimed goes on to the server untouched.
func (b *bench) claim(line string) bool {
	held, is := Held(line)
	if !is || held == "" {
		return false
	}

	ran, err := b.tree.Run(held, command.Calling(b.view.Bot(), b.view.Chat().Push))
	if !ran {
		return b.plugins.claim(line)
	}
	if err != nil {
		b.refuse(line, err)
	}

	return true
}

// refuse says what went wrong, points at it, and shows what was wanted, because
// the player has no other way to learn any of the three
func (b *bench) refuse(line string, err error) {
	b.says.Bad(err.Error())

	var spot *command.Refusal
	if errors.As(err, &spot) {
		b.says.Under(line, spot.From+1, spot.To+1)
	}

	var failed *command.Failure
	if errors.As(err, &failed) {
		b.says.Note(prefix + failed.Usage)
	}
}

func (b *bench) offer(line string, cursor int) ([]string, int, int) {
	held, is := Held(line)
	if !is {
		return nil, 0, 0
	}

	offer := b.tree.Complete(held, cursor-1, command.Calling(b.view.Bot(), nil))

	return offer.Words, offer.From + 1, offer.To + 1
}

func (b *bench) help(command.Call) error {
	b.says.Head("commands")

	for _, line := range b.tree.Help() {
		b.says.Note(prefix + line)
	}

	if len(b.plugins.plugins.Usage()) > 0 {
		b.plugins.help()
	}

	return nil
}

func (b *bench) goTo(c command.Call) error {
	at := target.Of(c)
	b.says.Note(fmt.Sprintf("walking to %d %d %d", at.X, at.Y, at.Z))

	go b.walkTo(at)

	return nil
}

func (b *bench) walkTo(at gocraft.Position) {
	if _, err := b.view.Bot().Navigate(b.ctx, pathfinder.GoalAt(at)); err != nil {
		b.says.Bad(err.Error())
	}
}

func (b *bench) mine(c command.Call) error {
	b.says.Note("mining " + ore.Of(c).Name.Path())

	return nil
}

func (b *bench) say(c command.Call) error {
	return b.view.Bot().Chat(saying.Of(c))
}

func (b *bench) stop(command.Call) error {
	b.view.Bot().Stop()
	b.says.Note("stopped")

	return nil
}

func (b *bench) draw(command.Call) error {
	b.view.Detach()
	b.says.Head("drawing a route")
	b.says.Keys("WASD", "fly", "hold still 3s", "mark a block", "F", "back in the body")

	return nil
}

func (b *bench) list(command.Call) error {
	if len(b.route.points) == 0 {
		b.says.Note("no points yet — " + prefix + "route draw")

		return nil
	}

	b.says.Head("route")
	for index, at := range b.route.points {
		b.says.Row(index+1, fmt.Sprintf("%d %d %d", at.X, at.Y, at.Z), spoken(b.route.reach[index]))
	}

	return nil
}

func (b *bench) walk(c command.Call) error {
	if len(b.route.points) == 0 {
		return errNoRoute
	}

	b.route.loops = times.Of(c)
	b.says.Note(fmt.Sprintf("walking %d points, %d times", len(b.route.points), b.route.loops))

	go b.follow()

	return nil
}

func (b *bench) follow() {
	bot := b.view.Bot()

	for loop := range b.route.loops {
		for index, at := range b.route.points {
			if _, err := bot.Navigate(b.ctx, pathfinder.GoalAt(at)); err != nil {
				b.says.Bad(fmt.Sprintf("stopped at point %d: %v", index+1, err))

				return
			}

			b.route.walked = index + 1
			b.says.Bar("route", b.route.walked, len(b.route.points))
		}

		b.says.Good(fmt.Sprintf("lap %d of %d done", loop+1, b.route.loops))
	}
}

func (b *bench) undo(command.Call) error {
	if len(b.route.points) == 0 {
		return errNoRoute
	}

	b.route.points = b.route.points[:len(b.route.points)-1]
	b.route.judge()
	b.says.Note(b.route.describe() + " left")

	return nil
}

func (b *bench) clear(command.Call) error {
	b.route.clear()
	b.says.Note("route forgotten")

	return nil
}
