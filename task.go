package taskmng

import (
	"context"
)

//go:generate mockgen --source=./task.go --destination=./testdata/task.go --package=testdata
type Task interface {
	Exec(context.Context, EventPublisher) error
}
