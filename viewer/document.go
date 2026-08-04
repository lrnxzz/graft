package viewer

// base is what every page gets before it writes a line: transparent so the world
// shows through, filling the window, and never selectable, since the pointer is
// a game cursor rather than a caret.
//
// The engine has no css filters at all — backdrop-filter, filter: blur, none of
// them move a pixel — so glass here is layered translucency and an inset
// highlight rather than a real blur.
const base = `html,body{margin:0;height:100%;background:transparent;color:#e9edf5;
font:14px/1.45 -apple-system,Segoe UI,Roboto,sans-serif;
-webkit-user-select:none;cursor:default;overflow:hidden}
*{box-sizing:border-box}
`

// Document assembles what the engine loads: the viewer's own rules, the sprites
// the body asks for, then the page's stylesheet, so a page can override anything
// above it.
func Document(style, body string) (string, error) {
	icons, err := IconStylesheet(body)
	if err != nil {
		return "", err
	}

	return `<!doctype html><meta charset="utf-8"><style>` +
		base + icons + style +
		`</style>` + body, nil
}
