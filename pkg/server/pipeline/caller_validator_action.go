package pipeline

import "github.com/itskovichanton/server/pkg/server/entities"

type ValidateCallerAction struct {
	BaseActionImpl

	CallerValidatorService ICallerValidatorService
}

func (c *ValidateCallerAction) GetName() string {
	return "Сервис:ВалидацияВызова"
}

func (c *ValidateCallerAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*entities.CallParams)
	err := c.CallerValidatorService.Check(p, "")
	return arg, err
}
