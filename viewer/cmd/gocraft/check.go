package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lrnxzz/go-craft/viewer/ultralight"
)

// check walks the html engine's bring-up one step at a time, with no server and
// no window, printing before each step rather than after. Two of these steps can
// kill the process outright instead of returning an error, so the last line
// printed is the one that failed.
func check() error {
	step("engine version")
	fmt.Println("   ", ultralight.Version())

	step("resources next to the binary")
	for _, needed := range [...]string{"icudt67l.dat", "cacert.pem"} {
		path := filepath.Join("resources", needed)

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("%s is missing, and the engine aborts without it: %w", path, err)
		}

		fmt.Printf("    %s (%d bytes)\n", path, info.Size())
	}

	step("guard against the engine naming its own threads")
	fmt.Println("    installed:", ultralight.Guard())

	step("raising that exception on purpose")
	ultralight.Probe()

	seen, caught := ultralight.Caught()
	fmt.Printf("    survived, the guard saw %d and caught %d\n", seen, caught)

	step("logging to ultralight.log")
	ultralight.Log("ultralight.log")

	step("starting the renderer")
	engine, err := ultralight.Open("resources")
	if err != nil {
		return err
	}

	step("opening a view")
	view := engine.View(320, 180)
	defer view.Close()

	step("loading a page")
	view.LoadHTML(`<body style="margin:0;background:linear-gradient(#123,#345)">
		<p style="color:#fff;font:20px sans-serif;padding:20px">gocraft</p>`)

	for range 400 {
		engine.Update()
		if !view.Loading() {
			break
		}
	}
	if view.Loading() {
		return errors.New("the page never finished loading")
	}

	seen, caught = ultralight.Caught()
	fmt.Printf("    the guard saw %d and caught %d while starting\n", seen, caught)

	step("painting")
	engine.Render()

	var alpha byte
	view.Pixels(func(pixels []byte, stride int) {
		alpha = pixels[90*stride+160*4+3]
	})
	if alpha == 0 {
		return errors.New("the page painted nothing")
	}

	step("writing check.png")
	if !view.WritePNG("check.png") {
		return errors.New("the frame could not be written")
	}

	fmt.Println("\nthe html engine works. if the game still crashes, it is the opengl side.")

	return nil
}

func step(what string) {
	fmt.Println("==>", what)

	_ = os.Stdout.Sync()
}
