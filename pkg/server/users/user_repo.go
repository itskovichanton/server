package users

import (
	"bitbucket.org/itskovich/core/pkg/core"
)

type IUserRepoService interface {
	FindByUsername(username string) *core.Account
	Put(account *core.Account)
	ContainsByUsername(username string) bool
	Init()
}

type UserRepoServiceImpl struct {
	IUserRepoService

	storage map[string]*core.Account
}

func (c *UserRepoServiceImpl) Init() {
	c.storage = make(map[string]*core.Account)
}

func (c *UserRepoServiceImpl) FindByUsername(username string) *core.Account {
	a, ok := c.storage[username]
	if ok {
		return a
	}
	return nil
}

func (c *UserRepoServiceImpl) Put(account *core.Account) {
	c.storage[account.Username] = account
}

func (c *UserRepoServiceImpl) ContainsByUsername(username string) bool {
	_, ok := c.storage[username]
	return ok
}
