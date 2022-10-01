package main

import (
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/server/pkg/server/di"
	"go.uber.org/dig"
)

func main() {

	diC := &di.DI{}
	container := dig.New()
	diC.InitDI(container)

	var outerAppRunner app.IAppRunner
	err := diC.Container.Invoke(func(app app.IAppRunner) {
		outerAppRunner = app
	})

	if err != nil {
		panic(err)
	}

	runningErr := outerAppRunner.Run()
	if runningErr != nil {
		panic(runningErr)
	}
}
