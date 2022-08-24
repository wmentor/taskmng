package taskmng

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrInvalidParam = errors.New("invalid param")
	ErrInternal     = errors.New("internal error")

	errorCallbackKey struct{}
)

const (
	defaultTaskQueueSize  = 128
	defaultWorkenPoolSize = 8
)

type CallbackFunc func(string)

type EventPublisher interface {
	PublishEvent(eventKey interface{}, msg string)
}

type Option func(*TaskManager) error

type Task interface {
	Exec(context.Context, EventPublisher) error
}

type TaskManager struct {
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	once            sync.Once
	taskQueueSize   int
	workerPoolSize  int
	inputTasks      chan Task
	callbackStorage callbackStorage
}

func New(opts ...Option) (*TaskManager, error) {
	manager := &TaskManager{
		ctx:             nil,
		taskQueueSize:   defaultTaskQueueSize,
		workerPoolSize:  defaultWorkenPoolSize,
		callbackStorage: make(callbackStorage),
	}

	for _, fn := range opts {
		if err := fn(manager); err != nil {
			return nil, err
		}
	}

	if manager.ctx == nil || manager.ctx == context.Context(nil) {
		manager.ctx, manager.cancel = context.WithCancel(context.Background())
	}

	manager.inputTasks = make(chan Task, manager.taskQueueSize)

	for i := 0; i < manager.workerPoolSize; i++ {
		manager.wg.Add(1)
		go func() {
			defer manager.wg.Done()
			manager.worker()
		}()
	}

	return manager, nil
}

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

func WithErrorCallback(callback func(string)) Option {
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

func (manager *TaskManager) worker() {
	for task := range manager.inputTasks {
		manager.processTask(task)
	}
}

func (manager *TaskManager) processTask(task Task) {
	defer func() {
		if r := recover(); r != nil {
			manager.callbackStorage.PublishEvent(errorCallbackKey, fmt.Sprintf("fatal error: %v", r))
		}
	}()

	if err := task.Exec(manager.ctx, manager.callbackStorage); err != nil {
		manager.callbackStorage.PublishEvent(errorCallbackKey, err.Error())
	}
}

func (manager *TaskManager) AddTask(task Task) {
	if task != nil && task != Task(nil) {
		manager.inputTasks <- task
	}
}

func (manager *TaskManager) Stop(force bool) {
	manager.once.Do(func() {
		if force {
			manager.cancel()
		} else {
			defer manager.cancel()
		}
		close(manager.inputTasks)
		manager.wg.Wait()
	})
}
