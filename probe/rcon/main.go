package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lrnxzz/go-craft/rcon"
)

const (
	address  = "localhost:25575"
	password = "gocraft"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	console, err := rcon.Dial(address, password)
	if err != nil {
		return err
	}
	defer func() {
		_ = console.Close()
	}()

	for _, command := range os.Args[1:] {
		answer, err := console.Run(command)
		if err != nil {
			return err
		}

		fmt.Println(answer)
	}

	return nil
}
