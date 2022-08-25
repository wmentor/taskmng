package taskmng

import (
	"context"
)

type EventPublisher interface {
	PublishEvent(ctx context.Context, eventKey interface{}, msg string)
}
