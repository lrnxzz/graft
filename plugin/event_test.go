package plugin_test

import (
	"path/filepath"
	"testing"

	"github.com/lrnxzz/go-craft/plugin"
)

type wired struct {
	idle
	watched []string
	guarded []string
	fire    func(map[string]any)
	veto    func(map[string]any) string
}

func (w *wired) Watch(event string, handle func(map[string]any)) bool {
	if event != "chat" {
		return false
	}

	w.watched = append(w.watched, event)
	w.fire = handle

	return true
}

func (w *wired) Guard(intent string, handle func(map[string]any) string) bool {
	if intent != "dig" {
		return false
	}

	w.guarded = append(w.guarded, intent)
	w.veto = handle

	return true
}

func listener(t *testing.T) (*plugin.Runtime, *wired) {
	t.Helper()

	source, err := plugin.Compile(filepath.Join("testdata", "events", "listener.ts"))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	bot := &wired{}

	runtime, err := plugin.Load(source, bot)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := runtime.Setup(nil); err != nil {
		t.Fatalf("setup: %v", err)
	}

	return runtime, bot
}

func TestAPluginSubscribesThroughTheHost(t *testing.T) {
	_, bot := listener(t)

	if len(bot.watched) != 1 || bot.watched[0] != "chat" {
		t.Errorf("watched = %v, want just the event the host knows", bot.watched)
	}
	if len(bot.guarded) != 1 || bot.guarded[0] != "dig" {
		t.Errorf("guarded = %v, want dig", bot.guarded)
	}
}

func TestANoticeReachesThePluginHandler(t *testing.T) {
	_, bot := listener(t)

	if bot.fire == nil {
		t.Fatal("the plugin never subscribed to chat")
	}

	bot.fire(map[string]any{"text": "olá"})
}

func TestAPluginCanRefuseAnIntent(t *testing.T) {
	_, bot := listener(t)

	if bot.veto == nil {
		t.Fatal("the plugin never guarded dig")
	}

	refused := bot.veto(map[string]any{"block": plugin.Vec3{Y: 64}})
	if refused != "nothing is mined here" {
		t.Errorf("refusal = %q, want the reason the plugin gave", refused)
	}
}

func TestAnUnguardedIntentIsNotRefused(t *testing.T) {
	runtime, _ := listener(t)

	exposed := runtime.Exposed()
	if !exposed["on"] || !exposed["before"] {
		t.Error("a bot that watches and guards was not given on and before")
	}
}
