package plugin_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/lrnxzz/go-craft/plugin"
)

// talker records what the page's handler asked the bot to do, which is the only
// way to tell a message that arrived from one that was merely accepted
type talker struct {
	said []string
}

func (t *talker) Name() string {
	return "talker"
}

func (t *talker) Position() plugin.Vec3 {
	return plugin.Vec3{}
}

func (t *talker) Say(line string) error {
	t.said = append(t.said, line)

	return nil
}

func openedPage(t *testing.T, bot plugin.Bot) *plugin.Menu {
	t.Helper()

	source, err := plugin.Compile(filepath.Join("testdata", "page", "panel.tsx"))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	runtime, err := plugin.Load(source, bot)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	menu, opened, err := runtime.Press("M", nil)
	if err != nil || !opened {
		t.Fatalf("press: opened=%v err=%v", opened, err)
	}

	return menu
}

func TestAMenuCanBeWrittenAsAPage(t *testing.T) {
	menu := openedPage(t, &talker{})

	written, is, err := menu.Page()
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if !is {
		t.Fatal("a menu that declared a page reported none")
	}

	if !strings.Contains(written.Style, ".ore:hover") {
		t.Error("the plugin's stylesheet was not carried")
	}
	if !strings.Contains(written.Body, "gocraft.send('pick', 'iron_ore')") {
		t.Error("the plugin's markup was not carried")
	}
	if !strings.Contains(written.Body, "<p>talker</p>") {
		t.Error("the page was built without the bot it is shown for")
	}
	if strings.Contains(written.Body, "<!doctype") {
		t.Error("the plugin assembled a document, which is the host's job")
	}
}

// the two ways of describing a menu stay apart: a page declares no node tree,
// and a node tree declares no page
func TestAPageDeclaresNoNodeTree(t *testing.T) {
	menu := openedPage(t, &talker{})

	_, built, err := menu.Body()
	if err != nil {
		t.Fatalf("body: %v", err)
	}
	if built {
		t.Fatal("a menu written as a page also produced a node tree")
	}
}

func TestTheNodeTreeStillDeclaresNoPage(t *testing.T) {
	runtime := loadExample(t)

	menu, opened, err := runtime.Press("M", nil)
	if err != nil || !opened {
		t.Fatalf("press: opened=%v err=%v", opened, err)
	}

	_, is, err := menu.Page()
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if is {
		t.Fatal("the component menu also produced a page")
	}
}

// the host picks a renderer from this alone, before either callback has run, so
// it has to agree with the two above without building anything
func TestAMenuSaysWhichRendererItWantsWithoutBuilding(t *testing.T) {
	page := openedPage(t, &talker{})
	if !page.Written() {
		t.Error("a menu declaring a page did not ask for the html renderer")
	}

	runtime := loadExample(t)

	tree, opened, err := runtime.Press("M", nil)
	if err != nil || !opened {
		t.Fatalf("press: opened=%v err=%v", opened, err)
	}
	if tree.Written() {
		t.Error("a menu declaring a node tree asked for the html renderer")
	}
}

// what the page sends has to run the plugin's own code, on the plugin's own bot
func TestWhatThePageSendsReachesTheBot(t *testing.T) {
	bot := &talker{}
	menu := openedPage(t, bot)

	if err := menu.Send("pick", []string{"iron_ore"}); err != nil {
		t.Fatalf("send: %v", err)
	}

	if len(bot.said) != 1 || bot.said[0] != "mining iron_ore" {
		t.Fatalf("the bot heard %q, want [mining iron_ore]", bot.said)
	}
}

// markup outlives the handler that answered it, so an unclaimed name is quiet
// rather than a failure
func TestAnUnclaimedMessageIsIgnored(t *testing.T) {
	bot := &talker{}
	menu := openedPage(t, bot)

	if err := menu.Send("vanished", []string{"x"}); err != nil {
		t.Fatalf("an unclaimed message failed instead of being ignored: %v", err)
	}
	if len(bot.said) != 0 {
		t.Fatalf("an unclaimed message still reached the bot: %q", bot.said)
	}
}

// A page is rebuilt from the plugin's own state, so what the page sends has to
// come back as different markup. Without that the panel would look frozen no
// matter what it was told.
func TestWhatThePageSendsChangesWhatItShows(t *testing.T) {
	source, err := plugin.Compile(filepath.Join("example", "ore-panel.tsx"))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	runtime, err := plugin.Load(source, &digger{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	menu, opened, err := runtime.Press("B", nil)
	if err != nil || !opened {
		t.Fatalf("press: opened=%v err=%v", opened, err)
	}

	ores, _, err := menu.Page()
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if !strings.Contains(ores.Body, "icon-diamond_ore") {
		t.Fatal("the ore tab listed no ores")
	}

	if err := menu.Send("tab", []string{"ajustes"}); err != nil {
		t.Fatalf("tab: %v", err)
	}

	settings, _, err := menu.Page()
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if strings.Contains(settings.Body, "icon-diamond_ore") {
		t.Fatal("switching tabs still showed the ore list")
	}
	if !strings.Contains(settings.Body, `class="toggle on"`) {
		t.Fatal("the settings tab drew no toggle in its default state")
	}

	if err := menu.Send("toggle", []string{"guardar"}); err != nil {
		t.Fatalf("toggle: %v", err)
	}

	flipped, _, err := menu.Page()
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if settings.Body == flipped.Body {
		t.Fatal("flipping a setting changed nothing on the page")
	}
}

// the ore panel asks for more than the panel testdata does, so it gets its own
type digger struct {
	talker
}

func (digger) Held() plugin.Item {
	return "iron_pickaxe"
}

func (digger) Inventory() []plugin.Stack {
	return nil
}

func (digger) Count(plugin.Item) int {
	return 3
}

func (digger) Hold(plugin.Item) error {
	return nil
}

func (digger) Pursue(plugin.Goal) error {
	return nil
}

func (digger) Abandon() {
}
