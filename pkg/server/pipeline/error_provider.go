package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/frmclient"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"errors"
	"git.molbulak.ru/a.itskovich/molbulak-services-golang/pkg/server/filestorage"
	"strings"
)

type ErrorProviderServiceImpl struct {
	IErrorProviderService

	Config *core.Config
}

type IErrorProviderService interface {
	ProvideError(err error) *Err
}

type Err struct {
	Error   error  `json:"-"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (c *ErrorProviderServiceImpl) ProvideError(err error) *Err {
	r := &Err{
		Error:   err,
		Reason:  c.getErrReason(err),
		Message: c.getErrMsg(err),
		Details: c.getErrDetails(err),
	}
	if strings.EqualFold(r.Reason, errs.ReasonOther) && c.Config.IsProfileProd() {
		r.Message = frmclient.InternalErrorMessage
	}
	return r
}

func (c *ErrorProviderServiceImpl) getErrDetails(err error) string {
	if c.Config.IsProfileProd() {
		return ""
	}
	return utils.GetErrorFullInfo(err)
}

func (c *ErrorProviderServiceImpl) getErrReason(err error) string {

	switch err.(type) {
	case *validation.ValidationError:
		return frmclient.ReasonValidation
	case *CallerUpdateRequiredError:
		return frmclient.ReasonAccessDenied
	}

	var fse *filestorage.FileStorageError
	if errors.As(err, &fse) {
		switch fse.Reason {
		case filestorage.ReasonNotFound, filestorage.ReasonNoUpdateNeeded:
			return frmclient.ReasonServerRespondedWithErrorNotFound
		}
	}

	ese := errs.FindBaseError(err)
	if ese != nil {
		return ese.Reason
	}

	return frmclient.ReasonInternal
}

func (c *ErrorProviderServiceImpl) getErrMsg(e error) string {

	be := errs.FindBaseError(e)
	if be != nil {
		r := be.Message
		if !c.Config.IsProfileProd() && len(be.Details) > 0 {
			r += ", Details: " + be.Details
		}
		if len(r) == 0 {
			r = frmclient.InternalErrorMessage
		}
		return r
	}

	if c.Config.IsProfileProd() {
		return frmclient.InternalErrorMessage
	}

	return e.Error()
}
