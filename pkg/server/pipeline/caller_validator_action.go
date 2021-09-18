package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core"
)

type ValidateCallerAction struct {
	BaseActionImpl

	CallerValidatorService ICallerValidatorService
}

func (c *ValidateCallerAction) GetName() string {
	return "Сервис:ВалидацияВызова"
}

func (c *ValidateCallerAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*core.CallParams)
	err := c.CallerValidatorService.Check(p, "")
	return arg, err
}
