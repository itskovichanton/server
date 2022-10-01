package users

import (
	"github.com/itskovichanton/server/pkg/server/entities"
)

type IUserRepoService interface {
	FindByUsername(username string) *entities.Account
	Put(account *entities.Account)
	ContainsByUsername(username string) bool
	Init()
}

type UserRepoServiceImpl struct {
	IUserRepoService

	storage map[string]*entities.Account
}

func (c *UserRepoServiceImpl) Init() {
	c.storage = make(map[string]*entities.Account)
}

func (c *UserRepoServiceImpl) FindByUsername(username string) *entities.Account {
	a, ok := c.storage[username]
	if ok {
		return a
	}
	return nil
}

func (c *UserRepoServiceImpl) Put(account *entities.Account) {
	c.storage[account.Username] = account
}

func (c *UserRepoServiceImpl) ContainsByUsername(username string) bool {
	_, ok := c.storage[username]
	return ok
}
