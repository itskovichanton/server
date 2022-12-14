package pipeline

import (
	"errors"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/server/pkg/server/filestorage"
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
	if len(r.Message) == 0 {
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
		r := ese.Reason
		if len(r) == 0 {
			return frmclient.ReasonServerRespondedWithError
		} else {
			return ese.Reason
		}
	}

	return frmclient.ReasonInternal
}

func (c *ErrorProviderServiceImpl) getErrMsg(e error) string {

	be := errs.FindBaseError(e)
	if be != nil {
		return be.Message
	}

	if c.Config.IsProfileProd() {
		return frmclient.InternalErrorMessage
	}

	return e.Error()
}
