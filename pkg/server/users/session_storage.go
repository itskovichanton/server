package users

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"fmt"
	"math/rand"
)

type ISessionStorageService interface {
	IsLoggedIn(token string) bool
	GetSessionByUsername(username string) *core.Session
	GetSessionByToken(token string) *core.Session
	LogoutByToken(token string) *core.Session
	AssignSession(account *core.Account) *core.Session
	GetUsersCount() int
	Clear()
	LogoutByUsername(username string) *core.Session
}

type SessionStorageServiceImpl struct {
	ISessionStorageService

	tokenToSession  map[string]*core.Session
	usernameToToken map[string]string
}

func (c *SessionStorageServiceImpl) LogoutByUsername(username string) *core.Session {
	if len(username) > 0 {
		token, ok := c.usernameToToken[username]
		if ok {
			return c.LogoutByToken(token)
		}
	}
	return nil
}

func (c *SessionStorageServiceImpl) IsLoggedIn(token string) bool {
	return c.GetSessionByToken(token) != nil
}

func (c *SessionStorageServiceImpl) GetSessionByUsername(username string) *core.Session {
	token, ok := c.usernameToToken[username]
	if !ok {
		return nil
	}
	session, ok := c.tokenToSession[token]
	if !ok {
		return nil
	}
	return session
}

func (c *SessionStorageServiceImpl) GetSessionByToken(token string) *core.Session {
	if len(token) == 0 {
		return nil
	}
	r, ok := c.tokenToSession[token]
	if !ok {
		return nil
	}

	return r
}

func (c *SessionStorageServiceImpl) LogoutByToken(token string) *core.Session {

	removedSession := c.GetSessionByToken(token)
	if removedSession != nil {
		delete(c.tokenToSession, token)
		delete(c.usernameToToken, removedSession.Account.Username)
	}

	return removedSession
}

func (c *SessionStorageServiceImpl) GetUsersCount() int {
	return len(c.tokenToSession)
}

func (c *SessionStorageServiceImpl) Clear() {
	c.tokenToSession = make(map[string]*core.Session)
	c.usernameToToken = make(map[string]string)
}

func (c *SessionStorageServiceImpl) AssignSession(account *core.Account) *core.Session {

	c.LogoutByUsername(account.Username)
	c.LogoutByToken(account.SessionToken)

	token, ok := c.usernameToToken[account.Username]
	if !ok {
		token = c.calcNewToken(account)
		c.usernameToToken[account.Username] = token
	}
	session, ok := c.tokenToSession[token]
	if !ok {
		session = &core.Session{
			Token:   token,
			Account: account,
		}
		c.tokenToSession[token] = session
	}
	account.SessionToken = session.Token
	return session

}

func (c *SessionStorageServiceImpl) calcNewToken(account *core.Account) string {

	if len(account.SessionToken) > 0 {
		return account.SessionToken
	}

	r := c.getInitialSessionVariant(account)
	for i := 1; c.GetSessionByToken(r) != nil; i++ {
		r = utils.MD5(fmt.Sprintf("%v:%v", r, i))
	}

	return r
}

func (c *SessionStorageServiceImpl) getInitialSessionVariant(account *core.Account) string {
	return utils.MD5(fmt.Sprintf("%v:%v:%v", utils.CurrentTimeMillis(), account.Username, rand.Intn(10e6)))
}
