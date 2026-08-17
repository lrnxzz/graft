package plugin_test

import (
	"path/filepath"
	"testing"

	"github.com/lrnxzz/graft/plugin"
)

type paper struct {
	width  float32
	height float32
	drawn  int
}

func (p *paper) Size() (float32, float32) {
	return p.width, p.height
}

func (p *paper) Fill(float32, float32, float32, float32, plugin.Color) {
	p.drawn++
}

func (p *paper) Text(string, float32, float32, float32, plugin.Color) {
	p.drawn++
}

func (p *paper) Icon(plugin.Sprite, float32, float32, float32) {
	p.drawn++
}

func (p *paper) Measure(text string, scale float32) float32 {
	return float32(len(text)) * 6 * scale
}

type idle struct{}

func (idle) Name() string                             { return "tester" }
func (idle) Position() plugin.Vec3                    { return plugin.Vec3{} }
func (idle) Health() float32                          { return 20 }
func (idle) Food() float32                            { return 20 }
func (idle) OnGround() bool                           { return true }
func (idle) Held() plugin.Item                        { return "air" }
func (idle) Looking() (plugin.Target, bool)           { return plugin.Target{}, false }
func (idle) Inventory() []plugin.Stack                { return nil }
func (idle) BlockAt(plugin.Vec3) plugin.Block         { return "air" }
func (idle) Count(plugin.Item) int                    { return 0 }
func (idle) Pursue(plugin.Goal) error                 { return nil }
func (idle) Abandon()                                 {}
func (idle) Goto(at plugin.Vec3) (plugin.Vec3, error) { return at, nil }
func (idle) Dig(plugin.Vec3) error                    { return nil }
func (idle) Place(plugin.Vec3) error                  { return nil }
func (idle) Hold(plugin.Item) error                   { return nil }
func (idle) Say(string) error                         { return nil }
func (idle) Look(plugin.Vec3)                         {}

func loadExample(t *testing.T) *plugin.Runtime {
	t.Helper()

	source, err := plugin.Compile(filepath.Join("example", "auto-miner.tsx"))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	runtime, err := plugin.Load(source, idle{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	return runtime
}

func TestExamplePluginDeclaresItself(t *testing.T) {
	declared := loadExample(t).Declaration()

	if declared.Name != "auto-miner" {
		t.Errorf("name = %q, want auto-miner", declared.Name)
	}
	if !declared.Allows(plugin.MayDig) {
		t.Error("the plugin asked for dig and did not get it")
	}
	if declared.Allows("fly") {
		t.Error("a permission that was never asked for was granted")
	}
	if len(declared.Reactions) != 3 {
		t.Errorf("reactions = %d, want 3", len(declared.Reactions))
	}
	if len(declared.Keys) != 2 {
		t.Errorf("keys = %d, want 2", len(declared.Keys))
	}
}

func TestExamplePluginParsesItsCommands(t *testing.T) {
	declared := loadExample(t).Declaration()

	mine, ok := declared.Commands["mine"]
	if !ok {
		t.Fatal("the mine command was not declared")
	}

	if usage := mine.Usage("mine"); usage != "mine <ore> [radius]" {
		t.Errorf("usage = %q, want the optional radius in brackets", usage)
	}
}

func TestCommandRefusesAMissingRequiredArgument(t *testing.T) {
	runtime := loadExample(t)

	fetch := runtime.Declaration().Commands["fetch"]
	if fetch == nil {
		t.Fatal("the fetch command was not declared")
	}

	_, err := runtime.Parse(fetch, nil)
	if err == nil {
		t.Error("a command with no arguments should not have parsed")
	}
}

func TestHudLaysOutAndPaints(t *testing.T) {
	runtime := loadExample(t)
	surface := &paper{width: 800, height: 600}

	tree, ok, err := runtime.Hud()
	if err != nil {
		t.Fatalf("hud: %v", err)
	}
	if !ok {
		t.Fatal("the plugin declares a hud and none was produced")
	}

	plugin.Paint(tree, surface)
	if surface.drawn == 0 {
		t.Error("the hud painted nothing")
	}
}

func TestMenuScrollsAndResolvesAClickToItsOption(t *testing.T) {
	runtime := loadExample(t)
	surface := &paper{width: 800, height: 600}

	menu, opened, err := runtime.Press("M", nil)
	if err != nil {
		t.Fatalf("press: %v", err)
	}
	if !opened {
		t.Fatal("pressing M opened no menu")
	}

	tree, built, err := menu.Body()
	if err != nil || !built {
		t.Fatalf("menu body: built=%v err=%v", built, err)
	}

	picks := plugin.Paint(tree, surface)
	if len(picks) == 0 {
		t.Fatal("the ore list produced nothing clickable")
	}
}

type watcher struct{}

func (watcher) Name() string          { return "watcher" }
func (watcher) Position() plugin.Vec3 { return plugin.Vec3{} }

func surfaceOf(t *testing.T, bot plugin.Bot) map[string]bool {
	t.Helper()

	source, err := plugin.Compile(filepath.Join("example", "auto-miner.tsx"))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	runtime, err := plugin.Load(source, bot)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	return runtime.Exposed()
}

func TestABotWithoutTheInterfaceSimplyHasNoAbility(t *testing.T) {
	full := surfaceOf(t, idle{})
	if !full["dig"] {
		t.Error("a bot that can dig was not given dig")
	}
	if !full["health"] {
		t.Error("a bot with vitals was not given health")
	}

	bare := surfaceOf(t, watcher{})
	if bare["dig"] {
		t.Error("a bot that cannot dig was handed dig anyway")
	}
	if bare["health"] {
		t.Error("a bot with no vitals was handed health anyway")
	}
	if !bare["name"] {
		t.Error("every bot has a name and this one lost it")
	}
}

func TestAnUndeclaredPermissionIsNeverBound(t *testing.T) {
	exposed := surfaceOf(t, idle{})

	if exposed["place"] {
		t.Error("place was bound although the plugin never asked for it")
	}
	if !exposed["say"] {
		t.Error("chat was asked for and say was not bound")
	}
}

// Every name a plugin can write comes from one of three catalogues onto one
// javascript object. Nothing stops two of them choosing the same word, and the
// loser would simply never be reachable, so the clash has to be an error.
func TestTheCataloguesDoNotOverlap(t *testing.T) {
	claimed := map[string]string{}

	for _, spec := range plugin.Components() {
		claimed[spec.Tag] = "component"
	}
	for _, spec := range plugin.Goals() {
		where, taken := claimed[string(spec.Type)]
		if taken {
			t.Errorf("goal %s is already a %s", spec.Type, where)
		}

		claimed[string(spec.Type)] = "goal"
	}
	for _, spec := range plugin.Markers() {
		where, taken := claimed[string(spec.Type)]
		if taken {
			t.Errorf("marker %s is already a %s", spec.Type, where)
		}

		claimed[string(spec.Type)] = "marker"
	}
}

func TestColorReadsCssNotation(t *testing.T) {
	opaque := plugin.ParseColor("#0f8")
	if opaque.Alpha != 1 || opaque.Green != 1 {
		t.Errorf("#0f8 = %+v, want full green and opaque", opaque)
	}

	faded := plugin.ParseColor("#000a")
	if faded.Alpha == 1 || faded.Alpha == 0 {
		t.Errorf("#000a alpha = %v, want a partial alpha", faded.Alpha)
	}

	if !plugin.ParseColor("nonsense").Transparent() {
		t.Error("an unreadable colour should read as unset")
	}
}
