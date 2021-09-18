package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core/frmclient"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bitbucket.org/itskovich/server/pkg/server/filestorage"
	"github.com/labstack/echo/v4"
	"net/http"
)

type IResponsePresenter interface {
	Write(context echo.Context, result *Result, httpStatus int) error
}

func (c *ResponsePresenterImpl) Write(context echo.Context, result *Result) error {
	return context.NoContent(c.GetHttpResponseCode(result, 0))
}

type ResponsePresenterImpl struct {
	IResponsePresenter

	ResponseModelProvider IResponseModelProvider
}

type IResponseModelProvider interface {
	ToModel(result *Result) interface{}
}

type ResponseModelProviderImpl struct {
	IResponseModelProvider
}

type JSONResponsePresenterImpl struct {
	ResponsePresenterImpl
}

func (c *JSONResponsePresenterImpl) Write(context echo.Context, result *Result, httpStatus int) error {
	//contentType := context.Request().Header.Get("Accept")
	if httpStatus == 0 {
		httpStatus = c.GetHttpResponseCode(result, httpStatus)
	}
	//if strings.EqualFold(contentType, "text/plain") {
	//	return context.String(httpStatus, cast.ToString(result.Res))
	//}
	return context.JSON(httpStatus, c.ResponseModelProvider.ToModel(result))
}

func (c *ResponseModelProviderImpl) ToModel(r *Result) interface{} {
	if r.Err != nil {
		switch ex := r.Err.Error.(type) {
		case *validation.ValidationError:
			return c.createValidationErrorModel(ex, r)
		}
	}
	return r
}

func (c *ResponseModelProviderImpl) createValidationErrorModel(e *validation.ValidationError, r *Result) interface{} {
	return &ValidationErrorModel{
		Result: *r,
		Error: &ValidationErrorErr{
			Err:          *r.Err,
			Param:        e.Param,
			InvalidValue: e.InvalidValue,
			Reason:       e.Reason,
		},
	}
}

type ValidationErrorModel struct {
	Result
	Error *ValidationErrorErr `json:"error,omitempty"`
}

type ValidationErrorErr struct {
	Err
	Param        string      `json:"param,omitempty"`
	InvalidValue interface{} `json:"invalidValue,omitempty"`
	Reason       string      `json:"reason,omitempty"`
}

type FileResponsePresenterImpl struct {
	JSONResponsePresenterImpl
}

func (c *FileResponsePresenterImpl) Write(context echo.Context, r *Result, httpStatus int) error {

	if r.Err != nil {
		switch ex := r.Err.Error.(type) {
		case *filestorage.FileStorageError:
			switch ex.Reason {
			case filestorage.ReasonNoUpdateNeeded:
				c.JSONResponsePresenterImpl.Write(context, r, http.StatusNotModified)
				return nil
			case filestorage.ReasonNotFound:
				c.JSONResponsePresenterImpl.Write(context, r, http.StatusNotFound)
				return nil
			}

			c.JSONResponsePresenterImpl.Write(context, r, http.StatusInternalServerError)
		}
	}

	f := r.Res.(*filestorage.FileInfo)

	if utils.FileExists(f.FullPath) {
		return context.File(f.FullPath)
	} else {
		return c.JSONResponsePresenterImpl.Write(context, r, http.StatusNotFound)
	}
}

func (c *ResponsePresenterImpl) GetHttpResponseCode(result *Result, httpStatus int) int {

	if httpStatus > 0 {
		return httpStatus
	}

	if result.Err == nil {
		return http.StatusOK
	}

	switch result.Err.Reason {
	case frmclient.ReasonTooManyRequests:
		return http.StatusTooManyRequests
	case frmclient.ReasonAccessDenied, ReasonProfileDenied, ReasonProfileDeniedByIP, frmclient.ReasonCallerUpdateRequired, frmclient.ReasonInactiveUser:
		return http.StatusForbidden
	case frmclient.ReasonAuthorizationRequired:
		return http.StatusUnauthorized
	case frmclient.ReasonServerRespondedWithError:
		return http.StatusOK
	case frmclient.ReasonValidation:
		return http.StatusBadRequest
	case frmclient.ReasonServerUnavailable:
		return http.StatusServiceUnavailable
	case frmclient.ReasonServerRespondedWithErrorNotFound:
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
