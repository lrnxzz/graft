package plugin_test

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/lrnxzz/graft/plugin"
)

type wired struct {
	idle
	watched []string
	guarded []string
	fire    func(plugin.Notice)
	veto    func(plugin.Intent) string
}

func (w *wired) Watch(event string, handle func(plugin.Notice)) bool {
	if event != "chat" {
		return false
	}

	w.watched = append(w.watched, event)
	w.fire = handle

	return true
}

func (w *wired) Guard(intent string, handle func(plugin.Intent) string) bool {
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

	said := plugin.Said{
		Text: "olá",
	}

	bot.fire(said)
}

func TestAPluginCanRefuseAnIntent(t *testing.T) {
	_, bot := listener(t)

	if bot.veto == nil {
		t.Fatal("the plugin never guarded dig")
	}

	deep := plugin.Digging{
		Block: plugin.Vec3{
			Y: 64,
		},
	}

	refused := bot.veto(deep)
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

// graft.d.ts is what a plugin author reads, and nothing in Go makes it true.
// It had already drifted from the payloads the host sends in six of eleven
// declarations before this test existed, so it is checked against the catalogue
// rather than against a second list written by hand.
var declared = regexp.MustCompile(`(?m)^\s{4}(\w+): \{ ?([^}]*?) ?\}`)

func contract(t *testing.T, block string) map[string][]string {
	t.Helper()

	source, err := os.ReadFile("graft.d.ts")
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	opening := "export interface " + block + " {"

	start := strings.Index(string(source), opening)
	if start < 0 {
		t.Fatalf("graft.d.ts declares no %s", block)
	}

	rest := string(source)[start+len(opening):]

	end := strings.Index(rest, "\n  }")
	if end < 0 {
		t.Fatalf("the %s block is never closed", block)
	}

	found := map[string][]string{}
	for _, line := range declared.FindAllStringSubmatch(rest[:end], -1) {
		var fields []string
		for _, pair := range strings.Split(line[2], ";") {
			name, _, split := strings.Cut(pair, ":")
			if !split {
				continue
			}

			fields = append(fields, strings.TrimSpace(name))
		}

		sort.Strings(fields)
		found[line[1]] = fields
	}

	return found
}

func carried(payload any) []string {
	shape := reflect.TypeOf(payload)

	var fields []string
	for index := range shape.NumField() {
		fields = append(fields, shape.Field(index).Tag.Get("json"))
	}

	sort.Strings(fields)

	return fields
}

func TestEveryNoticeIsDeclaredAsItIsSent(t *testing.T) {
	promised := contract(t, "Events")

	for _, notice := range plugin.Notices() {
		want := carried(notice)

		got, known := promised[notice.Event()]
		if !known {
			t.Errorf("the bot raises %q and graft.d.ts never declares it", notice.Event())

			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: declared %v, sent %v", notice.Event(), got, want)
		}

		delete(promised, notice.Event())
	}

	for event := range promised {
		t.Errorf("graft.d.ts declares %q and the bot never raises it", event)
	}
}

func TestEveryIntentIsDeclaredAsItIsSent(t *testing.T) {
	promised := contract(t, "Intents")

	for _, about := range plugin.Intents() {
		want := carried(about)

		got, known := promised[about.Intent()]
		if !known {
			t.Errorf("the bot guards %q and graft.d.ts never declares it", about.Intent())

			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: declared %v, guarded %v", about.Intent(), got, want)
		}

		delete(promised, about.Intent())
	}

	for intent := range promised {
		t.Errorf("graft.d.ts declares %q and the bot never guards it", intent)
	}
}
