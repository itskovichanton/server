package users

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/goava/pkg/goava/errs"
)

type IAuthService interface {
	Register(account *core.Account) (*core.Session, error)
	Login(authArgs *core.AuthArgs) (*core.Session, error)
	Logout(token string) *core.Session
	LogoutAll()
	RegisterAdmin() (*core.Session, error)
}

const (
	ReasonAlreadyExist                       = "REASON_ALREADY_EXIST"
	ReasonAuthorizationFailedInvalidPassword = "REASON_AUTHORIZATION_FAILED_INVALID_PASSWORD"
	ReasonAuthorizationFailedUserNotExist    = "REASON_AUTHORIZATION_FAILED_USER_NOT_EXIST"
)

// Implementation

type AuthServiceImpl struct {
	IAuthService

	UserRepo              IUserRepoService
	SessionStorageService ISessionStorageService
}

func (c *AuthServiceImpl) LogoutAll() {
	c.SessionStorageService.Clear()
}

func (c *AuthServiceImpl) Logout(token string) *core.Session {
	return c.SessionStorageService.LogoutByToken(token)
}

func (c *AuthServiceImpl) Login(a *core.AuthArgs) (*core.Session, error) {

	if len(a.SessionToken) > 0 {
		return c.SessionStorageService.GetSessionByToken(a.SessionToken), nil
	}

	_, err := validation.CheckNotEmptyStr("username", a.Username)
	if err != nil {
		return nil, err
	}
	_, err = validation.CheckNotEmptyStr("password", a.Password)
	if err != nil {
		return nil, err
	}
	user := c.UserRepo.FindByUsername(a.Username)
	if user == nil {
		return nil, errs.NewBaseErrorWithReason(fmt.Sprintf("Пользователь с именем %v не существует", a.Username), ReasonAuthorizationFailedUserNotExist)
	}

	if a.Password == user.Password {
		return c.SessionStorageService.AssignSession(user), nil
	}

	return nil, errs.NewBaseErrorWithReason(fmt.Sprintf("Неверный пароль", a.Username), ReasonAuthorizationFailedInvalidPassword)
}

func (c *AuthServiceImpl) Register(a *core.Account) (*core.Session, error) {
	_, err := validation.CheckNotEmptyStr("username", a.Username)
	if err != nil {
		return nil, err
	}
	_, err = validation.CheckNotEmptyStr("password", a.Password)
	if err != nil {
		return nil, err
	}

	if c.UserRepo.ContainsByUsername(a.Username) {
		return nil, errs.NewBaseErrorWithReason(fmt.Sprintf("Пользователь с именем %v уже существует", a.Username), ReasonAlreadyExist)
	}

	c.UserRepo.Put(a)

	return c.SessionStorageService.AssignSession(a), nil
}

func (c *AuthServiceImpl) RegisterAdmin() (*core.Session, error) {
	return c.Register(&core.Account{
		Username:     "admin",
		SessionToken: "admin-sessiontoken",
		Role:         core.RoleAdmin,
		Password:     "admin",
	})
}
