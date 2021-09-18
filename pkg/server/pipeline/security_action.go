package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core"
)

type CheckSecurityAction struct {
	BaseActionImpl

	SecurityService   ISecurityService
	CheckedActionName string
}

func (c *CheckSecurityAction) GetName() string {
	return "ПроверкаБезопасности"
}

func (c *CheckSecurityAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*core.CallParams)
	err := c.SecurityService.Check(p, c.CheckedActionName)
	return arg, err
}
