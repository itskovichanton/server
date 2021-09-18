package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"git.molbulak.ru/a.itskovich/molbulak-services-golang/pkg/server/filestorage"
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
	p := arg.(*core.CallParams)
	key, err := validation.CheckNotEmptyStr("key", p.GetParamStr("key"))
	if err != nil {
		return nil, err
	}
	lastModifiedTimeStampUTC, err := validation.CheckInt64("lastModifiedTimeStampUTC", p.GetParamStr("lastModifiedTimeStampUTC"))
	if err != nil {
		lastModifiedTimeStampUTC, err = validation.CheckInt64("if-modified-since", p.Context().Request().Header.Get("if-modified-since"))
	}
	var t time.Time
	if err == nil {
		t = time.Unix(lastModifiedTimeStampUTC, 0).UTC()
	}
	return c.FileStorageService.GetFile(key, &t)

}
