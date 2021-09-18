package pipeline

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"bitbucket.org/itskovich/goava/pkg/goava/httputils"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
)

func (c *EntityFromHTTPReaderServiceImpl) ReadCaller(r echo.Context) *core.Caller {

	res := core.Caller{
		IP:       r.RealIP(),
		Version:  ReadVersion(r.Request()),
		Type:     ReadCallerType(r.Request()),
		Language: c.ReadLanguage(r),
		AuthArgs: ReadAuthArgs(r.Request()),
	}

	sessionToken := r.Request().Header.Get("sessionToken")
	if len(sessionToken) == 0 && res.AuthArgs != nil && len(res.AuthArgs.SessionToken) > 0 {
		sessionToken = res.AuthArgs.SessionToken
	}
	if len(sessionToken) > 0 {
		res.Session = &core.Session{
			Token:   sessionToken,
			Account: nil,
		}
	}
	if res.AuthArgs != nil {
		res.Session = &core.Session{
			Token: res.AuthArgs.SessionToken,
		}
		res.Session.Account = &core.Account{
			Username:     res.AuthArgs.Username,
			Lang:         res.Language,
			FullName:     "",
			SessionToken: res.AuthArgs.SessionToken,
			Role:         core.RoleUser,
			Password:     res.AuthArgs.Password,
			IP:           res.IP,
		}
	}
	return &res
}

func ReadCallerType(r *http.Request) string {
	t := r.URL.Query().Get("caller-type")
	if len(t) == 0 {
		t = r.UserAgent()
	}
	return t
}

func ReadAuthArgs(r *http.Request) *core.AuthArgs {
	username, password, authOK := r.BasicAuth()
	if !authOK {
		return nil
	}

	return &core.AuthArgs{
		Username:     username,
		Password:     password,
		SessionToken: r.Header.Get("sessionToken"),
	}

}

type IEntityFromHTTPReaderService interface {
	ReadLanguage(r echo.Context) string
	ReadCallParams(r echo.Context) (*core.CallParams, error)
	GetParameters(r echo.Context) (map[string][]interface{}, error)
	ReadCaller(r echo.Context) *core.Caller
}

type EntityFromHTTPReaderServiceImpl struct {
	IEntityFromHTTPReaderService
	Config *core.Config
}

func (c *EntityFromHTTPReaderServiceImpl) ReadLanguage(r echo.Context) string {
	lang := r.QueryParam("lang")
	if len(lang) == 0 {
		lang = r.Request().Header.Get("lang")
	}
	if len(lang) == 0 {
		lang = c.Config.Actions.DefaultLang
	}
	return lang
}

func ReadVersion(r *http.Request) *core.Version {
	vc := r.Header.Get("Caller-Version-Code")
	if len(vc) == 0 {
		return nil
	}

	code, _ := strconv.Atoi(vc)
	return &core.Version{
		Code: code,
		Name: r.Header.Get("Caller-Version-Name"),
	}
}

func (c *EntityFromHTTPReaderServiceImpl) ReadCallParams(r echo.Context) (*core.CallParams, error) {
	params, err := c.GetParameters(r)
	if err != nil {
		return nil, err
	}
	return &core.CallParams{
		Request:    r,
		Parameters: params,
		URL:        httputils.GetUrl(r.Request()),
		Caller:     c.ReadCaller(r),
		Raw:        r.Request().URL.RawQuery,
	}, nil
}

func (c *EntityFromHTTPReaderServiceImpl) GetParameters(r echo.Context) (map[string][]interface{}, error) {
	var res map[string][]interface{}
	if strings.EqualFold("get", r.Request().Method) {
		res = utils.UpcastMapOfSlicesStr(r.Request().URL.Query())
	} else {

		// Parse multipart
		maxMem, err := c.Config.Server.Http.Multipart.GetMaxRequestSizeBytes()
		if err != nil {
			return nil, err
		}

		err = r.Request().ParseMultipartForm(int64(maxMem))
		if err == nil {
			res = utils.UpcastMapOfSlicesStr(r.Request().MultipartForm.Value)

			// Parse multipart files
			for field, fileHeaders := range r.Request().MultipartForm.File {
				v := []interface{}{}
				for _, fileHeader := range fileHeaders {
					//file, err := fileHeader.Open()
					//if err != nil {
					//	return nil, err
					//}

					//savedMultipartFormFilePath, err := c.Config.GetTempFile("multipart_" + fileHeader.Filename + "_*")
					//buf, err := ioutil.ReadAll(file)
					//if err != nil {
					//	return nil, err
					//}

					//err = ioutil.WriteFile(savedMultipartFormFilePath.Name(), buf, os.ModePerm)
					//if err != nil {
					//	return nil, err
					//}
					v = append(v, fileHeader)
				}
				res[field] = v
			}
		}
	}

	if res == nil {
		res = map[string][]interface{}{}
	}
	for k, paramName := range r.ParamNames() {
		res["path__"+paramName] = []interface{}{r.ParamValues()[k]}
	}
	if r.Request().Form != nil {
		for k, v := range r.Request().Form {
			if len(v) > 0 {
				res[k] = []interface{}{v[0]}
			}
		}
	}

	return res, nil
}

func (c *EntityFromHTTPReaderServiceImpl) ReadSession(r *http.Request) *core.Session {
	return &core.Session{
		//Token:         "",
		Account: nil,
	}
}

func ReadAccount(p *core.CallParams) *core.Account {
	a := core.Account{
		Password:     p.GetParamStr("password"),
		Username:     p.GetParamStr("username"),
		Role:         p.GetParamStr("role"),
		Lang:         p.GetParamStr("lang"),
		FullName:     p.GetParamStr("fullName"),
		SessionToken: p.GetParamStr("token"),
		IP:           p.Caller.IP,
	}
	cid, err := validation.CheckInt64("c_id", p.GetParamStr("c_id"))
	if err != nil {
		a.CID = cid
	}
	mclId, err := validation.CheckInt64("mcl_id", p.GetParamStr("mcl_id"))
	if err != nil {
		a.MCLID = mclId
	}
	if p.Caller.AuthArgs != nil {
		if len(a.Password) == 0 {
			a.Password = p.Caller.AuthArgs.Password
		}
		if len(a.Username) == 0 {
			a.Username = p.Caller.AuthArgs.Username
		}
		if len(a.SessionToken) == 0 {
			a.SessionToken = p.Caller.AuthArgs.SessionToken
		}
	}

	if p.Caller.Session.Account != nil {
		if len(a.Role) == 0 {
			a.Role = p.Caller.Session.Account.Role
		}
		if len(a.Lang) == 0 {
			a.Lang = p.Caller.Session.Account.Lang
		}
		if a.MCLID == 0 {
			a.MCLID = p.Caller.Session.Account.MCLID
		}
		if a.CID == 0 {
			a.CID = p.Caller.Session.Account.CID
		}
	}
	return &a
}
