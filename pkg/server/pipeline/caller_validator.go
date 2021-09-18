package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/frmclient"
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
)

type ICallerValidatorService interface {
	Check(a *core.CallParams, name string) error
}

type CallerValidatorServiceImpl struct {
	ICallerValidatorService

	ServerSettingsProviderService IServerSettingsProviderService
}

const (
	InvalidCallerErrorReasonEmptyVersion = "EMPTY_VERSION"
)

func (c *CallerValidatorServiceImpl) Check(a *core.CallParams, name string) error {
	err := c.validateVersionNotEmpty(a)
	if err != nil {
		return err
	}
	return c.validateVersionCode(a)
}

func (c *CallerValidatorServiceImpl) validateVersionNotEmpty(a *core.CallParams) error {
	if a.Caller.Version == nil {
		return errs.NewBaseErrorWithReason("Версия клиента не указана", InvalidCallerErrorReasonEmptyVersion)
	}
	return nil
}

type CallerUpdateRequiredError struct {
	errs.BaseError

	RequiredVersion *core.Version
	UpdateUrl       string
}

func (c *CallerValidatorServiceImpl) validateVersionCode(a *core.CallParams) error {

	settings := c.ServerSettingsProviderService.GetSettings()

	if settings.Version == nil || settings.Version.Code <= a.Caller.Version.Code {
		return nil
	}

	return &CallerUpdateRequiredError{
		BaseError:       *errs.NewBaseErrorWithReason("Неверная версия клиента, необходимо обновить клиент", frmclient.ReasonCallerUpdateRequired),
		RequiredVersion: settings.Version,
		UpdateUrl:       settings.UpdateAppUrl,
	}

}