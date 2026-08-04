package viewer

// transparent so the world shows through, and unselectable because the pointer
// is a game cursor rather than a caret
const base = `html,body{margin:0;height:100%;background:transparent;color:#e9edf5;
font:14px/1.45 -apple-system,Segoe UI,Roboto,sans-serif;
-webkit-user-select:none;cursor:default;overflow:hidden}
*{box-sizing:border-box}
`

// the page's stylesheet goes last so it can override everything above it
func Document(style, body string) (string, error) {
	icons, err := IconStylesheet(body)
	if err != nil {
		return "", err
	}

	return `<!doctype html><meta charset="utf-8"><style>` +
		base + icons + style +
		`</style>` + body, nil
}
