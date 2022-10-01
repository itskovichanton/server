package pipeline

import (
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/server/pkg/server/entities"
)

type ICallerValidatorService interface {
	Check(a *entities.CallParams, name string) error
}

type CallerValidatorServiceImpl struct {
	ICallerValidatorService

	ServerSettingsProviderService IServerSettingsProviderService
}

const (
	InvalidCallerErrorReasonEmptyVersion = "EMPTY_VERSION"
)

func (c *CallerValidatorServiceImpl) Check(a *entities.CallParams, name string) error {
	err := c.validateVersionNotEmpty(a)
	if err != nil {
		return err
	}
	return c.validateVersionCode(a)
}

func (c *CallerValidatorServiceImpl) validateVersionNotEmpty(a *entities.CallParams) error {
	if a.Caller.Version == nil {
		return errs.NewBaseErrorWithReason("Версия клиента не указана", InvalidCallerErrorReasonEmptyVersion)
	}
	return nil
}

type CallerUpdateRequiredError struct {
	errs.BaseError

	RequiredVersion *entities.Version
	UpdateUrl       string
}

func (c *CallerValidatorServiceImpl) validateVersionCode(a *entities.CallParams) error {

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
