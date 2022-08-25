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

func (manager *TaskManager) worker() {
	for task := range manager.inputTasks {
		manager.processTask(task)
	}
}

func (manager *TaskManager) processTask(task Task) {
	defer func() {
		if r := recover(); r != nil {
			manager.callbackStorage.PublishEvent(manager.ctx, errorCallbackKey, fmt.Sprintf("fatal error: %v", r))
		}
	}()

	if err := task.Exec(manager.ctx, manager.callbackStorage); err != nil {
		manager.callbackStorage.PublishEvent(manager.ctx, errorCallbackKey, err.Error())
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
