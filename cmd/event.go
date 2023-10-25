package main

import (
	"flag"
	"github.com/ryanjarv/msh/cmd/common"
	"github.com/ryanjarv/msh/cmd/event"
	"github.com/ryanjarv/msh/pkg/app"
	L "github.com/ryanjarv/msh/pkg/logger"
	"log"
	"os"
)

func main() {
	flag.Parse()

	app, err := app.GetApp(common.Registry, os.Stdin, os.Stdout)
	if err != nil {
		L.Error.Fatalln("%s: get app: %w", os.Args[0], err)
	}

	l, err := event.NewEvent()
	if err != nil {
		L.Error.Fatalln("%s: new", l.GetName(), err)
	}

	err = app.Run(l)
	if err != nil {
		log.Fatalf("%s: run: %s", l.GetName(), err)
	}
}
