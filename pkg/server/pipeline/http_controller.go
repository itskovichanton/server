package pipeline

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/httputils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"strings"
)

type IHttpController interface {
	Start() error
	AddRouterModifier(modifier func(router *echo.Echo))
}

type HttpControllerImpl struct {
	IHttpController

	CheckSecurityAction   *CheckSecurityAction
	GetUserAction         *GetUserAction
	NopAction             *NopActionImpl
	ValidateCallerAction  *ValidateCallerAction
	RegisterAccountAction *RegisterAccountAction
	GetFileAction         *GetFileAction
	GetSessionAction      *GetSessionAction

	Config                      *core.Config
	ActionRunner                IActionRunner
	EntityFromHTTPReaderService IEntityFromHTTPReaderService
	DefaultResponsePresenter    IResponsePresenter
	FileResponsePresenter       IResponsePresenter
	routerModifiers             []func(e *HttpControllerImpl)
	EchoEngine                  *echo.Echo
}

func (c *HttpControllerImpl) AddRouterModifier(modifier func(e *HttpControllerImpl)) {
	c.routerModifiers = append(c.routerModifiers, modifier)
}

func (c *HttpControllerImpl) Start() error {

	if c.Config.Server == nil {
		return nil
	}

	c.init()

	c.EchoEngine.Use(middleware.Logger())
	c.EchoEngine.Use(middleware.Recover())
	if c.Config.Server.EnableCORS {
		c.EchoEngine.Use(middleware.CORS())
	}
	c.EchoEngine.HideBanner = true
	c.EchoEngine.Debug = true

	for _, modifier := range c.routerModifiers {
		modifier(c)
	}

	ssl := c.Config.Server.Http.Ssl
	protocol := "https"
	if ssl == nil || !ssl.Enabled {
		protocol = "http"
	}

	var err error
	if ssl != nil && strings.EqualFold(protocol, "https") {
		err = c.EchoEngine.StartTLS(fmt.Sprintf(":%v", c.Config.Server.Port), ssl.CertFile, ssl.KeyFile)
	} else {
		err = c.EchoEngine.Start(fmt.Sprintf(":%v", c.Config.Server.Port))
	}
	c.EchoEngine.Logger.Fatal(err)

	return err
}

func (c *HttpControllerImpl) GetDefaultHandler(action IAction) func(context echo.Context) error {
	return c.GetDefaultHandlerByFunc(func() IAction {
		return action
	})
}

func (c *HttpControllerImpl) GetDefaultHandlerByFunc(action func() IAction) func(context echo.Context) error {
	return c.GetHandlerByFuncPresenterErrorProvider(action, nil, nil)
}

func (c *HttpControllerImpl) GetHandlerByActionPresenter(action IAction, presenter IResponsePresenter) func(context echo.Context) error {
	return c.GetHandlerByFuncPresenterErrorProvider(func() IAction {
		return action
	}, presenter, nil)
}

func (c *HttpControllerImpl) GetHandlerByPresenterErrorProvider(action IAction, presenter IResponsePresenter, errorProviderService IErrorProviderService) func(context echo.Context) error {
	return c.GetHandlerByFuncPresenterErrorProvider(func() IAction {
		return action
	}, presenter, errorProviderService)
}

func (c *HttpControllerImpl) GetHandlerByFuncPresenterErrorProvider(action func() IAction, presenter IResponsePresenter, errorProviderService IErrorProviderService) func(context echo.Context) error {
	return func(context echo.Context) error {

		result := c.ActionRunner.Run(
			action(),
			func() (interface{}, error) {
				return c.EntityFromHTTPReaderService.ReadCallParams(context)
			},
			errorProviderService,
		)

		if presenter == nil {
			presenter = c.DefaultResponsePresenter
		}
		return presenter.Write(context, result, 0)
	}
}

func (c *HttpControllerImpl) init() {
	c.AddRouterModifier(func(e *HttpControllerImpl) {
		//c.GETPOST("/error", c.GetDefaultHandler(&ChainedActionImpl{Actions: []IAction{c.ValidateCallerAction, &ImmediateFailedAction{}}}))
		//r.GET("/setServerStateAction", c.GetDefaultHandler(c.SetServerStateAction))
		c.GETPOST("/api/admin/registerAccount", c.GetDefaultHandler(&ChainedActionImpl{
			Actions: []IAction{c.ValidateCallerAction, c.GetUserAction /*c.CheckSecurityAction.WithActionName("AccountRegistration"),*/, c.RegisterAccountAction},
		}))
		c.GETPOST("/api/admin/getAccount", c.GetDefaultHandler(&ChainedActionImpl{
			Actions: []IAction{c.ValidateCallerAction, c.GetUserAction},
		}))
		//r.GET("/api/getFile", c.GetHandlerByActionPresenter(&ChainedActionImpl{
		//	Actions: []IAction{c.CheckServerStateAction /*c.ValidateCallerAction,*/, c.GetUserAction /*checkWithSecurityAction(SecurityService.Params.Builder().admin().build()),*/, c.GetFileAction},
		//}, c.FileResponsePresenter))
	})
}

func (c *HttpControllerImpl) GETPOST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	//if c.Config.Server.EnableThrottleMode {
	//	m = append(m, c.getThrottleMiddlewareFunc())
	//}
	httputils.GETPOST(c.EchoEngine, path, h, m...)
}
