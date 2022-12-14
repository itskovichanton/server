package pipeline

import (
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/echo-http"
	"github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/filestorage"
	"time"
)

type GetFileAction struct {
	BaseActionImpl

	FileStorageService filestorage.IFileStorageService
}

func (c *GetFileAction) GetName() string {
	return "GetFile"
}

func (c *GetFileAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*entities.CallParams)
	key, err := validation.CheckNotEmptyStr("key", p.GetParamStr("key"))
	if err != nil {
		return nil, err
	}
	lastModifiedTimeStampUTC, err := validation.CheckInt64("lastModifiedTimeStampUTC", p.GetParamStr("lastModifiedTimeStampUTC"))
	if err != nil {
		lastModifiedTimeStampUTC, err = validation.CheckInt64("if-modified-since", p.Request.(echo.Context).Request().Header.Get("if-modified-since"))
	}
	var t time.Time
	if err == nil {
		t = time.Unix(lastModifiedTimeStampUTC, 0).UTC()
	}
	return c.FileStorageService.GetFile(key, &t)

}
