package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/frmclient"
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"bitbucket.org/itskovich/server/pkg/server/users"
	"strings"
)

type GetUserAction struct {
	BaseActionImpl

	AuthService users.IAuthService
}

func (c *GetUserAction) PrepareErrorAlert(alertParams *core.AlertParams, e *Err, arg interface{}) {

	if strings.EqualFold(e.Reason, frmclient.ReasonAuthorizationRequired) || strings.EqualFold(e.Reason, frmclient.ReasonInactiveUser) {
		alertParams.Send = false
		return
	}

}

func (c *GetUserAction) Run(arg interface{}) (interface{}, error) {

	p := arg.(*core.CallParams)

	if p.Caller.AuthArgs == nil && p.Caller.Session != nil && len(p.Caller.Session.Token) > 0 {
		p.Caller.AuthArgs = &core.AuthArgs{
			SessionToken: p.Caller.Session.Token,
		}
	}

	if p.Caller.AuthArgs == nil {
		return nil, errs.NewBaseErrorWithReason("Пользователь не авторизован", frmclient.ReasonAuthorizationRequired)
	}

	session, err := c.AuthService.Login(p.Caller.AuthArgs)
	if err != nil {
		return nil, err
	}

	p.Caller.Session = session
	if p.Caller.Session == nil {
		return nil, errs.NewBaseErrorWithReason("Пользователь не существует", users.ReasonAuthorizationFailedUserNotExist)
	}

	return p, nil
}

func (c *GetUserAction) GetName() string {
	return "GetUser"
}

type RegisterAccountAction struct {
	BaseActionImpl

	AuthService users.IAuthService
}

func (c *RegisterAccountAction) GetName() string {
	return "AccountRegistration"
}

func (c *RegisterAccountAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*core.CallParams)
	return c.AuthService.Register(ReadAccount(p))
}

type ValidateActiveUserAction struct {
	BaseActionImpl
}

func (c *ValidateActiveUserAction) GetName() string {
	return "ValidateActiveUser"
}

func (c *ValidateActiveUserAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*core.CallParams)
	if p.Caller.Session == nil || p.Caller.Session.Account == nil {
		return nil, errs.NewBaseErrorWithReason("Пользователь не авторизован", frmclient.ReasonAuthorizationRequired)
	}
	if p.Caller.Session.Account.CID == 0 {
		return nil, errs.NewBaseErrorWithReason("Пользователь не активный", frmclient.ReasonInactiveUser)
	}
	return p, nil
}

func (c *ValidateActiveUserAction) PrepareErrorAlert(alertParams *core.AlertParams, e *Err, arg interface{}) {

	if strings.EqualFold(e.Reason, frmclient.ReasonAuthorizationRequired) || strings.EqualFold(e.Reason, frmclient.ReasonInactiveUser) {
		alertParams.Send = false
		return
	}

}

type GetSessionAction struct {
	BaseActionImpl
}

func (c *GetSessionAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*core.CallParams)
	if p.Caller != nil && p.Caller.Session != nil {
		return p.Caller.Session, nil
	}
	return nil, nil
}
