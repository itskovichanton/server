package app

import (
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/server/pkg/server"
	"github.com/itskovichanton/server/pkg/server/pipeline"
)

type StubServerApp struct {
	app.IApp

	Config         *server.Config
	HttpController *pipeline.HttpControllerImpl
}

func (c *StubServerApp) Run() error {
	return c.HttpController.Start()
}
