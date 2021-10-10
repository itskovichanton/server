package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/logger"
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"fmt"
	"log"
	"strings"
)

type IAction interface {
	Run(arg interface{}) (interface{}, error)
	OnBeforeRun(arg interface{})
	OnError(arg interface{}, err *Err)
	OnSuccess(arg interface{}, result interface{})
	OnFinished(arg interface{}, result *Result)
	GetName() string
	Log(v ...interface{})
	GetLog() string
	PrepareErrorAlert(alertParams *core.AlertParams, err *Err, arg interface{})
	GetArgsForLogger(arg interface{}) interface{}
}

type BaseActionImpl struct {
	IAction

	logText strings.Builder
}

func (c *BaseActionImpl) GetArgsForLogger(arg interface{}) interface{} {
	return arg
}

func (c *BaseActionImpl) OnSuccess(arg interface{}, result interface{}) {}

func (c *BaseActionImpl) OnBeforeRun(arg interface{}) {}

func (c *BaseActionImpl) OnFinished(arg interface{}, result *Result) {}

func (c *BaseActionImpl) OnError(arg interface{}, err *Err) {}

func (c *BaseActionImpl) GetLog() string {
	return c.logText.String()
}

func (c *BaseActionImpl) Log(v ...interface{}) {
	c.logText.WriteString(fmt.Sprintf("%v", v))
}

func (c *BaseActionImpl) GetName() string {
	return utils.GetType(*c)
}

func (c *BaseActionImpl) PrepareErrorAlert(alertParams *core.AlertParams, err *Err, arg interface{}) {
}

type IActionRunner interface {
	Run(action IAction, argsProvider func() (interface{}, error), service IErrorProviderService) *Result
}

type ActionRunnerImpl struct {
	IActionRunner

	LoggerService               logger.ILoggerService
	ErrorHandler                core.IErrorHandler
	DefaultErrorProviderService IErrorProviderService
}

type Result struct {
	Res             interface{} `json:"result,omitempty"`
	Err             *Err        `json:"error,omitempty"`
	ExecutionTimeMs int64       `json:"executionTimeMs"`
}

type ActionContext struct {
	IActionContext

	Arg    interface{}
	Action IAction
	Ld     map[string]interface{}
}

type IActionContext interface {
}

type ActionContextImpl struct {
	ActionContext

	Logger *log.Logger
}

func (c *ActionRunnerImpl) Run(action IAction, argsProvider func() (interface{}, error), errorProviderService IErrorProviderService) *Result {

	ctx := ActionContextImpl{
		ActionContext: ActionContext{
			Action: action,
			Ld:     logger.NewLD(),
		},
		Logger: c.LoggerService.GetDefaultActionsLogger(),
	}

	result := &Result{}
	result.ExecutionTimeMs = utils.CurrentTimeMillis()

	arg, err := argsProvider()

	defer func() {
		result.ExecutionTimeMs = utils.CurrentTimeMillis() - result.ExecutionTimeMs
		action.OnFinished(arg, result)
		if len(action.GetLog()) > 0 {
			logger.Field(ctx.Ld, "log", action.GetLog())
		}
		logger.Result(ctx.Ld, result)
		if result.Err != nil {
			logger.Err(ctx.Ld, result.Err.Error)
		}
		logger.Print(ctx.Logger, ctx.Ld)
	}()

	logger.Action(ctx.Ld, action.GetName())
	//logger.Field(ctx.Ld, "c", caller)

	if err == nil {
		ctx.Arg = arg
		logger.Args(ctx.Ld, action.GetArgsForLogger(arg))
		action.OnBeforeRun(arg)
		result.Res, err = action.Run(arg)
		if err == nil {
			action.OnSuccess(arg, result.Res)
		}
	}

	if err != nil {
		result.Res = nil
		if errorProviderService == nil {
			errorProviderService = c.DefaultErrorProviderService
		}
		result.Err = errorProviderService.ProvideError(err)
		action.OnError(arg, result.Err)
		c.ErrorHandler.HandleWithCustomParams(result.Err.Error, func(alertParams *core.AlertParams) {
			action.PrepareErrorAlert(alertParams, result.Err, arg)
		})
	}

	return result

}

type ChainedActionImpl struct {
	BaseActionImpl

	Name       string
	Actions    []IAction
	lastAction IAction
}

func (c *ChainedActionImpl) PrepareErrorAlert(alertParams *core.AlertParams, err *Err, arg interface{}) {
	if c.lastAction != nil {
		c.lastAction.PrepareErrorAlert(alertParams, err, arg)
	}
}

func (c *ChainedActionImpl) GetName() string {
	if len(c.Name) > 0 {
		return c.Name
	}
	r := ""
	for _, a := range c.Actions {
		r += a.GetName() + "-"
	}
	c.Name = strings.TrimRight(r, "-")
	return c.Name
}

func (c *ChainedActionImpl) Run(arg interface{}) (interface{}, error) {

	var lastResult interface{}
	lastResult = nil

	for _, p := range c.Actions {
		c.lastAction = p
		if lastResult != nil {
			arg = lastResult
		}
		p.OnBeforeRun(arg)
		r, err := p.Run(arg)
		p.OnSuccess(arg, r)
		lastResult = r

		var errObj *Err
		errObj = nil

		if err != nil {
			errObj = &Err{
				Error:   err,
				Reason:  "",
				Message: err.Error(),
				Details: utils.GetErrorFullInfo(err),
			}
			be := errs.FindBaseError(err)
			if be != nil {
				errObj.Reason = be.Reason
			}
			p.OnError(arg, errObj)
		}
		p.OnFinished(arg, &Result{
			Res: lastResult,
			Err: errObj,
		})
		if err != nil {
			return lastResult, err
		}
	}

	return lastResult, nil
}
