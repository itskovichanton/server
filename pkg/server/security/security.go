package security

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"github.com/asaskevich/EventBus"
)

const TopicSecurityViolation = "TOPIC_SECURITY_VIOLATION"

type Security struct {
	EventBus                   EventBus.Bus
	OnSecurityViolationHandled func(err error)
	ErrorHandler               core.IErrorHandler
}

func (c *Security) Init() {
	c.OnSecurityViolationHandled = func(err error) {
		c.ErrorHandler.Handle(err, true)
	}
	c.EventBus.Subscribe(TopicSecurityViolation, c.OnSecurityViolationHandled)
}
