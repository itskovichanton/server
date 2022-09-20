package security

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core"
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
