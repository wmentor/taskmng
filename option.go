package taskmng

import (
	"context"

	"github.com/pkg/errors"
)

type Option func(*TaskManager) error

func WithContext(ctx context.Context) Option {
	return func(manager *TaskManager) error {
		if manager.ctx == nil || manager.ctx == context.Context(nil) {
			manager.ctx, manager.cancel = context.WithCancel(ctx)
			return nil
		}
		return errors.WithMessage(ErrInvalidParam, "context already set")
	}
}

func WithTaskQueueSize(size int) Option {
	return func(manager *TaskManager) error {
		if size < 1 {
			return errors.WithMessage(ErrInvalidParam, "invalid task queue size")
		}
		manager.taskQueueSize = size
		return nil
	}
}

func WithWorkerPoolSize(size int) Option {
	return func(manager *TaskManager) error {
		if size < 1 {
			return errors.WithMessage(ErrInvalidParam, "invalid worker pool size")
		}
		manager.workerPoolSize = size
		return nil
	}
}

func WithErrorCallback(callback CallbackFunc) Option {
	return func(manager *TaskManager) error {
		manager.callbackStorage.RegCallback(errorCallbackKey, callback)
		return nil
	}
}

func WithEventCallback(eventKey interface{}, callback CallbackFunc) Option {
	return func(manager *TaskManager) error {
		manager.callbackStorage.RegCallback(eventKey, callback)
		return nil
	}
}
