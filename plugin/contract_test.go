package plugin_test

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/lrnxzz/graft/plugin"
)

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
