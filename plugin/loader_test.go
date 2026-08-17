package plugin_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/lrnxzz/graft/plugin"
)

func loadFolder(t *testing.T) (*plugin.Plugins, []plugin.Failure) {
	t.Helper()

	return plugin.LoadAll(filepath.Join("testdata", "plugins"), idle{})
}

func TestABrokenPluginDoesNotTakeTheOthersDown(t *testing.T) {
	plugins, refused := loadFolder(t)

	names := plugins.Names()
	if !slices.Contains(names, "alpha") || !slices.Contains(names, "beta") {
		t.Errorf("loaded = %v, want the two sound plugins", names)
	}
	if slices.Contains(names, "broken") {
		t.Error("a plugin with no default export was loaded anyway")
	}

	if !reported(refused, "broken") {
		t.Errorf("refusals = %v, want the broken plugin named", refused)
	}
}

func TestTwoPluginsCannotClaimTheSameCommand(t *testing.T) {
	plugins, refused := loadFolder(t)

	if !reported(refused, "already belongs to alpha") {
		t.Errorf("refusals = %v, want the clash reported against the first claimant", refused)
	}

	usage := plugins.Usage()
	if _, offered := usage["there"]; !offered {
		t.Error("beta lost a command it did not clash on")
	}
	if usage["mine"] != "mine <ore>" {
		t.Errorf("mine = %q, want alpha's, which claimed it first", usage["mine"])
	}
}

func TestACommandRunsOnThePluginThatClaimedIt(t *testing.T) {
	plugins, _ := loadFolder(t)

	found, err := plugins.Run("here", nil)
	if !found || err != nil {
		t.Errorf("run here: found=%v err=%v", found, err)
	}

	found, _ = plugins.Run("nowhere", nil)
	if found {
		t.Error("an unclaimed word was reported as a command")
	}
}

func TestAKeyBelongsToOnePluginOnly(t *testing.T) {
	plugins, refused := loadFolder(t)

	if !reported(refused, "the key \"M\"") {
		t.Errorf("refusals = %v, want the key clash reported", refused)
	}

	_, opened, err := plugins.Press("M", nil)
	if err != nil {
		t.Fatalf("press: %v", err)
	}
	if opened {
		t.Error("alpha binds M to abandon and opened a menu")
	}

	_, opened, _ = plugins.Press("Z", nil)
	if opened {
		t.Error("an unbound key opened something")
	}
}

func TestEveryLoadedPluginContributesItsHud(t *testing.T) {
	plugins, _ := plugin.LoadAll("example", idle{})

	roots, failures := plugins.Hud()
	if len(failures) != 0 {
		t.Fatalf("hud failures: %v", failures)
	}
	if len(roots) != 1 {
		t.Fatalf("huds = %d, want the one the example declares", len(roots))
	}

	surface := &paper{width: 800, height: 600}
	plugin.Paint(roots[0], surface)

	if surface.drawn == 0 {
		t.Error("the loaded plugin painted nothing")
	}
}

func TestAMissingFolderIsNotAFailure(t *testing.T) {
	plugins, refused := plugin.LoadAll(filepath.Join("testdata", "absent"), idle{})

	if len(refused) != 0 {
		t.Errorf("refusals = %v, want none for a bot with no plugins", refused)
	}
	if len(plugins.Names()) != 0 {
		t.Errorf("loaded = %v, want nothing", plugins.Names())
	}
}

func reported(failures []plugin.Failure, needle string) bool {
	for _, failure := range failures {
		if strings.Contains(failure.Error(), needle) {
			return true
		}
	}

	return false
}
