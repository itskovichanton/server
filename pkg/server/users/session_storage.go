package users

import (
	"fmt"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/server/pkg/server/entities"
	"math/rand"
)

type ISessionStorageService interface {
	IsLoggedIn(token string) bool
	GetSessionByUsername(username string) *entities.Session
	GetSessionByToken(token string) *entities.Session
	LogoutByToken(token string) *entities.Session
	AssignSession(account *entities.Account) *entities.Session
	GetUsersCount() int
	Clear()
	LogoutByUsername(username string) *entities.Session
}

type SessionStorageServiceImpl struct {
	ISessionStorageService

	tokenToSession  map[string]*entities.Session
	usernameToToken map[string]string
}

func (c *SessionStorageServiceImpl) LogoutByUsername(username string) *entities.Session {
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

func (c *SessionStorageServiceImpl) GetSessionByUsername(username string) *entities.Session {
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

func (c *SessionStorageServiceImpl) GetSessionByToken(token string) *entities.Session {
	if len(token) == 0 {
		return nil
	}
	r, ok := c.tokenToSession[token]
	if !ok {
		return nil
	}

	return r
}

func (c *SessionStorageServiceImpl) LogoutByToken(token string) *entities.Session {

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
	c.tokenToSession = make(map[string]*entities.Session)
	c.usernameToToken = make(map[string]string)
}

func (c *SessionStorageServiceImpl) AssignSession(account *entities.Account) *entities.Session {

	c.LogoutByUsername(account.Username)
	c.LogoutByToken(account.SessionToken)

	token, ok := c.usernameToToken[account.Username]
	if !ok {
		token = c.calcNewToken(account)
		c.usernameToToken[account.Username] = token
	}
	session, ok := c.tokenToSession[token]
	if !ok {
		session = &entities.Session{
			Token:   token,
			Account: account,
		}
		c.tokenToSession[token] = session
	}
	account.SessionToken = session.Token
	return session

}

func (c *SessionStorageServiceImpl) calcNewToken(account *entities.Account) string {

	if len(account.SessionToken) > 0 {
		return account.SessionToken
	}

	r := c.getInitialSessionVariant(account)
	for i := 1; c.GetSessionByToken(r) != nil; i++ {
		r = utils.MD5(fmt.Sprintf("%v:%v", r, i))
	}

	return r
}

func (c *SessionStorageServiceImpl) getInitialSessionVariant(account *entities.Account) string {
	return utils.MD5(fmt.Sprintf("%v:%v:%v", utils.CurrentTimeMillis(), account.Username, rand.Intn(10e6)))
}
